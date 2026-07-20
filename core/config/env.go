package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const DefaultEnvFile = ".env"

func LoadFromEnv() (Options, error) {
	path := os.Getenv("FBA_ENV_FILE")
	if path == "" {
		path = DefaultEnvFile
	}
	return LoadFromEnvFile(path)
}

func LoadFromEnvFile(path string) (Options, error) {
	values, err := readDotEnv(path)
	if err != nil {
		return Options{}, err
	}
	// Pydantic settings uses real environment variables before dotenv values.
	// Preserve that precedence so deployment-level env overrides checked-in .env.
	for _, entry := range os.Environ() {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	return optionsFromEnv(values).WithDefaults(), nil
}

func readDotEnv(path string) (map[string]string, error) {
	values := make(map[string]string)
	if path == "" {
		return values, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("parse dotenv %s:%d: missing '='", path, lineNo)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("parse dotenv %s:%d: empty key", path, lineNo)
		}
		values[key] = parseDotEnvValue(strings.TrimSpace(value))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func parseDotEnvValue(value string) string {
	if value == "" {
		return ""
	}
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '\'' || quote == '"') && value[len(value)-1] == quote {
			value = value[1 : len(value)-1]
			if quote == '"' {
				value = strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, `"`, `\\`, `\`).Replace(value)
			}
			return value
		}
	}
	if index := strings.Index(value, " #"); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func optionsFromEnv(values map[string]string) Options {
	var opts Options
	applyAppEnv(&opts, values)
	applyDatabaseEnv(&opts, values)
	applyRedisEnv(&opts, values)
	applyAuthEnv(&opts, values)
	applyIPLocationEnv(&opts, values)
	applyCORSEnv(&opts, values)
	applyMiddlewareEnv(&opts, values)
	applyRealtimeEnv(&opts, values)
	applyTaskEnv(&opts, values)
	applyLoggerEnv(&opts, values)
	return opts
}

func applyAppEnv(opts *Options, values map[string]string) {
	if value, ok := values["ENVIRONMENT"]; ok {
		opts.App.Environment = value
	}
	if value, ok := values["FASTAPI_TITLE"]; ok {
		opts.App.Name = value
	}
	if value, ok := values["FASTAPI_API_V1_PATH"]; ok {
		opts.App.APIBasePath = value
	}
	if value, ok := values["DATETIME_TIMEZONE"]; ok {
		opts.App.Timezone = value
	}
}

func applyDatabaseEnv(opts *Options, values map[string]string) {
	driver := firstEnv(values, "DATABASE_TYPE", "DATABASE_DRIVER")
	if driver != "" {
		opts.Database.Driver = driver
	}
	if value := firstEnv(values, "DATABASE_DSN", "DATABASE_WRITE_DSN"); value != "" {
		opts.Database.WriteDSN = value
	}
	if value := firstEnv(values, "DATABASE_READ_DSN"); value != "" {
		opts.Database.ReadDSN = value
	}
	if opts.Database.WriteDSN == "" && driver != "" {
		opts.Database.WriteDSN = databaseDSN(values, driver)
	}
	if value, ok := envBool(values, "DATABASE_AUTO_MIGRATE"); ok {
		opts.Database.AutoMigrate = value
	}
	if value := firstEnv(values, "DATABASE_MIGRATION_LOCK_KEY"); value != "" {
		opts.Database.MigrationLockKey = value
	}
}

func applyRedisEnv(opts *Options, values map[string]string) {
	if value := firstEnv(values, "REDIS_MODE"); value != "" {
		opts.Redis.Mode = value
	}
	if value := firstEnv(values, "REDIS_ADDR"); value != "" {
		opts.Redis.Addr = value
	} else if host := firstEnv(values, "REDIS_HOST"); host != "" {
		port := firstEnv(values, "REDIS_PORT")
		if port == "" {
			port = "6379"
		}
		opts.Redis.Addr = host + ":" + port
	}
	if value := firstEnv(values, "REDIS_ADDRS"); value != "" {
		opts.Redis.Addrs = splitList(value)
	}
	if value, ok := values["REDIS_USERNAME"]; ok {
		opts.Redis.Username = value
	}
	if value, ok := values["REDIS_PASSWORD"]; ok {
		opts.Redis.Password = value
	}
	if value, ok := envInt(values, "REDIS_DATABASE", "REDIS_DB"); ok {
		opts.Redis.DB = value
	}
	if value := firstEnv(values, "REDIS_MASTER_NAME"); value != "" {
		opts.Redis.MasterName = value
	}
	if value, ok := envInt(values, "REDIS_POOL_SIZE"); ok {
		opts.Redis.PoolSize = value
	}
	if value, ok := envInt(values, "REDIS_MIN_IDLE_CONNS"); ok {
		opts.Redis.MinIdleConns = value
	}
	if timeout, ok := envDurationSeconds(values, "REDIS_TIMEOUT"); ok {
		opts.Redis.DialTimeout = timeout
		opts.Redis.ReadTimeout = timeout
		opts.Redis.WriteTimeout = timeout
	}
	if value := redisKeyPrefix(values); value != "" {
		opts.Redis.KeyPrefix = value
	}
}

func applyAuthEnv(opts *Options, values map[string]string) {
	if value, ok := values["TOKEN_SECRET_KEY"]; ok {
		opts.Auth.JWTSecret = value
	}
	if value := firstEnv(values, "TOKEN_ISSUER", "JWT_ISSUER"); value != "" {
		opts.Auth.JWTIssuer = value
	}
	if ttl, ok := envDurationSeconds(values, "TOKEN_EXPIRE_SECONDS"); ok {
		opts.Auth.AccessTokenTTL = ttl
	}
	if ttl, ok := envDurationSeconds(values, "TOKEN_REFRESH_EXPIRE_SECONDS"); ok {
		opts.Auth.RefreshTokenTTL = ttl
	}
}

func applyIPLocationEnv(opts *Options, values map[string]string) {
	if value := firstEnv(values, "IP_LOCATION_PARSE", "IP_LOCATION_PROVIDER"); value != "" {
		opts.IPLocation.Provider = value
	}
	if value := firstEnv(values, "IP_LOCATION_V4_XDB_PATH", "IP_LOCATION_XDB_PATH", "IP_LOCATION_DB_PATH"); value != "" {
		opts.IPLocation.V4XDBPath = value
	}
	if value := firstEnv(values, "IP_LOCATION_V6_XDB_PATH"); value != "" {
		opts.IPLocation.V6XDBPath = value
	}
	if value := firstEnv(values, "IP_LOCATION_CACHE_POLICY"); value != "" {
		opts.IPLocation.CachePolicy = value
	}
	if value, ok := envInt(values, "IP_LOCATION_SEARCHERS"); ok {
		opts.IPLocation.Searchers = value
	}
}

func applyCORSEnv(opts *Options, values map[string]string) {
	if value, ok := envBool(values, "MIDDLEWARE_CORS", "CORS_ENABLED"); ok {
		opts.CORS.Enabled = value
		opts.CORS.Disabled = !value
		opts.CORS.enabledSet = true
	}
	if value := firstEnv(values, "CORS_ALLOWED_ORIGINS"); value != "" {
		opts.CORS.AllowedOrigins = splitList(value)
	}
	if value := firstEnv(values, "CORS_ALLOW_METHODS"); value != "" {
		opts.CORS.AllowMethods = splitList(value)
	}
	if value := firstEnv(values, "CORS_ALLOW_HEADERS"); value != "" {
		opts.CORS.AllowHeaders = splitList(value)
	}
	if value := firstEnv(values, "CORS_EXPOSE_HEADERS"); value != "" {
		opts.CORS.ExposeHeaders = splitList(value)
	}
	if value, ok := envBool(values, "CORS_ALLOW_CREDENTIALS"); ok {
		opts.CORS.AllowCredentials = value
		opts.CORS.allowCredentialsSet = true
	}
}

func applyMiddlewareEnv(opts *Options, values map[string]string) {
	if value, ok := envBool(values, "MIDDLEWARE_REQUEST_ID"); ok {
		opts.Middleware.RequestID.Enabled = value
		opts.Middleware.RequestID.Disabled = !value
		opts.Middleware.RequestID.enabledSet = true
	}
	if value, ok := envBool(values, "MIDDLEWARE_RECOVER"); ok {
		opts.Middleware.Recover.Enabled = value
		opts.Middleware.Recover.Disabled = !value
		opts.Middleware.Recover.enabledSet = true
	}
	if value, ok := envBool(values, "MIDDLEWARE_RECOVER_STACK_TRACE"); ok {
		opts.Middleware.Recover.EnableStackTrace = value
		opts.Middleware.Recover.stackTraceSet = true
	}
	if value, ok := envBool(values, "MIDDLEWARE_ACCESS_LOG"); ok {
		opts.Middleware.AccessLog.Enabled = value
		opts.Middleware.AccessLog.Disabled = !value
		opts.Middleware.AccessLog.enabledSet = true
	}
	if value := firstEnv(values, "MIDDLEWARE_ACCESS_LOG_SKIP_PATHS"); value != "" {
		opts.Middleware.AccessLog.SkipPaths = splitList(value)
	}
	if value, ok := envBool(values, "MIDDLEWARE_ERROR_LOG"); ok {
		opts.Middleware.ErrorLog.Enabled = value
		opts.Middleware.ErrorLog.Disabled = !value
		opts.Middleware.ErrorLog.enabledSet = true
	}
	if value, ok := envBool(values, "ERROR_RESPONSE_INCLUDE_DETAIL"); ok {
		opts.Middleware.ErrorResponse.IncludeDetail = value
		opts.Middleware.ErrorResponse.HideDetail = !value
		opts.Middleware.ErrorResponse.includeDetailSet = true
	}
}

func applyRealtimeEnv(opts *Options, values map[string]string) {
	if value, ok := envBool(values, "REALTIME_DISABLED"); ok {
		opts.Realtime.Disabled = value
		opts.Realtime.Enabled = !value
	}
	if value := firstEnv(values, "REALTIME_PATH", "SOCKETIO_PATH"); value != "" {
		opts.Realtime.Path = value
	}
	if value := firstEnv(values, "REALTIME_NAMESPACE", "SOCKETIO_NAMESPACE"); value != "" {
		opts.Realtime.Namespace = value
	}
	if value, ok := envBool(values, "REALTIME_DISABLE_POLLING"); ok {
		opts.Realtime.DisablePolling = value
	}
	if value, ok := envBool(values, "REALTIME_ENABLE_POLLING"); ok {
		opts.Realtime.EnablePolling = value
		if !value {
			opts.Realtime.DisablePolling = true
		}
	}
	if value, ok := envBool(values, "REALTIME_MULTI_INSTANCE_ENABLED"); ok {
		opts.Realtime.MultiInstance.Enabled = value
	}
	if value := firstEnv(values, "REALTIME_MULTI_INSTANCE_NODE_ID"); value != "" {
		opts.Realtime.MultiInstance.NodeID = value
	}
	if value := firstEnv(values, "REALTIME_MULTI_INSTANCE_CHANNEL"); value != "" {
		opts.Realtime.MultiInstance.Channel = value
	}
}

func applyTaskEnv(opts *Options, values map[string]string) {
	if value, ok := envBool(values, "TASK_ENABLED"); ok {
		opts.Task.Enabled = value
	}
	if value := firstEnv(values, "TASK_REDIS_MODE"); value != "" {
		opts.Task.RedisMode = value
	}
	if value := firstEnv(values, "TASK_REDIS_ADDR"); value != "" {
		opts.Task.RedisAddr = value
	}
	if value := firstEnv(values, "TASK_REDIS_ADDRS"); value != "" {
		opts.Task.RedisAddrs = splitList(value)
	}
	if value, ok := envInt(values, "CELERY_BROKER_REDIS_DATABASE", "TASK_REDIS_DB"); ok {
		opts.Task.RedisDB = value
	}
	if value, ok := values["TASK_REDIS_PASSWORD"]; ok {
		opts.Task.RedisPassword = value
	}
	if value := firstEnv(values, "TASK_REDIS_MASTER_NAME"); value != "" {
		opts.Task.RedisMasterName = value
	}
	if value, ok := envInt(values, "TASK_CONCURRENCY"); ok {
		opts.Task.Concurrency = value
	}
	if value := firstEnv(values, "TASK_QUEUES"); value != "" {
		opts.Task.Queues = parseTaskQueues(value)
	}
	if value, ok := envBool(values, "TASK_SCHEDULER_ENABLED"); ok {
		opts.Task.SchedulerEnabled = value
	}
	if value := firstEnv(values, "TASK_SCHEDULER_LOCK_KEY"); value != "" {
		opts.Task.SchedulerLockKey = value
	}
	if ttl, ok := envDurationSeconds(values, "TASK_SCHEDULER_LOCK_TTL_SECONDS"); ok {
		opts.Task.SchedulerLockTTL = ttl
	}
}

func applyLoggerEnv(opts *Options, values map[string]string) {
	if value := firstEnv(values, "LOG_STD_LEVEL", "LOG_LEVEL"); value != "" {
		opts.Logger.Level = strings.ToLower(value)
	}
	if value := firstEnv(values, "LOG_ACCESS_FILENAME"); value != "" {
		opts.Logger.AccessLogPath = value
	}
	if value := firstEnv(values, "LOG_ERROR_FILENAME"); value != "" {
		opts.Logger.ErrorLogPath = value
	}
}

func databaseDSN(values map[string]string, driver string) string {
	host := firstEnv(values, "DATABASE_HOST")
	port := firstEnv(values, "DATABASE_PORT")
	user := firstEnv(values, "DATABASE_USER")
	password := firstEnv(values, "DATABASE_PASSWORD")
	database := firstEnv(values, "DATABASE_SCHEMA", "DATABASE_NAME")
	charset := firstEnv(values, "DATABASE_CHARSET")
	if charset == "" {
		charset = "utf8mb4"
	}
	switch driver {
	case "postgresql", "postgres":
		if port == "" {
			port = "5432"
		}
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
	case "mysql":
		if port == "" {
			port = "3306"
		}
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", user, password, host, port, database, charset)
	default:
		return ""
	}
}

func redisKeyPrefix(values map[string]string) string {
	if value := firstEnv(values, "REDIS_KEY_PREFIX"); value != "" {
		return strings.TrimRight(value, ":")
	}
	candidates := map[string]string{
		"TOKEN_REDIS_PREFIX":            ":token",
		"TOKEN_ONLINE_REDIS_PREFIX":     ":token_online",
		"TOKEN_REFRESH_REDIS_PREFIX":    ":refresh_token",
		"TOKEN_EXTRA_INFO_REDIS_PREFIX": ":token_extra_info",
		"JWT_USER_REDIS_PREFIX":         ":user",
		"LOGIN_CAPTCHA_REDIS_PREFIX":    ":login:captcha",
		"LOGIN_FAILURE_PREFIX":          ":login:failure",
		"USER_LOCK_REDIS_PREFIX":        ":user:lock",
		"CACHE_DICT_REDIS_PREFIX":       ":cache:dict",
		"CACHE_CONFIG_REDIS_PREFIX":     ":cache:config",
		"PLUGIN_REDIS_PREFIX":           ":plugin",
	}
	for key, suffix := range candidates {
		value := firstEnv(values, key)
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	return ""
}

func firstEnv(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value
		}
	}
	return ""
}

func envInt(values map[string]string, keys ...string) (int, bool) {
	raw := firstEnv(values, keys...)
	if raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}

func envDurationSeconds(values map[string]string, keys ...string) (time.Duration, bool) {
	value, ok := envInt(values, keys...)
	if !ok {
		return 0, false
	}
	return time.Duration(value) * time.Second, true
}

func envBool(values map[string]string, keys ...string) (bool, bool) {
	raw := strings.ToLower(strings.TrimSpace(firstEnv(values, keys...)))
	switch raw {
	case "1", "true", "t", "yes", "y", "on":
		return true, true
	case "0", "false", "f", "no", "n", "off":
		return false, true
	default:
		return false, false
	}
}

func splitList(value string) []string {
	value = strings.TrimSpace(value)
	var decoded []string
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		if err := json.Unmarshal([]byte(value), &decoded); err == nil {
			return cleanList(decoded)
		}
		// Python-style environment examples often use single-quoted lists,
		// which are not JSON but should preserve the same element boundaries.
		value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	}
	return cleanList(strings.Split(value, ","))
}

func parseTaskQueues(value string) map[string]int {
	queues := make(map[string]int)
	var decoded map[string]int
	if err := json.Unmarshal([]byte(strings.TrimSpace(value)), &decoded); err == nil {
		for name, priority := range decoded {
			name = strings.TrimSpace(name)
			if name != "" && priority > 0 {
				queues[name] = priority
			}
		}
		return queues
	}
	for _, item := range splitList(value) {
		name, rawPriority, ok := strings.Cut(item, ":")
		if !ok {
			name, rawPriority, ok = strings.Cut(item, "=")
		}
		priority, err := strconv.Atoi(strings.TrimSpace(rawPriority))
		name = strings.TrimSpace(name)
		if ok && err == nil && name != "" && priority > 0 {
			queues[name] = priority
		}
	}
	return queues
}

func cleanList(parts []string) []string {
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(strings.TrimSpace(part), `"'`)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}
