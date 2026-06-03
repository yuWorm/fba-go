package realtime_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/realtime"
	"github.com/yuWorm/fba-go/core/redisx"
)

func TestRedisOnlineStoreTracksPythonCompatibleOnlineKeys(t *testing.T) {
	client := newFakeRealtimeRedis()
	store := realtime.NewRedisOnlineStore(client, redisx.NewKeys("fba"))

	store.Connect("sid-1", "session-1")
	store.Connect("sid-2", "session-1")

	if sessions := store.Sessions(); len(sessions) != 1 || sessions[0] != "session-1" {
		t.Fatalf("Sessions() = %v, want [session-1]", sessions)
	}
	if got := store.SessionForSID("sid-1"); got != "session-1" {
		t.Fatalf("SessionForSID() = %q, want session-1", got)
	}
	if sids := store.SIDs("session-1"); len(sids) != 2 || sids[0] != "sid-1" || sids[1] != "sid-2" {
		t.Fatalf("SIDs() = %v, want [sid-1 sid-2]", sids)
	}

	store.Disconnect("sid-1")
	if sessions := store.Sessions(); len(sessions) != 1 || sessions[0] != "session-1" {
		t.Fatalf("Sessions() after first disconnect = %v, want [session-1]", sessions)
	}

	store.Disconnect("sid-2")
	if sessions := store.Sessions(); len(sessions) != 0 {
		t.Fatalf("Sessions() after final disconnect = %v, want empty", sessions)
	}
	if _, ok := client.values["fba:token_online:session:session-1"]; ok {
		t.Fatalf("session set key was not removed")
	}
}

func TestRedisBroadcasterPublishesJSONEnvelope(t *testing.T) {
	client := newFakeRealtimeRedis()
	broadcaster := realtime.NewRedisBroadcaster(client, "fba:realtime:broadcast")
	message := realtime.BroadcastMessage{
		Origin: "node-a",
		Event:  realtime.EventTaskNotification,
		Data:   json.RawMessage(`{"msg":"ok"}`),
	}

	if err := broadcaster.Publish(context.Background(), message); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if client.publishedChannel != "fba:realtime:broadcast" {
		t.Fatalf("published channel = %q", client.publishedChannel)
	}
	var got realtime.BroadcastMessage
	if err := json.Unmarshal([]byte(client.publishedMessage), &got); err != nil {
		t.Fatalf("published payload is not JSON: %v", err)
	}
	if got.Origin != "node-a" || got.Event != realtime.EventTaskNotification || string(got.Data) != `{"msg":"ok"}` {
		t.Fatalf("published message = %+v", got)
	}
}

func TestSocketIOHubPublishesBroadcastWhenBroadcasterConfigured(t *testing.T) {
	broadcaster := &captureBroadcaster{}
	hub := realtime.NewSocketIOHub(
		realtime.NewMemoryOnlineStore(),
		realtime.WithNodeID("node-a"),
		realtime.WithBroadcaster(broadcaster),
	)

	if err := hub.Emit(realtime.EventTaskNotification, realtime.TaskNotification{Msg: "ok"}); err != nil {
		t.Fatalf("Emit() error = %v", err)
	}

	if len(broadcaster.messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(broadcaster.messages))
	}
	if broadcaster.messages[0].Origin != "node-a" || broadcaster.messages[0].Event != realtime.EventTaskNotification {
		t.Fatalf("broadcast envelope = %+v", broadcaster.messages[0])
	}
	if string(broadcaster.messages[0].Data) != `{"msg":"ok"}` {
		t.Fatalf("broadcast data = %s", broadcaster.messages[0].Data)
	}
	if hub.ReceiveBroadcast(broadcaster.messages[0]) {
		t.Fatal("ReceiveBroadcast() accepted self-origin message")
	}
}

type captureBroadcaster struct {
	messages []realtime.BroadcastMessage
	handler  func(realtime.BroadcastMessage)
	closed   bool
}

func (b *captureBroadcaster) Publish(_ context.Context, message realtime.BroadcastMessage) error {
	b.messages = append(b.messages, message)
	return nil
}

func (b *captureBroadcaster) Start(_ context.Context, handler func(realtime.BroadcastMessage)) error {
	b.handler = handler
	return nil
}

func (b *captureBroadcaster) Close() error {
	b.closed = true
	return nil
}

type fakeRealtimeRedis struct {
	values           map[string]string
	sets             map[string]map[string]struct{}
	publishedChannel string
	publishedMessage string
}

func newFakeRealtimeRedis() *fakeRealtimeRedis {
	return &fakeRealtimeRedis{
		values: make(map[string]string),
		sets:   make(map[string]map[string]struct{}),
	}
}

func (r *fakeRealtimeRedis) Set(_ context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
	r.values[key] = StringValue(value)
	return redis.NewStatusResult("OK", nil)
}

func (r *fakeRealtimeRedis) Get(_ context.Context, key string) *redis.StringCmd {
	value, ok := r.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (r *fakeRealtimeRedis) Del(_ context.Context, keys ...string) *redis.IntCmd {
	var count int64
	for _, key := range keys {
		if _, ok := r.values[key]; ok {
			delete(r.values, key)
			count++
		}
		if _, ok := r.sets[key]; ok {
			delete(r.sets, key)
			count++
		}
	}
	return redis.NewIntResult(count, nil)
}

func (r *fakeRealtimeRedis) SAdd(_ context.Context, key string, members ...any) *redis.IntCmd {
	if r.sets[key] == nil {
		r.sets[key] = make(map[string]struct{})
	}
	var count int64
	for _, member := range members {
		value := StringValue(member)
		if _, ok := r.sets[key][value]; !ok {
			count++
		}
		r.sets[key][value] = struct{}{}
	}
	return redis.NewIntResult(count, nil)
}

func (r *fakeRealtimeRedis) SRem(_ context.Context, key string, members ...any) *redis.IntCmd {
	var count int64
	for _, member := range members {
		value := StringValue(member)
		if _, ok := r.sets[key][value]; ok {
			delete(r.sets[key], value)
			count++
		}
	}
	return redis.NewIntResult(count, nil)
}

func (r *fakeRealtimeRedis) SCard(_ context.Context, key string) *redis.IntCmd {
	return redis.NewIntResult(int64(len(r.sets[key])), nil)
}

func (r *fakeRealtimeRedis) SMembers(_ context.Context, key string) *redis.StringSliceCmd {
	values := make([]string, 0, len(r.sets[key]))
	for value := range r.sets[key] {
		values = append(values, value)
	}
	return redis.NewStringSliceResult(values, nil)
}

func (r *fakeRealtimeRedis) Publish(_ context.Context, channel string, message any) *redis.IntCmd {
	r.publishedChannel = channel
	r.publishedMessage = StringValue(message)
	return redis.NewIntResult(1, nil)
}

func (r *fakeRealtimeRedis) Subscribe(context.Context, ...string) *redis.PubSub {
	return nil
}

func StringValue(value any) string {
	switch item := value.(type) {
	case string:
		return item
	case []byte:
		return string(item)
	default:
		raw, _ := json.Marshal(item)
		return string(raw)
	}
}
