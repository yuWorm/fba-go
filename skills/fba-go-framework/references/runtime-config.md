# Runtime Configuration

## Source Map

- Config types and defaults: `core/config/options.go`
- Environment loading: `core/config/env.go`
- HTTP setup and CORS: `core/fiberx/app.go`
- Admin runtime database provisioning: `../fba-go-admin/internal/runtime/runtime.go`

## Loading Rules

`fba.LoadOptionsFromEnv` reads `FBA_ENV_FILE` or `.env` by default, then overlays real environment variables on top of dotenv values.

This mirrors Python settings behavior: deployment environment values override checked-in dotenv values.

## Important Defaults

`Options.WithDefaults()` sets:

- API base path: `/api/v1`
- timezone: `Asia/Shanghai`
- environment: `dev`
- request ID middleware enabled
- recover middleware enabled with stack capture
- access log middleware enabled, skipping `/healthz`, `/readyz`, and `/metrics`
- error log middleware enabled for HTTP 500 errors
- internal error response details enabled in `dev` and hidden outside `dev`
- CORS enabled unless explicitly disabled
- CORS origins: `http://127.0.0.1`, `http://localhost:5173`
- CORS credentials: enabled
- CORS methods and headers: `*`
- CORS exposed headers: `X-Request-ID`
- realtime path: `/ws/socket.io`
- realtime namespace: `/ws`
- realtime no-auth marker: `internal`
- realtime polling enabled unless disabled
- realtime multi-instance channel based on Redis key prefix, defaulting to `fba:realtime:broadcast`

When changing defaults, check Python compatibility first if the setting maps to `sources/fastapi-best-architecture/`.

## Environment Keys

Application:

- `ENVIRONMENT`
- `FASTAPI_TITLE`
- `FASTAPI_API_V1_PATH`
- `DATETIME_TIMEZONE`

Database:

- `DATABASE_TYPE` or `DATABASE_DRIVER`
- `DATABASE_DSN` or `DATABASE_WRITE_DSN`
- `DATABASE_READ_DSN`
- `DATABASE_AUTO_MIGRATE`
- `DATABASE_MIGRATION_LOCK_KEY`

Redis:

- `REDIS_MODE`
- `REDIS_ADDR` or `REDIS_HOST` and `REDIS_PORT`
- `REDIS_ADDRS`
- `REDIS_USERNAME`
- `REDIS_PASSWORD`
- `REDIS_DATABASE` or `REDIS_DB`
- `REDIS_MASTER_NAME`
- `REDIS_POOL_SIZE`
- `REDIS_MIN_IDLE_CONNS`
- `REDIS_TIMEOUT`
- Redis key prefix variables handled by `redisKeyPrefix`

Auth:

- `TOKEN_SECRET_KEY`
- `TOKEN_ISSUER` or `JWT_ISSUER`
- `TOKEN_EXPIRE_SECONDS`
- `TOKEN_REFRESH_EXPIRE_SECONDS`

CORS:

- `MIDDLEWARE_CORS` or `CORS_ENABLED`
- `CORS_ALLOWED_ORIGINS`
- `CORS_ALLOW_METHODS`
- `CORS_ALLOW_HEADERS`
- `CORS_EXPOSE_HEADERS`
- `CORS_ALLOW_CREDENTIALS`

HTTP diagnostics and logging:

- `MIDDLEWARE_REQUEST_ID`
- `MIDDLEWARE_RECOVER`
- `MIDDLEWARE_RECOVER_STACK_TRACE`
- `MIDDLEWARE_ACCESS_LOG`
- `MIDDLEWARE_ACCESS_LOG_SKIP_PATHS`
- `MIDDLEWARE_ERROR_LOG`
- `ERROR_RESPONSE_INCLUDE_DETAIL`
- `LOG_STD_LEVEL` or `LOG_LEVEL`
- `LOG_ACCESS_FILENAME`
- `LOG_ERROR_FILENAME`

Realtime:

- `REALTIME_DISABLED`
- `REALTIME_PATH` or `SOCKETIO_PATH`
- `REALTIME_NAMESPACE` or `SOCKETIO_NAMESPACE`
- `REALTIME_DISABLE_POLLING`
- `REALTIME_ENABLE_POLLING`
- `REALTIME_MULTI_INSTANCE_ENABLED`
- `REALTIME_MULTI_INSTANCE_NODE_ID`
- `REALTIME_MULTI_INSTANCE_CHANNEL`

Task:

- `TASK_ENABLED`
- task Redis fields
- `TASK_CONCURRENCY`
- `TASK_QUEUES`
- scheduler lock fields

## Database Provisioning

The admin runtime provides `db.Provider` before plugin registration when database config is present. This is required because modules decide during registration whether to use memory repositories or GORM repositories and migrations.

If `Database.Driver` and `Database.WriteDSN` are empty, runtime does not open a DB. Plugins should keep memory fallback working in this mode.

## CORS

`fiberx.New` installs Fiber's CORS middleware only when `opts.CORS.Enabled` is true. Default behavior is enabled and Python-compatible.

Do not add per-plugin CORS logic. CORS is an application-level concern.

## HTTP Diagnostics

`fiberx.New` installs request ID, HTTP logging, recover, and CORS middleware from application-level config. Request IDs feed the `trace_id` response field and both access/error logs. Error responses keep production-safe messages unless `ERROR_RESPONSE_INCLUDE_DETAIL=true`; server logs still record the original error detail for HTTP 500 responses.

Logger encoding is sink-aware by default:

- `stdout` and `stderr` use human-readable console output with ISO-8601 time
  and ANSI-colored levels.
- File sinks, including `LOG_ACCESS_FILENAME` and `LOG_ERROR_FILENAME`, use
  structured JSON.
- File logging remains mirrored to the corresponding console sink.
- Programmatic `LoggerOptions.Encoding` values `json` or `console` explicitly
  force that encoder for every sink; leaving it empty enables the split default.

## Realtime

`core/app.New` always provides `realtime.Hub` and `realtime.OnlineStore`. When realtime is not disabled, it mounts the Socket.IO server and registers shutdown hooks.

When multi-instance realtime is enabled, Redis client and broadcaster are provided and started through hooks.
