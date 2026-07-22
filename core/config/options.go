package config

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Options struct {
	App        AppOptions
	Fiber      fiber.Config
	Logger     LoggerOptions
	Middleware MiddlewareOptions
	CORS       CORSOptions
	Database   DatabaseOptions
	Redis      RedisOptions
	Auth       AuthOptions
	IPLocation IPLocationOptions
	Realtime   RealtimeOptions
	Task       TaskOptions
	Pools      map[string]PoolOptions
	Hooks      Hooks
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
	Level string
	// Empty Encoding keeps machine-readable JSON for files while rendering
	// stdout and stderr with the human-oriented colored console encoder.
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
	AccessLogPath    string
	ErrorLogPath     string
	Rotation         RotationOptions
}

type MiddlewareOptions struct {
	RequestID     RequestIDOptions
	Recover       RecoverOptions
	AccessLog     AccessLogOptions
	ErrorLog      ErrorLogOptions
	ErrorResponse ErrorResponseOptions
}

type RequestIDOptions struct {
	Enabled  bool
	Disabled bool

	// enabledSet preserves explicit false from env parsing; otherwise diagnostics
	// middleware defaults to enabled for Python-compatible traceability.
	enabledSet bool
}

type RecoverOptions struct {
	Enabled          bool
	Disabled         bool
	EnableStackTrace bool

	enabledSet    bool
	stackTraceSet bool
}

type AccessLogOptions struct {
	Enabled   bool
	Disabled  bool
	SkipPaths []string

	enabledSet bool
}

type ErrorLogOptions struct {
	Enabled  bool
	Disabled bool

	enabledSet bool
}

type ErrorResponseOptions struct {
	IncludeDetail bool
	HideDetail    bool

	includeDetailSet bool
}

type CORSOptions struct {
	Enabled          bool
	Disabled         bool
	AllowedOrigins   []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string

	// enabledSet distinguishes an explicit MIDDLEWARE_CORS=false from the zero
	// value Options{}, where CORS should default to Python-compatible enabled.
	enabledSet          bool
	allowCredentialsSet bool
}

var ErrCORSWildcardCredentials = errors.New("CORS wildcard origin cannot be combined with credentials")

func ValidateCORSOptions(opts CORSOptions) error {
	if !opts.Enabled || !opts.AllowCredentials {
		return nil
	}
	for _, origin := range opts.AllowedOrigins {
		if strings.TrimSpace(origin) == "*" {
			return ErrCORSWildcardCredentials
		}
	}
	return nil
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
	JWTSecret              string
	JWTIssuer              string
	AdminBootstrapPassword string
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
}

type IPLocationOptions struct {
	Provider    string
	V4XDBPath   string
	V6XDBPath   string
	CachePolicy string
	Searchers   int
}

type RealtimeOptions struct {
	Enabled        bool
	Disabled       bool
	Path           string
	Namespace      string
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
	o.Middleware.RequestID.Enabled = defaultEnabled(o.Middleware.RequestID.Disabled, o.Middleware.RequestID.Enabled, o.Middleware.RequestID.enabledSet)
	o.Middleware.Recover.Enabled = defaultEnabled(o.Middleware.Recover.Disabled, o.Middleware.Recover.Enabled, o.Middleware.Recover.enabledSet)
	if !o.Middleware.Recover.stackTraceSet {
		o.Middleware.Recover.EnableStackTrace = true
	}
	o.Middleware.AccessLog.Enabled = defaultEnabled(o.Middleware.AccessLog.Disabled, o.Middleware.AccessLog.Enabled, o.Middleware.AccessLog.enabledSet)
	if len(o.Middleware.AccessLog.SkipPaths) == 0 {
		o.Middleware.AccessLog.SkipPaths = []string{"/healthz", "/readyz", "/metrics"}
	}
	o.Middleware.ErrorLog.Enabled = defaultEnabled(o.Middleware.ErrorLog.Disabled, o.Middleware.ErrorLog.Enabled, o.Middleware.ErrorLog.enabledSet)
	if !strings.EqualFold(o.App.Environment, "dev") || o.Middleware.ErrorResponse.HideDetail {
		o.Middleware.ErrorResponse.IncludeDetail = false
	} else if !o.Middleware.ErrorResponse.includeDetailSet && !o.Middleware.ErrorResponse.IncludeDetail {
		o.Middleware.ErrorResponse.IncludeDetail = true
	}
	if o.CORS.Disabled {
		o.CORS.Enabled = false
	} else if !o.CORS.enabledSet {
		o.CORS.Enabled = true
	}
	if len(o.CORS.AllowedOrigins) == 0 {
		o.CORS.AllowedOrigins = []string{"http://127.0.0.1", "http://localhost:5173"}
	}
	if !o.CORS.allowCredentialsSet {
		o.CORS.AllowCredentials = true
	}
	if len(o.CORS.AllowMethods) == 0 {
		o.CORS.AllowMethods = []string{"*"}
	}
	if len(o.CORS.AllowHeaders) == 0 {
		o.CORS.AllowHeaders = []string{"*"}
	}
	if len(o.CORS.ExposeHeaders) == 0 {
		o.CORS.ExposeHeaders = []string{"X-Request-ID"}
	}
	if o.IPLocation.Provider == "" {
		o.IPLocation.Provider = "none"
	}
	if o.IPLocation.CachePolicy == "" {
		o.IPLocation.CachePolicy = "vectorIndex"
	}
	if o.IPLocation.Searchers <= 0 {
		o.IPLocation.Searchers = 20
	}
	if !o.Realtime.Enabled {
		o.Realtime.Disabled = true
	}
	if o.Realtime.Disabled {
		o.Realtime.Enabled = false
	}
	if o.Realtime.Path == "" {
		o.Realtime.Path = "/ws/socket.io"
	}
	if o.Realtime.Namespace == "" {
		o.Realtime.Namespace = "/ws"
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

func defaultEnabled(disabled bool, enabled bool, set bool) bool {
	if disabled {
		return false
	}
	if !set {
		return true
	}
	return enabled
}
