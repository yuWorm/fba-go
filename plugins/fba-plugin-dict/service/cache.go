package service

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/redisx"
)

type CacheInvalidator interface {
	InvalidateDict(context.Context) error
}

type NoopInvalidator struct{}

func (NoopInvalidator) InvalidateDict(context.Context) error {
	return nil
}

type RedisInvalidator struct {
	publisher publisher
	channel   string
	cacheKey  string
}

type publisher interface {
	Publish(ctx context.Context, channel string, message any) *redis.IntCmd
}

func NewRedisInvalidator(client redisx.RedisClient, channel string, cacheKey string) RedisInvalidator {
	return RedisInvalidator{publisher: client, channel: channel, cacheKey: cacheKey}
}

func (i RedisInvalidator) InvalidateDict(ctx context.Context) error {
	if i.publisher == nil || i.channel == "" || i.cacheKey == "" {
		return nil
	}
	payload, err := json.Marshal(map[string]any{
		"key":              i.cacheKey,
		"is_delete_prefix": true,
	})
	if err != nil {
		return err
	}
	return i.publisher.Publish(ctx, i.channel, payload).Err()
}
