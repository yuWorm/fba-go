package fiberx_test

import (
	"bufio"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestNewLogsFinalWireStatusesOverTCP(t *testing.T) {
	const (
		bodyLimit          = 16 * 1024 * 1024
		oversizedRequestID = "oversized-request"
		preflightRequestID = "preflight-request"
		methodRequestID    = "method-request"
		successRequestID   = "success-request"
	)

	dir := t.TempDir()
	accessLogPath := filepath.Join(dir, "access.log")
	fx := fiberx.New(config.Options{
		App: config.AppOptions{Environment: "prod"},
		Fiber: fiber.Config{
			BodyLimit: bodyLimit,
		},
		Logger: config.LoggerOptions{
			AccessLogPath: accessLogPath,
		},
	})
	fx.API.Post("/upload", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
	fx.API.Get("/ready", func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	addr := startFiberApp(t, fx.App)
	wireStatuses := []struct {
		requestID string
		status    int
	}{
		{
			requestID: oversizedRequestID,
			status: tcpStatus(t, addr, http.MethodPost, fmt.Sprintf(
				"POST /api/v1/upload HTTP/1.1\r\nHost: %s\r\nX-Request-ID: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
				addr,
				oversizedRequestID,
				bodyLimit+1,
			)),
		},
		{
			requestID: preflightRequestID,
			status: tcpStatus(t, addr, http.MethodOptions, fmt.Sprintf(
				"OPTIONS /api/v1/upload HTTP/1.1\r\nHost: %s\r\nOrigin: http://localhost:5173\r\nAccess-Control-Request-Method: POST\r\nX-Request-ID: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n",
				addr,
				preflightRequestID,
				bodyLimit+1,
			)),
		},
		{
			requestID: methodRequestID,
			status: tcpStatus(t, addr, http.MethodGet, fmt.Sprintf(
				"GET /api/v1/upload HTTP/1.1\r\nHost: %s\r\nX-Request-ID: %s\r\nConnection: close\r\n\r\n",
				addr,
				methodRequestID,
			)),
		},
		{
			requestID: successRequestID,
			status: tcpStatus(t, addr, http.MethodGet, fmt.Sprintf(
				"GET /api/v1/ready HTTP/1.1\r\nHost: %s\r\nX-Request-ID: %s\r\nConnection: close\r\n\r\n",
				addr,
				successRequestID,
			)),
		},
	}
	for index, want := range []struct {
		name   string
		status int
	}{
		{"oversized request", fiber.StatusRequestEntityTooLarge},
		{"preflight request", fiber.StatusRequestEntityTooLarge},
		{"method request", http.StatusMethodNotAllowed},
		{"success request", http.StatusNoContent},
	} {
		if got := wireStatuses[index].status; got != want.status {
			t.Fatalf("%s status = %d, want %d", want.name, got, want.status)
		}
	}

	entries := readAccessLogEntries(t, accessLogPath)
	if len(entries) != len(wireStatuses) {
		t.Fatalf("access entries = %d, want exactly %d: %#v", len(entries), len(wireStatuses), entries)
	}
	for _, wire := range wireStatuses {
		var matches int
		for _, entry := range entries {
			if entry.TraceID != wire.requestID {
				continue
			}
			matches++
			if entry.Status != wire.status {
				t.Fatalf("access status for request %q = %d, want wire status %d", wire.requestID, entry.Status, wire.status)
			}
		}
		if matches != 1 {
			t.Fatalf("access entries for request %q = %d, want 1: %#v", wire.requestID, matches, entries)
		}
	}
}

func TestNewPreservesSkipPathsForFinalErrors(t *testing.T) {
	dir := t.TempDir()
	accessLogPath := filepath.Join(dir, "access.log")
	errorLogPath := filepath.Join(dir, "error.log")
	fx := fiberx.New(config.Options{
		App: config.AppOptions{Environment: "prod"},
		Logger: config.LoggerOptions{
			AccessLogPath: accessLogPath,
			ErrorLogPath:  errorLogPath,
		},
		Middleware: config.MiddlewareOptions{
			AccessLog: config.AccessLogOptions{
				SkipPaths: []string{"/api/v1/quiet"},
			},
		},
	})
	fx.API.Get("/quiet", func(fiber.Ctx) error {
		return stderrors.New("quiet storage failure")
	})

	resp, err := fx.App.Test(httptest.NewRequest(http.MethodGet, "/api/v1/quiet", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusInternalServerError)
	}
	if entries := readAccessLogEntries(t, accessLogPath); len(entries) != 0 {
		t.Fatalf("access entries = %#v, want none for skipped path", entries)
	}

	errorLog, err := os.ReadFile(errorLogPath)
	if err != nil {
		t.Fatalf("read error log: %v", err)
	}
	if !strings.Contains(string(errorLog), "quiet storage failure") {
		t.Fatalf("error log = %s, missing request error", errorLog)
	}
}

type accessLogEntry struct {
	Message string `json:"msg"`
	TraceID string `json:"trace_id"`
	Status  int    `json:"status"`
}

func readAccessLogEntries(t *testing.T, path string) []accessLogEntry {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read access log: %v", err)
	}
	var entries []accessLogEntry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var entry accessLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode access log entry %q: %v", line, err)
		}
		if entry.Message == "http request" {
			entries = append(entries, entry)
		}
	}
	return entries
}

func startFiberApp(t *testing.T, app *fiber.App) string {
	t.Helper()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ready := make(chan struct{})
	serverDone := make(chan error, 1)
	running := true
	t.Cleanup(func() {
		if !running {
			return
		}
		if err := app.ShutdownWithTimeout(time.Second); err != nil {
			t.Errorf("shutdown: %v", err)
		}
		if err := <-serverDone; err != nil && !stderrors.Is(err, net.ErrClosed) {
			t.Errorf("serve: %v", err)
		}
	})
	go func() {
		serverDone <- app.Listener(listener, fiber.ListenConfig{
			DisableStartupMessage: true,
			BeforeServeFunc: func(*fiber.App) error {
				close(ready)
				return nil
			},
		})
	}()

	select {
	case <-ready:
		return listener.Addr().String()
	case err := <-serverDone:
		running = false
		t.Fatalf("serve: %v", err)
	case <-time.After(time.Second):
		t.Fatal("server did not start")
	}
	return ""
}

func tcpStatus(t *testing.T, addr, method, request string) int {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatalf("write request: %v", err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: method})
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("read response body: %v", err)
	}
	return resp.StatusCode
}
