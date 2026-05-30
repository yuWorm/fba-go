package redisx

import (
	"github.com/redis/go-redis/v9"
	"github.com/yuWorm/fba-go/core/config"
)

type RedisClient interface {
	redis.UniversalClient
}

func UniversalOptions(opts config.RedisOptions) *redis.UniversalOptions {
	addrs := opts.Addrs
	if len(addrs) == 0 && opts.Addr != "" {
		addrs = []string{opts.Addr}
	}
	if len(addrs) == 0 {
		addrs = []string{"127.0.0.1:6379"}
	}

	return &redis.UniversalOptions{
		Addrs:         addrs,
		DB:            opts.DB,
		Username:      opts.Username,
		Password:      opts.Password,
		MasterName:    sentinelMasterName(opts),
		PoolSize:      opts.PoolSize,
		MinIdleConns:  opts.MinIdleConns,
		DialTimeout:   opts.DialTimeout,
		ReadTimeout:   opts.ReadTimeout,
		WriteTimeout:  opts.WriteTimeout,
		IsClusterMode: opts.Mode == "cluster",
	}
}

func NewUniversalClient(opts config.RedisOptions) redis.UniversalClient {
	return redis.NewUniversalClient(UniversalOptions(opts))
}

func sentinelMasterName(opts config.RedisOptions) string {
	if opts.Mode != "sentinel" {
		return ""
	}
	return opts.MasterName
}
