package middleware

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"
	"github.com/yuWorm/fba-go/core/config"
	corelogger "github.com/yuWorm/fba-go/core/logger"
	"go.uber.org/zap"
)

type requestLogStateKey struct{}

type serverErrorMarkerKey struct{}

var (
	requestLogStateLocalKey requestLogStateKey
	serverErrorLocalKey     serverErrorMarkerKey
)

// HTTPLogger installs final-response logging on app. It must be registered
// before routes and paired with NewErrorHandler so parser errors are logged
// after Fiber writes their final response.
func HTTPLogger(app *fiber.App, opts config.Options) fiber.Handler {
	opts = opts.WithDefaults()
	logger := newRequestLogger(opts)

	server := app.Server()
	serverErrorHandler := server.ErrorHandler
	server.ErrorHandler = func(fctx *fasthttp.RequestCtx, err error) {
		// Fiber runs use middleware while handling parser failures. Mark the
		// underlying request before that traversal so a short-circuiting
		// middleware cannot emit its provisional response as an access event.
		fctx.SetUserValue(serverErrorLocalKey, true)
		serverErrorHandler(fctx, err)
	}

	return func(c fiber.Ctx) error {
		start := time.Now()
		var state *requestLogState
		if isServerErrorRequest(c) {
			state = beginRequestLog(c, logger, start)
		}

		err := c.Next()
		if err != nil {
			if state == nil {
				beginRequestLog(c, logger, start)
			}
			return err
		}
		if state != nil {
			return nil
		}

		logger.log(c, responseStatus(c), start, nil)
		return nil
	}
}

type requestLogger struct {
	accessEnabled bool
	errorEnabled  bool
	accessLogger  *zap.Logger
	errorLogger   *zap.Logger
	skipAccessLog map[string]bool
}

type requestLogState struct {
	logger *requestLogger
	start  time.Time
	logged bool
}

func newRequestLogger(opts config.Options) *requestLogger {
	logger := &requestLogger{
		accessEnabled: opts.Middleware.AccessLog.Enabled,
		errorEnabled:  opts.Middleware.ErrorLog.Enabled,
	}
	if logger.accessEnabled {
		logger.skipAccessLog = skipPathSet(opts.Middleware.AccessLog.SkipPaths)
		logger.accessLogger = newHTTPLogger(opts.Logger, accessLogPaths(opts.Logger))
	}
	if logger.errorEnabled {
		logger.errorLogger = newHTTPLogger(opts.Logger, errorLogPaths(opts.Logger))
	}
	return logger
}

func isServerErrorRequest(c fiber.Ctx) bool {
	return c.RequestCtx().UserValue(serverErrorLocalKey) != nil
}

func beginRequestLog(c fiber.Ctx, logger *requestLogger, start time.Time) *requestLogState {
	state := &requestLogState{logger: logger, start: start}
	c.Locals(requestLogStateLocalKey, state)
	return state
}

func finishRequestLog(c fiber.Ctx, err error) {
	state, ok := c.Locals(requestLogStateLocalKey).(*requestLogState)
	if !ok || state.logged {
		return
	}
	state.logged = true
	state.logger.log(c, responseStatus(c), state.start, err)
}

func (l *requestLogger) log(c fiber.Ctx, status int, start time.Time, err error) {
	if l.accessEnabled && !l.skipAccessLog[c.Path()] {
		l.accessLogger.Info("http request", requestLogFields(c, status, start)...)
		_ = l.accessLogger.Sync()
	}
	if l.errorEnabled && err != nil && status >= fiber.StatusInternalServerError {
		fields := requestLogFields(c, status, start)
		fields = append(fields, zap.Error(err))
		if value := localString(c, PanicLocalKey); value != "" {
			fields = append(fields, zap.String("panic", value))
		}
		if value := localString(c, PanicStackLocalKey); value != "" {
			fields = append(fields, zap.String("stack", value))
		}
		l.errorLogger.Error("http error", fields...)
		_ = l.errorLogger.Sync()
	}
}

func responseStatus(c fiber.Ctx) int {
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
		// LOG_ACCESS_FILENAME means "also write an access log file". Keep
		// stdout so local development and process managers still see requests.
		return appendLogPath([]string{opts.AccessLogPath}, "stdout")
	}
	if len(opts.OutputPaths) > 0 {
		return opts.OutputPaths
	}
	return []string{"stdout"}
}

func errorLogPaths(opts config.LoggerOptions) []string {
	if opts.ErrorLogPath != "" {
		// LOG_ERROR_FILENAME should not hide server errors from stderr; file
		// output is an additional sink for later inspection.
		return appendLogPath([]string{opts.ErrorLogPath}, "stderr")
	}
	if len(opts.ErrorOutputPaths) > 0 {
		return opts.ErrorOutputPaths
	}
	return []string{"stderr"}
}

func appendLogPath(paths []string, path string) []string {
	for _, existing := range paths {
		if existing == path {
			return paths
		}
	}
	return append(paths, path)
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
