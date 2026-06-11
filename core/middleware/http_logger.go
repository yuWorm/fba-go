package middleware

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	corelogger "github.com/yuWorm/fba-go/core/logger"
	"go.uber.org/zap"
)

func HTTPLogger(opts config.Options) fiber.Handler {
	opts = opts.WithDefaults()
	accessLogger := newHTTPLogger(opts.Logger, accessLogPaths(opts.Logger))
	errorLogger := newHTTPLogger(opts.Logger, errorLogPaths(opts.Logger))
	skipAccessLog := skipPathSet(opts.Middleware.AccessLog.SkipPaths)

	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := responseStatus(c, err)

		if opts.Middleware.AccessLog.Enabled && !skipAccessLog[c.Path()] {
			accessLogger.Info("http request", requestLogFields(c, status, start)...)
			_ = accessLogger.Sync()
		}
		if opts.Middleware.ErrorLog.Enabled && err != nil && status >= fiber.StatusInternalServerError {
			fields := requestLogFields(c, status, start)
			fields = append(fields, zap.Error(err))
			if value := localString(c, PanicLocalKey); value != "" {
				fields = append(fields, zap.String("panic", value))
			}
			if value := localString(c, PanicStackLocalKey); value != "" {
				fields = append(fields, zap.String("stack", value))
			}
			errorLogger.Error("http error", fields...)
			_ = errorLogger.Sync()
		}

		return err
	}
}

func responseStatus(c fiber.Ctx, err error) int {
	if err != nil {
		return mapError(err).status
	}
	status := c.Response().StatusCode()
	if status == 0 {
		return fiber.StatusOK
	}
	return status
}

func requestLogFields(c fiber.Ctx, status int, start time.Time) []zap.Field {
	return []zap.Field{
		zap.String("trace_id", RequestIDFromCtx(c)),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("route", c.FullPath()),
		zap.Int("status", status),
		zap.Duration("latency", time.Since(start)),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get(fiber.HeaderUserAgent)),
	}
}

func newHTTPLogger(opts config.LoggerOptions, paths []string) *zap.Logger {
	ensureLogDirs(paths)
	log, err := corelogger.New(config.LoggerOptions{
		Level:       opts.Level,
		Encoding:    opts.Encoding,
		OutputPaths: paths,
	})
	if err == nil {
		return log
	}
	fallback, fallbackErr := zap.NewProduction()
	if fallbackErr != nil {
		return zap.NewNop()
	}
	return fallback
}

func accessLogPaths(opts config.LoggerOptions) []string {
	if opts.AccessLogPath != "" {
		return []string{opts.AccessLogPath}
	}
	if len(opts.OutputPaths) > 0 {
		return opts.OutputPaths
	}
	return []string{"stdout"}
}

func errorLogPaths(opts config.LoggerOptions) []string {
	if opts.ErrorLogPath != "" {
		return []string{opts.ErrorLogPath}
	}
	if len(opts.ErrorOutputPaths) > 0 {
		return opts.ErrorOutputPaths
	}
	return []string{"stderr"}
}

func ensureLogDirs(paths []string) {
	for _, path := range paths {
		if path == "" || path == "stdout" || path == "stderr" {
			continue
		}
		dir := filepath.Dir(path)
		if dir == "." || dir == "" {
			continue
		}
		// Log paths often come from .env. Create parent directories at startup so
		// diagnostics do not silently disappear because the log directory is missing.
		_ = os.MkdirAll(dir, 0o755)
	}
}

func skipPathSet(paths []string) map[string]bool {
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		if path != "" {
			set[path] = true
		}
	}
	return set
}

func localString(c fiber.Ctx, key string) string {
	value, ok := c.Locals(key).(string)
	if !ok {
		return ""
	}
	return value
}
