package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yuWorm/fba-go/core/config"
)

func TestLoadFromEnvFilePrefersSystemEnvOverDotEnv(t *testing.T) {
	path := writeEnvFile(t, `
ENVIRONMENT='dev'
FASTAPI_API_V1_PATH='/from-dotenv'
REDIS_HOST='dotenv-redis'
REDIS_PORT=6380
TOKEN_SECRET_KEY='dotenv-secret'
WS_NO_AUTH_MARKER='dotenv-marker'
`)
	t.Setenv("FASTAPI_API_V1_PATH", "/from-system")
	t.Setenv("TOKEN_SECRET_KEY", "system-secret")

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if opts.App.APIBasePath != "/from-system" {
		t.Fatalf("APIBasePath = %q, want /from-system", opts.App.APIBasePath)
	}
	if opts.Auth.JWTSecret != "system-secret" {
		t.Fatalf("JWTSecret = %q, want system-secret", opts.Auth.JWTSecret)
	}
	if opts.Redis.Addr != "dotenv-redis:6380" {
		t.Fatalf("Redis.Addr = %q, want dotenv-redis:6380", opts.Redis.Addr)
	}
	if opts.Realtime.NoAuthMarker != "dotenv-marker" {
		t.Fatalf("NoAuthMarker = %q, want dotenv-marker", opts.Realtime.NoAuthMarker)
	}
}

func TestLoadFromEnvFileMapsPythonCoreSettings(t *testing.T) {
	path := writeEnvFile(t, `
ENVIRONMENT='prod'
FASTAPI_TITLE='fba-prod'
FASTAPI_API_V1_PATH='/api/custom'
DATETIME_TIMEZONE='UTC'
DATABASE_TYPE='postgresql'
DATABASE_HOST='db'
DATABASE_PORT=5433
DATABASE_USER='postgres'
DATABASE_PASSWORD='secret'
DATABASE_SCHEMA='fba_app'
REDIS_HOST='redis'
REDIS_PORT=6381
REDIS_PASSWORD='redis-secret'
REDIS_DATABASE=2
REDIS_TIMEOUT=7
TOKEN_SECRET_KEY='token-secret'
TOKEN_EXPIRE_SECONDS=3600
TOKEN_REFRESH_EXPIRE_SECONDS=7200
TOKEN_REDIS_PREFIX='acme:token'
WS_NO_AUTH_MARKER='ws-internal'
CELERY_BROKER_REDIS_DATABASE=3
`)

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if opts.App.Environment != "prod" || opts.App.Name != "fba-prod" || opts.App.APIBasePath != "/api/custom" || opts.App.Timezone != "UTC" {
		t.Fatalf("App options = %+v", opts.App)
	}
	if opts.Database.Driver != "postgresql" {
		t.Fatalf("Database.Driver = %q", opts.Database.Driver)
	}
	for _, part := range []string{"host=db", "port=5433", "user=postgres", "password=secret", "dbname=fba_app"} {
		if !strings.Contains(opts.Database.WriteDSN, part) {
			t.Fatalf("WriteDSN = %q, missing %s", opts.Database.WriteDSN, part)
		}
	}
	if opts.Redis.Addr != "redis:6381" || opts.Redis.Password != "redis-secret" || opts.Redis.DB != 2 {
		t.Fatalf("Redis options = %+v", opts.Redis)
	}
	if opts.Redis.DialTimeout != 7*time.Second || opts.Redis.ReadTimeout != 7*time.Second || opts.Redis.WriteTimeout != 7*time.Second {
		t.Fatalf("Redis timeout = dial %s read %s write %s", opts.Redis.DialTimeout, opts.Redis.ReadTimeout, opts.Redis.WriteTimeout)
	}
	if opts.Redis.KeyPrefix != "acme" {
		t.Fatalf("Redis.KeyPrefix = %q, want acme", opts.Redis.KeyPrefix)
	}
	if opts.Auth.JWTSecret != "token-secret" || opts.Auth.AccessTokenTTL != time.Hour || opts.Auth.RefreshTokenTTL != 2*time.Hour {
		t.Fatalf("Auth options = %+v", opts.Auth)
	}
	if opts.Realtime.NoAuthMarker != "ws-internal" {
		t.Fatalf("Realtime.NoAuthMarker = %q", opts.Realtime.NoAuthMarker)
	}
	if opts.Task.RedisDB != 3 {
		t.Fatalf("Task.RedisDB = %d, want 3", opts.Task.RedisDB)
	}
}

func TestLoadFromEnvFileMapsGoRealtimeMultiInstanceEnv(t *testing.T) {
	path := writeEnvFile(t, `
REALTIME_MULTI_INSTANCE_ENABLED=true
REALTIME_MULTI_INSTANCE_NODE_ID='node-a'
REALTIME_MULTI_INSTANCE_CHANNEL='custom:channel'
REALTIME_DISABLE_POLLING=true
`)

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if !opts.Realtime.MultiInstance.Enabled {
		t.Fatal("MultiInstance.Enabled = false, want true")
	}
	if opts.Realtime.MultiInstance.NodeID != "node-a" {
		t.Fatalf("NodeID = %q, want node-a", opts.Realtime.MultiInstance.NodeID)
	}
	if opts.Realtime.MultiInstance.Channel != "custom:channel" {
		t.Fatalf("Channel = %q, want custom:channel", opts.Realtime.MultiInstance.Channel)
	}
	if opts.Realtime.EnablePolling {
		t.Fatal("EnablePolling = true, want false when REALTIME_DISABLE_POLLING=true")
	}
}

func TestLoadFromEnvFileMapsPythonCORSSettings(t *testing.T) {
	path := writeEnvFile(t, `
MIDDLEWARE_CORS=false
CORS_ALLOWED_ORIGINS='http://localhost:5173,http://127.0.0.1:3000'
CORS_EXPOSE_HEADERS='X-Request-ID,X-Trace-ID'
CORS_ALLOW_CREDENTIALS=false
`)

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if opts.CORS.Enabled {
		t.Fatal("CORS.Enabled = true, want false")
	}
	if got := strings.Join(opts.CORS.AllowedOrigins, ","); got != "http://localhost:5173,http://127.0.0.1:3000" {
		t.Fatalf("CORS.AllowedOrigins = %q", got)
	}
	if got := strings.Join(opts.CORS.ExposeHeaders, ","); got != "X-Request-ID,X-Trace-ID" {
		t.Fatalf("CORS.ExposeHeaders = %q", got)
	}
	if opts.CORS.AllowCredentials {
		t.Fatal("CORS.AllowCredentials = true, want false")
	}
}

func TestLoadFromEnvFileMapsHTTPMiddlewareSettings(t *testing.T) {
	path := writeEnvFile(t, `
MIDDLEWARE_REQUEST_ID=false
MIDDLEWARE_RECOVER=false
MIDDLEWARE_RECOVER_STACK_TRACE=false
MIDDLEWARE_ACCESS_LOG=false
MIDDLEWARE_ACCESS_LOG_SKIP_PATHS=/healthz,/readyz,/metrics
MIDDLEWARE_ERROR_LOG=false
ERROR_RESPONSE_INCLUDE_DETAIL=true
LOG_ACCESS_FILENAME=/tmp/fba-access.log
LOG_ERROR_FILENAME=/tmp/fba-error.log
`)

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if opts.Middleware.RequestID.Enabled {
		t.Fatal("RequestID.Enabled = true, want false")
	}
	if opts.Middleware.Recover.Enabled {
		t.Fatal("Recover.Enabled = true, want false")
	}
	if opts.Middleware.Recover.EnableStackTrace {
		t.Fatal("Recover.EnableStackTrace = true, want false")
	}
	if opts.Middleware.AccessLog.Enabled {
		t.Fatal("AccessLog.Enabled = true, want false")
	}
	if got := strings.Join(opts.Middleware.AccessLog.SkipPaths, ","); got != "/healthz,/readyz,/metrics" {
		t.Fatalf("AccessLog.SkipPaths = %q", got)
	}
	if opts.Middleware.ErrorLog.Enabled {
		t.Fatal("ErrorLog.Enabled = true, want false")
	}
	if !opts.Middleware.ErrorResponse.IncludeDetail {
		t.Fatal("ErrorResponse.IncludeDetail = false, want true")
	}
	if opts.Logger.AccessLogPath != "/tmp/fba-access.log" {
		t.Fatalf("AccessLogPath = %q", opts.Logger.AccessLogPath)
	}
	if opts.Logger.ErrorLogPath != "/tmp/fba-error.log" {
		t.Fatalf("ErrorLogPath = %q", opts.Logger.ErrorLogPath)
	}
}

func TestLoadFromEnvFileMapsIPLocationSettings(t *testing.T) {
	path := writeEnvFile(t, `
IP_LOCATION_PARSE=ip2region
IP_LOCATION_XDB_PATH=/opt/fba/ip2region_v4.xdb
IP_LOCATION_V6_XDB_PATH=/opt/fba/ip2region_v6.xdb
IP_LOCATION_CACHE_POLICY=content
IP_LOCATION_SEARCHERS=8
`)

	opts, err := config.LoadFromEnvFile(path)
	if err != nil {
		t.Fatalf("LoadFromEnvFile() error = %v", err)
	}

	if opts.IPLocation.Provider != "ip2region" {
		t.Fatalf("IPLocation.Provider = %q, want ip2region", opts.IPLocation.Provider)
	}
	if opts.IPLocation.V4XDBPath != "/opt/fba/ip2region_v4.xdb" {
		t.Fatalf("IPLocation.V4XDBPath = %q, want v4 path", opts.IPLocation.V4XDBPath)
	}
	if opts.IPLocation.V6XDBPath != "/opt/fba/ip2region_v6.xdb" {
		t.Fatalf("IPLocation.V6XDBPath = %q, want v6 path", opts.IPLocation.V6XDBPath)
	}
	if opts.IPLocation.CachePolicy != "content" {
		t.Fatalf("IPLocation.CachePolicy = %q, want content", opts.IPLocation.CachePolicy)
	}
	if opts.IPLocation.Searchers != 8 {
		t.Fatalf("IPLocation.Searchers = %d, want 8", opts.IPLocation.Searchers)
	}
}

func writeEnvFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return path
}
