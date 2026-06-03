package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
)

type StatePayload struct {
	Type   string `json:"type"`
	UserID int    `json:"user_id,omitempty"`
}

type StateStore interface {
	Set(ctx context.Context, state string, payload StatePayload, ttl time.Duration) error
	Pop(ctx context.Context, state string) (StatePayload, error)
}

type RedisStateStore struct {
	redis  adminservice.RedisClient
	prefix string
}

func NewRedisStateStore(redisClient adminservice.RedisClient, prefix string) *RedisStateStore {
	return &RedisStateStore{redis: redisClient, prefix: strings.TrimRight(prefix, ":")}
}

func (s *RedisStateStore) Set(ctx context.Context, state string, payload StatePayload, ttl time.Duration) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.redis.Set(ctx, s.key(state), string(raw), ttl).Err()
}

func (s *RedisStateStore) Pop(ctx context.Context, state string) (StatePayload, error) {
	raw, err := s.redis.Get(ctx, s.key(state)).Result()
	if errors.Is(err, redis.Nil) {
		return StatePayload{}, ErrStateNotFound
	}
	if err != nil {
		return StatePayload{}, err
	}
	if err := s.redis.Del(ctx, s.key(state)).Err(); err != nil {
		return StatePayload{}, err
	}
	var payload StatePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return StatePayload{}, err
	}
	return payload, nil
}

func (s *RedisStateStore) key(state string) string {
	return s.prefix + ":" + state
}

type MemoryStateStore struct {
	mu      sync.Mutex
	entries map[string]memoryStateEntry
}

type memoryStateEntry struct {
	payload StatePayload
	expires time.Time
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{entries: make(map[string]memoryStateEntry)}
}

func (s *MemoryStateStore) Set(_ context.Context, state string, payload StatePayload, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[state] = memoryStateEntry{payload: payload, expires: time.Now().Add(ttl)}
	return nil
}

func (s *MemoryStateStore) Pop(_ context.Context, state string) (StatePayload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[state]
	if !ok {
		return StatePayload{}, ErrStateNotFound
	}
	delete(s.entries, state)
	if !entry.expires.IsZero() && time.Now().After(entry.expires) {
		return StatePayload{}, ErrStateNotFound
	}
	return entry.payload, nil
}

func (s *MemoryStateStore) Peek(state string) (StatePayload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[state]
	return entry.payload, ok
}
