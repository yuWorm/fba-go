package config

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	App      AppOptions
	Fiber    fiber.Config
	Logger   LoggerOptions
	Database DatabaseOptions
	Redis    RedisOptions
	Auth     AuthOptions
	Realtime RealtimeOptions
	Task     TaskOptions
	Pools    map[string]PoolOptions
	Hooks    Hooks
}

type AppOptions struct {
	Name        string
	Version     string
	Environment string
	APIBasePath string
	Timezone    string
}

type Hook func(context.Context) error

type Hooks struct {
	OnStart    []Hook
	OnShutdown []Hook
}

type LoggerOptions struct {
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
	AccessLogPath    string
	ErrorLogPath     string
	Rotation         RotationOptions
}

type RotationOptions struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type RedisOptions struct {
	Mode string

	Addr  string
	Addrs []string

	Username string
	Password string
	DB       int

	MasterName string

	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	KeyPrefix string
}

type DatabaseOptions struct {
	Driver string

	WriteDSN string
	ReadDSN  string

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	AutoMigrate      bool
	MigrationLockKey string
}

type AuthOptions struct {
	JWTSecret       string
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type RealtimeOptions struct {
	Disabled       bool
	Path           string
	Namespace      string
	NoAuthMarker   string
	EnablePolling  bool
	DisablePolling bool
	MultiInstance  RealtimeMultiInstanceOptions
}

type RealtimeMultiInstanceOptions struct {
	Enabled bool
	NodeID  string
	Channel string
}

type TaskOptions struct {
	Enabled bool

	RedisMode       string
	RedisAddr       string
	RedisAddrs      []string
	RedisDB         int
	RedisPassword   string
	RedisMasterName string

	Concurrency int
	Queues      map[string]int

	SchedulerEnabled bool
	SchedulerLockKey string
	SchedulerLockTTL time.Duration
}

type PoolOptions struct {
	MaxWorkers int
	QueueSize  int
}

func (o Options) WithDefaults() Options {
	if o.App.APIBasePath == "" {
		o.App.APIBasePath = "/api/v1"
	}
	if o.App.Timezone == "" {
		o.App.Timezone = "Asia/Shanghai"
	}
	if o.App.Environment == "" {
		o.App.Environment = "dev"
	}
	if o.Realtime.Path == "" {
		o.Realtime.Path = "/ws/socket.io"
	}
	if o.Realtime.Namespace == "" {
		o.Realtime.Namespace = "/ws"
	}
	if o.Realtime.NoAuthMarker == "" {
		o.Realtime.NoAuthMarker = "internal"
	}
	if !o.Realtime.DisablePolling {
		o.Realtime.EnablePolling = true
	}
	if o.Realtime.MultiInstance.Channel == "" {
		prefix := o.Redis.KeyPrefix
		if prefix == "" {
			prefix = "fba"
		}
		o.Realtime.MultiInstance.Channel = prefix + ":realtime:broadcast"
	}
	return o
}
