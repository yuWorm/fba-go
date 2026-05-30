package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestRedisInvalidatorPublishesPythonCompatiblePayload(t *testing.T) {
	publisher := &capturePublisher{}
	invalidator := RedisInvalidator{
		publisher: publisher,
		channel:   "fba:cache:invalidate",
		cacheKey:  "fba:cache:dict",
	}

	if err := invalidator.InvalidateDict(context.Background()); err != nil {
		t.Fatalf("InvalidateDict() error = %v", err)
	}

	if publisher.channel != "fba:cache:invalidate" {
		t.Fatalf("channel = %q, want fba:cache:invalidate", publisher.channel)
	}
	var payload map[string]any
	raw, ok := publisher.message.([]byte)
	if !ok {
		t.Fatalf("message = %T, want []byte", publisher.message)
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload["key"] != "fba:cache:dict" {
		t.Fatalf("key = %v, want fba:cache:dict", payload["key"])
	}
	if payload["is_delete_prefix"] != true {
		t.Fatalf("is_delete_prefix = %v, want true", payload["is_delete_prefix"])
	}
}

type capturePublisher struct {
	channel string
	message any
}

func (p *capturePublisher) Publish(_ context.Context, channel string, message any) *redis.IntCmd {
	p.channel = channel
	p.message = message
	return redis.NewIntResult(1, nil)
}
