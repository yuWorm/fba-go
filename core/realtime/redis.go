package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/redisx"
)

type RedisOnlineClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SAdd(ctx context.Context, key string, members ...any) *redis.IntCmd
	SRem(ctx context.Context, key string, members ...any) *redis.IntCmd
	SCard(ctx context.Context, key string) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

type RedisOnlineStore struct {
	client RedisOnlineClient
	keys   redisx.Keys
}

func NewRedisOnlineStore(client RedisOnlineClient, keys redisx.Keys) *RedisOnlineStore {
	return &RedisOnlineStore{client: client, keys: keys}
}

func (s *RedisOnlineStore) Connect(sid string, sessionUUID string) {
	if s == nil || s.client == nil || sid == "" || sessionUUID == "" {
		return
	}
	ctx := context.Background()
	// Mirror Python's Socket.IO online Redis shape so admin monitor can read
	// session online status from fba:token_online independently of login rows.
	_ = s.client.Set(ctx, s.keys.OnlineSID(sid), sessionUUID, 0).Err()
	_ = s.client.SAdd(ctx, s.keys.OnlineSession(sessionUUID), sid).Err()
	_ = s.client.SAdd(ctx, s.keys.OnlineSet(), sessionUUID).Err()
}

func (s *RedisOnlineStore) Disconnect(sid string) {
	if s == nil || s.client == nil || sid == "" {
		return
	}
	ctx := context.Background()
	sessionUUID, err := s.client.Get(ctx, s.keys.OnlineSID(sid)).Result()
	if errors.Is(err, redis.Nil) || sessionUUID == "" {
		return
	}
	if err != nil {
		return
	}
	sessionKey := s.keys.OnlineSession(sessionUUID)
	_ = s.client.Del(ctx, s.keys.OnlineSID(sid)).Err()
	_ = s.client.SRem(ctx, sessionKey, sid).Err()
	// Only remove the online session marker after the last Socket.IO SID for
	// that browser session disconnects; multiple tabs can share one session_uuid.
	if count, err := s.client.SCard(ctx, sessionKey).Result(); err == nil && count == 0 {
		_ = s.client.Del(ctx, sessionKey).Err()
		_ = s.client.SRem(ctx, s.keys.OnlineSet(), sessionUUID).Err()
	}
}

func (s *RedisOnlineStore) SessionForSID(sid string) string {
	if s == nil || s.client == nil || sid == "" {
		return ""
	}
	value, err := s.client.Get(context.Background(), s.keys.OnlineSID(sid)).Result()
	if err != nil {
		return ""
	}
	return value
}

func (s *RedisOnlineStore) Sessions() []string {
	if s == nil || s.client == nil {
		return []string{}
	}
	items, err := s.client.SMembers(context.Background(), s.keys.OnlineSet()).Result()
	if err != nil {
		return []string{}
	}
	sort.Strings(items)
	return items
}

func (s *RedisOnlineStore) SIDs(sessionUUID string) []string {
	if s == nil || s.client == nil || sessionUUID == "" {
		return []string{}
	}
	items, err := s.client.SMembers(context.Background(), s.keys.OnlineSession(sessionUUID)).Result()
	if err != nil {
		return []string{}
	}
	sort.Strings(items)
	return items
}

type BroadcastMessage struct {
	Origin string          `json:"origin"`
	Event  string          `json:"event"`
	Data   json.RawMessage `json:"data"`
}

type Broadcaster interface {
	Publish(ctx context.Context, message BroadcastMessage) error
	Start(ctx context.Context, handler func(BroadcastMessage)) error
	Close() error
}

type RedisBroadcastClient interface {
	Publish(ctx context.Context, channel string, message any) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
}

type RedisBroadcaster struct {
	client  RedisBroadcastClient
	channel string
	mu      sync.Mutex
	pubsub  *redis.PubSub
}

func NewRedisBroadcaster(client RedisBroadcastClient, channel string) *RedisBroadcaster {
	return &RedisBroadcaster{client: client, channel: channel}
}

func (b *RedisBroadcaster) Publish(ctx context.Context, message BroadcastMessage) error {
	if b == nil || b.client == nil || b.channel == "" {
		return nil
	}
	raw, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, b.channel, string(raw)).Err()
}

func (b *RedisBroadcaster) Start(ctx context.Context, handler func(BroadcastMessage)) error {
	if b == nil || b.client == nil || b.channel == "" || handler == nil {
		return nil
	}
	b.mu.Lock()
	if b.pubsub != nil {
		b.mu.Unlock()
		return nil
	}
	pubsub := b.client.Subscribe(ctx, b.channel)
	b.pubsub = pubsub
	b.mu.Unlock()

	go func() {
		// Pub/Sub messages are best-effort realtime fan-out. Invalid JSON is
		// ignored so one bad producer cannot stop local Socket.IO delivery.
		for {
			select {
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				var message BroadcastMessage
				if err := json.Unmarshal([]byte(msg.Payload), &message); err == nil {
					handler(message)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (b *RedisBroadcaster) Close() error {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	pubsub := b.pubsub
	b.pubsub = nil
	b.mu.Unlock()
	if pubsub == nil {
		return nil
	}
	return pubsub.Close()
}
