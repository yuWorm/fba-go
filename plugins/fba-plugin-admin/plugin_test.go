package admin_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/plugin"
	admin "github.com/yuWorm/fba-plugin-admin"
)

func TestAdminPluginRegistersPriorityEndpoints(t *testing.T) {
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{
		Container: di.New(),
		Router:    app,
		APIGroup:  app.Group("/api/v1"),
	})

	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	registerRoutes(ctx.APIGroup(), ctx.Routes())

	for _, tc := range []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/auth/captcha"},
		{"POST", "/api/v1/auth/login"},
		{"POST", "/api/v1/auth/refresh"},
		{"POST", "/api/v1/auth/logout"},
		{"GET", "/api/v1/auth/codes"},
		{"GET", "/api/v1/sys/users/me"},
		{"GET", "/api/v1/sys/menus/sidebar"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", tc.method, tc.path, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != fiber.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("%s %s status = %d body = %s", tc.method, tc.path, resp.StatusCode, body)
		}
	}
}

func TestLoginSetsRefreshCookie(t *testing.T) {
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{APIGroup: app.Group("/api/v1")})
	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	registerRoutes(ctx.APIGroup(), ctx.Routes())

	resp, err := app.Test(httptest.NewRequest("POST", "/api/v1/auth/login", nil))
	if err != nil {
		t.Fatalf("POST /auth/login error = %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Set-Cookie"); got == "" {
		t.Fatal("Set-Cookie header is empty")
	}
}

func registerRoutes(router fiber.Router, routes []plugin.Route) {
	for _, route := range routes {
		router.Add([]string{route.Method}, route.Path, route.Handler)
	}
}
