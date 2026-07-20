package fiberx_test

import (
	stderrors "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/fiberx"
)

func TestNewCreatesAPIGroupWithDefaultBasePath(t *testing.T) {
	fx := fiberx.New(config.Options{})
	fx.API.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	resp, err := fx.App.Test(httptest.NewRequest("GET", "/api/v1/ping", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "pong" {
		t.Fatalf("body = %q, want pong", body)
	}
}

func TestRouteKeepsCompatibilityMetadata(t *testing.T) {
	route := fiberx.Route{
		Method:       "DELETE",
		Path:         "/sys/users/:pk",
		Permission:   "sys:user:del",
		AuthRequired: true,
		Summary:      "删除用户",
		Tags:         []string{"系统用户"},
	}

	if route.Permission != "sys:user:del" {
		t.Fatalf("Permission = %q", route.Permission)
	}
	if !route.AuthRequired {
		t.Fatal("AuthRequired = false, want true")
	}
}

func TestNewAddsPythonCompatibleCORSMiddleware(t *testing.T) {
	fx := fiberx.New(config.Options{})
	fx.API.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := fx.App.Test(req)
	if err != nil {
		t.Fatalf("App.Test(preflight) error = %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("Allow-Origin = %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("Allow-Credentials = %q", resp.Header.Get("Access-Control-Allow-Credentials"))
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp, err = fx.App.Test(req)
	if err != nil {
		t.Fatalf("App.Test(simple) error = %v", err)
	}
	defer resp.Body.Close()
	if resp.Header.Get("Access-Control-Expose-Headers") != "X-Request-ID" {
		t.Fatalf("Expose-Headers = %q", resp.Header.Get("Access-Control-Expose-Headers"))
	}
}

func TestNewSkipsCORSMiddlewareWhenDisabled(t *testing.T) {
	fx := fiberx.New(config.Options{
		CORS: config.CORSOptions{Disabled: true},
	})
	fx.API.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	resp, err := fx.App.Test(req)
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, want empty", got)
	}
}

func TestNewInstallsRequestIDAndRecoverByDefault(t *testing.T) {
	fx := fiberx.New(config.Options{})
	fx.API.Get("/panic", func(c fiber.Ctx) error {
		panic("boom")
	})

	resp, err := fx.App.Test(httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	traceID := resp.Header.Get("X-Request-ID")
	if traceID == "" {
		t.Fatal("X-Request-ID response header is empty")
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"trace_id":"`+traceID+`"`) {
		t.Fatalf("body = %s, want trace_id %q", body, traceID)
	}
	if !strings.Contains(string(body), "panic: boom") {
		t.Fatalf("body = %s, want dev panic detail", body)
	}
}

func TestNewControlsInternalErrorResponseDetails(t *testing.T) {
	fx := fiberx.New(config.Options{
		App: config.AppOptions{Environment: "prod"},
	})
	fx.API.Get("/err", func(c fiber.Ctx) error {
		return stderrors.New("database connection refused")
	})

	resp, err := fx.App.Test(httptest.NewRequest(http.MethodGet, "/api/v1/err", nil))
	if err != nil {
		t.Fatalf("App.Test(prod) error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll(prod) error = %v", err)
	}
	if strings.Contains(string(body), "database connection refused") {
		t.Fatalf("body = %s, should hide internal error detail in prod", body)
	}

	fx = fiberx.New(config.Options{
		App: config.AppOptions{Environment: "prod"},
		Middleware: config.MiddlewareOptions{
			ErrorResponse: config.ErrorResponseOptions{IncludeDetail: true},
		},
	})
	fx.API.Get("/err", func(c fiber.Ctx) error {
		return stderrors.New("database connection refused")
	})

	resp, err = fx.App.Test(httptest.NewRequest(http.MethodGet, "/api/v1/err", nil))
	if err != nil {
		t.Fatalf("App.Test(include detail) error = %v", err)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll(include detail) error = %v", err)
	}
	if strings.Contains(string(body), "database connection refused") {
		t.Fatalf("body = %s, production must ignore attempts to expose internal details", body)
	}
}

func TestNewWritesAccessAndErrorLogs(t *testing.T) {
	dir := t.TempDir()
	accessLogPath := filepath.Join(dir, "access.log")
	errorLogPath := filepath.Join(dir, "error.log")
	fx := fiberx.New(config.Options{
		App: config.AppOptions{Environment: "prod"},
		Logger: config.LoggerOptions{
			AccessLogPath: accessLogPath,
			ErrorLogPath:  errorLogPath,
		},
	})
	fx.API.Get("/err", func(c fiber.Ctx) error {
		return stderrors.New("storage backend unavailable")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/err", nil)
	req.Header.Set("X-Request-ID", "trace-1")
	resp, err := fx.App.Test(req)
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}

	accessLog, err := os.ReadFile(accessLogPath)
	if err != nil {
		t.Fatalf("read access log: %v", err)
	}
	for _, want := range []string{"http request", "GET", "/api/v1/err", "500", "trace-1"} {
		if !strings.Contains(string(accessLog), want) {
			t.Fatalf("access log = %s, missing %q", accessLog, want)
		}
	}

	errorLog, err := os.ReadFile(errorLogPath)
	if err != nil {
		t.Fatalf("read error log: %v", err)
	}
	for _, want := range []string{"http error", "storage backend unavailable", "/api/v1/err", "trace-1"} {
		if !strings.Contains(string(errorLog), want) {
			t.Fatalf("error log = %s, missing %q", errorLog, want)
		}
	}
}
