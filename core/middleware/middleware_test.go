package middleware_test

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/middleware"
)

func TestRequestIDPreservesIncomingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/id", func(c fiber.Ctx) error {
		return c.SendString(c.Locals(middleware.RequestIDLocalKey).(string))
	})

	req := httptest.NewRequest("GET", "/id", nil)
	req.Header.Set("X-Request-ID", "request-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "request-1" {
		t.Fatalf("request id body = %q, want request-1", body)
	}
	if resp.Header.Get("X-Request-ID") != "request-1" {
		t.Fatalf("X-Request-ID response header = %q", resp.Header.Get("X-Request-ID"))
	}
}

func TestRequestIDGeneratesMissingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	app.Get("/id", func(c fiber.Ctx) error {
		return c.SendString(c.Locals(middleware.RequestIDLocalKey).(string))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/id", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Request-ID") == "" {
		t.Fatal("X-Request-ID response header is empty")
	}
}

func TestRecoverReturnsCompatibleErrorEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	app.Use(middleware.Recover())
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("boom")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"code":500`) {
		t.Fatalf("body = %s, want code 500", body)
	}
	if !strings.Contains(string(body), `"trace_id"`) {
		t.Fatalf("body = %s, want trace_id", body)
	}
}

func TestErrorHandlerNeverExposesInternalDetailInProduction(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: middleware.NewErrorHandler(config.Options{
		App: config.AppOptions{Environment: "prod"},
		Middleware: config.MiddlewareOptions{
			ErrorResponse: config.ErrorResponseOptions{IncludeDetail: true},
		},
	})})
	app.Get("/failure", func(fiber.Ctx) error {
		return errors.New("postgres://user:secret@database.internal/private")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/failure", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if strings.Contains(string(body), "secret") || !strings.Contains(string(body), "内部服务器错误") {
		t.Fatalf("body = %s, want public production error", body)
	}
}
