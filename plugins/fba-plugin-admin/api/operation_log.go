package api

import (
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-plugin-admin/model"
)

const (
	operationLogSuccessStatus = 1
	operationLogFailStatus    = 0
	operationLogBodyMaxSize   = 10 * 1024
	operationLogRedactedValue = "******"
)

var operationLogRedactKeys = map[string]struct{}{
	"password":         {},
	"old_password":     {},
	"new_password":     {},
	"confirm_password": {},
}

func WithOperationLogging(handler Handler, apiBasePath string, routes []plugin.Route) []plugin.Route {
	if handler.logs == nil {
		return routes
	}
	if apiBasePath == "" {
		apiBasePath = "/api/v1"
	}
	result := make([]plugin.Route, 0, len(routes))
	for _, route := range routes {
		item := route
		original := item.Handler
		item.Handler = handler.operationLogHandler(apiBasePath, item, original)
		result = append(result, item)
	}
	return result
}

func (h Handler) operationLogHandler(apiBasePath string, route plugin.Route, next fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		snapshot := operationLogSnapshotFromCtx(c)
		err := next(c)
		if shouldRecordOperationLog(apiBasePath, snapshot, route) {
			h.recordOperationLog(c, route, snapshot, err)
		}
		return err
	}
}

type operationLogSnapshot struct {
	Method    string
	Path      string
	IP        string
	UserAgent *string
	Args      map[string]any
	Started   time.Time
}

func operationLogSnapshotFromCtx(c fiber.Ctx) operationLogSnapshot {
	return operationLogSnapshot{
		Method:    stableString(c.Method()),
		Path:      stableString(c.Path()),
		IP:        stableString(c.IP()),
		UserAgent: optionalString(stableString(c.Get("User-Agent"))),
		Args:      operationLogArgs(c),
		Started:   time.Now(),
	}
}

func shouldRecordOperationLog(apiBasePath string, snapshot operationLogSnapshot, route plugin.Route) bool {
	if strings.EqualFold(snapshot.Method, fiber.MethodOptions) {
		return false
	}
	if !strings.HasPrefix(snapshot.Path, apiBasePath) {
		return false
	}
	switch snapshot.Path {
	case "/favicon.ico", "/docs", "/redoc", "/openapi", apiBasePath + "/auth/login/swagger":
		return false
	default:
		return route.Handler != nil
	}
}

func (h Handler) recordOperationLog(c fiber.Ctx, route plugin.Route, snapshot operationLogSnapshot, err error) {
	statusCode := c.Response().StatusCode()
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	status := operationLogSuccessStatus
	message := "请求成功"
	if err != nil {
		status = operationLogFailStatus
		message = err.Error()
		var appErr *fbaerrors.AppError
		if stderrors.As(err, &appErr) {
			statusCode = appErr.Code()
			message = appErr.PublicMessage()
		}
	} else if statusCode >= http.StatusBadRequest {
		status = operationLogFailStatus
		message = http.StatusText(statusCode)
	}
	username := currentUsername(c)
	ip := snapshot.IP
	if ip == "" {
		ip = "127.0.0.1"
	}
	msg := message
	// Python writes operation logs through a background queue. Keep persistence
	// best-effort so a logging outage cannot change the business response.
	_ = h.logs.CreateOpera(c.RequestCtx(), model.OperaLog{
		TraceID:     middleware.RequestIDFromCtx(c),
		Username:    username,
		Method:      snapshot.Method,
		Title:       route.Summary,
		Path:        snapshot.Path,
		IP:          ip,
		UserAgent:   snapshot.UserAgent,
		Args:        snapshot.Args,
		Status:      status,
		Code:        strconv.Itoa(statusCode),
		Msg:         &msg,
		CostTime:    float64(time.Since(snapshot.Started).Microseconds()) / 1000,
		OperaTime:   snapshot.Started,
		CreatedTime: time.Now(),
	})
}

func currentUsername(c fiber.Ctx) *string {
	user := currentUser(c)
	if user == nil || user.Username == "" {
		return nil
	}
	return &user.Username
}

func operationLogArgs(c fiber.Ctx) map[string]any {
	args := map[string]any{}
	queryArgs := map[string]any{}
	c.Request().URI().QueryArgs().VisitAll(func(key []byte, value []byte) {
		queryArgs[string(key)] = string(value)
	})
	if len(queryArgs) > 0 {
		args["query_params"] = redactAnyMap(queryArgs)
	}
	body := c.Body()
	if len(body) > 0 {
		if len(body) > operationLogBodyMaxSize {
			args["body"] = map[string]any{
				"is_truncated": true,
				"max_size":     operationLogBodyMaxSize,
				"actual_size":  len(body),
			}
		} else if decoded := decodeOperationLogBody(body); decoded != nil {
			args["body"] = redactAnyMap(decoded)
		}
	}
	if len(args) == 0 {
		return nil
	}
	return args
}

func stableString(value string) string {
	if value == "" {
		return ""
	}
	return string([]byte(value))
}

func decodeOperationLogBody(body []byte) map[string]any {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil
	}
	return decoded
}

func redactAnyMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		if _, ok := operationLogRedactKeys[strings.ToLower(key)]; ok {
			result[key] = operationLogRedactedValue
			continue
		}
		result[key] = redactAny(value)
	}
	return result
}

func redactAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactAnyMap(typed)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, redactAny(item))
		}
		return result
	default:
		return typed
	}
}
