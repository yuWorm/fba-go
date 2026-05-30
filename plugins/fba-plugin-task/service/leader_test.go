package service

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRedisLeaderLeaseUsesSchedulerLeaderKey(t *testing.T) {
	client := &captureRedisLeaseClient{}
	lease := RedisLeaderLease{
		client: client,
		key:    "fba:task:scheduler:leader",
		value:  "node-1",
		ttl:    30 * time.Second,
	}

	acquired, err := lease.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	if !acquired {
		t.Fatal("acquired = false, want true")
	}
	if client.setKey != "fba:task:scheduler:leader" {
		t.Fatalf("SetNX key = %q, want fba:task:scheduler:leader", client.setKey)
	}
	if client.ttl != 30*time.Second {
		t.Fatalf("ttl = %s, want 30s", client.ttl)
	}

	if err := lease.Release(context.Background()); err != nil {
		t.Fatalf("Release() error = %v", err)
	}
	if client.delKey != "fba:task:scheduler:leader" {
		t.Fatalf("Del key = %q, want fba:task:scheduler:leader", client.delKey)
	}
}

type captureRedisLeaseClient struct {
	setKey string
	delKey string
	ttl    time.Duration
}

func (c *captureRedisLeaseClient) SetNX(_ context.Context, key string, _ any, expiration time.Duration) *redis.BoolCmd {
	c.setKey = key
	c.ttl = expiration
	return redis.NewBoolResult(true, nil)
}

func (c *captureRedisLeaseClient) Del(_ context.Context, keys ...string) *redis.IntCmd {
	if len(keys) > 0 {
		c.delKey = keys[0]
	}
	return redis.NewIntResult(1, nil)
}
