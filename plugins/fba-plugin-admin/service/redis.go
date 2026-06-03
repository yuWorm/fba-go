package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Info(ctx context.Context, section ...string) *redis.StringCmd
}

const (
	defaultRedisKeyPrefix        = "fba"
	defaultEmailCaptchaKeyPrefix = "fba:email:captcha"
)
