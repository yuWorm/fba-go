package fiberx_test

import (
	"io"
	"net/http"
	"net/http/httptest"
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
