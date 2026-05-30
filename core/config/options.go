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
	return o
}
