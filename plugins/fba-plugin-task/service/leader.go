package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/redisx"
)

type LeaderLease interface {
	Acquire(context.Context) (bool, error)
	Release(context.Context) error
}

type NoopLeaderLease struct{}

func (NoopLeaderLease) Acquire(context.Context) (bool, error) {
	return true, nil
}

func (NoopLeaderLease) Release(context.Context) error {
	return nil
}

type RedisLeaderLease struct {
	client redisLeaseClient
	key    string
	value  string
	ttl    time.Duration
}

type redisLeaseClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

func NewRedisLeaderLease(client redisx.RedisClient, key string, value string, ttl time.Duration) RedisLeaderLease {
	return RedisLeaderLease{client: client, key: key, value: value, ttl: ttl}
}

func (l RedisLeaderLease) Acquire(ctx context.Context) (bool, error) {
	if l.client == nil || l.key == "" {
		return true, nil
	}
	if l.ttl <= 0 {
		l.ttl = 30 * time.Second
	}
	acquired, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	return acquired, err
}

func (l RedisLeaderLease) Release(ctx context.Context) error {
	if l.client == nil || l.key == "" {
		return nil
	}
	return l.client.Del(ctx, l.key).Err()
}
