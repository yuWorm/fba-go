package admin_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
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
		{"POST", "/api/v1/auth/login/swagger"},
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

func TestAdminPluginRegistersPythonCompatibleRouteMetadata(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got := make(map[string]plugin.Route)
	for _, route := range ctx.Routes() {
		got[route.Method+" "+route.Path] = route
	}

	want := map[string]bool{
		"GET /auth/captcha":        false,
		"POST /auth/login/swagger": false,
		"POST /auth/login":         false,
		"POST /auth/refresh":       false,
		"POST /auth/logout":        false,
		"GET /auth/codes":          true,
		"GET /sys/users/me":        true,
		"GET /sys/menus/sidebar":   true,
	}
	for key, authRequired := range want {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, maps.Keys(got))
		}
		if route.AuthRequired != authRequired {
			t.Fatalf("%s AuthRequired = %v, want %v", key, route.AuthRequired, authRequired)
		}
	}
}

func TestCaptchaMatchesPythonSchema(t *testing.T) {
	app := newAdminApp(t)
	resp, body := requestJSON(t, app, "GET", "/api/v1/auth/captcha", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	assertKeys(t, data, "is_enabled", "expire_seconds", "uuid", "image")
	if data["is_enabled"] != true {
		t.Fatalf("captcha is_enabled = %v, want true", data["is_enabled"])
	}
	if data["expire_seconds"] != float64(300) {
		t.Fatalf("captcha expire_seconds = %v, want 300", data["expire_seconds"])
	}
	if data["uuid"] == "" {
		t.Fatal("captcha uuid is empty")
	}
}

func TestLoginSwaggerMatchesPythonSchema(t *testing.T) {
	app := newAdminApp(t)
	resp, body := requestJSON(t, app, "POST", "/api/v1/auth/login/swagger", "")

	assertStatusOK(t, resp)
	if _, ok := body["code"]; ok {
		t.Fatalf("login/swagger response has envelope code: %v", body)
	}
	assertKeys(t, body, "access_token", "token_type", "user")
	if body["token_type"] != "Bearer" {
		t.Fatalf("token_type = %v, want Bearer", body["token_type"])
	}
	user := assertMap(t, body["user"])
	assertUserInfoDetail(t, user)
}

func TestLoginMatchesPythonSchemaAndSetsRefreshCookie(t *testing.T) {
	app := newAdminApp(t)
	resp, body := requestJSON(t, app, "POST", "/api/v1/auth/login", `{"username":"admin","password":"admin","uuid":"fixture-captcha","captcha":"1234"}`)

	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	assertKeys(t, data, "access_token", "access_token_expire_time", "session_uuid", "password_expire_days_remaining", "user")
	if _, ok := data["password_expire_days_remaining"]; !ok {
		t.Fatal("password_expire_days_remaining key missing")
	}
	user := assertMap(t, data["user"])
	assertUserInfoDetail(t, user)
	assertRefreshCookie(t, resp.Header.Get("Set-Cookie"))
}

func TestRefreshMatchesPythonSchemaAndSetsRefreshCookie(t *testing.T) {
	app := newAdminApp(t)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "fba_refresh_token", Value: "fixture-refresh-token"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /auth/refresh error = %v", err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode refresh body: %v", err)
	}

	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	assertKeys(t, data, "access_token", "access_token_expire_time", "session_uuid")
	assertRefreshCookie(t, resp.Header.Get("Set-Cookie"))
}

func TestCurrentUserMatchesPythonSchema(t *testing.T) {
	app := newAdminApp(t)
	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/users/me", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	assertUserInfoDetail(t, data)
	assertKeys(t, data, "dept", "roles")
	if _, ok := data["menus"]; ok {
		t.Fatal("current user contains menus, not present in Python schema")
	}
	if _, ok := data["depts"]; ok {
		t.Fatal("current user contains depts, not present in Python schema")
	}
	if _, ok := data["roles"].([]any); !ok {
		t.Fatalf("roles = %T, want JSON array", data["roles"])
	}
}

func TestSidebarMenusMatchesPythonVben5Schema(t *testing.T) {
	app := newAdminApp(t)
	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/menus/sidebar", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeSlice(t, body)
	if len(data) == 0 {
		t.Fatal("sidebar menu data is empty")
	}
	menu := assertMap(t, data[0])
	assertKeys(t, menu, "id", "name", "path", "parent_id", "sort", "type", "component", "perms", "remark", "children", "meta")
	meta := assertMap(t, menu["meta"])
	assertKeys(t, meta, "title", "icon", "iframeSrc", "link", "keepAlive", "hideInMenu", "menuVisibleWithForbidden")
}

func newAdminApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{APIGroup: app.Group("/api/v1")})
	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	registerRoutes(ctx.APIGroup(), ctx.Routes())
	return app
}

func requestJSON(t *testing.T, app *fiber.App, method string, path string, body string) (*http.Response, map[string]any) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	defer resp.Body.Close()
	var decoded map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode %s %s response: %v", method, path, err)
	}
	return resp, decoded
}

func assertStatusOK(t *testing.T, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func assertEnvelopeMap(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	if body["code"] != float64(200) {
		t.Fatalf("code = %v, want 200; body = %v", body["code"], body)
	}
	if body["msg"] != "请求成功" {
		t.Fatalf("msg = %v, want 请求成功", body["msg"])
	}
	return assertMap(t, body["data"])
}

func assertEnvelopeSlice(t *testing.T, body map[string]any) []any {
	t.Helper()
	if body["code"] != float64(200) {
		t.Fatalf("code = %v, want 200; body = %v", body["code"], body)
	}
	if body["msg"] != "请求成功" {
		t.Fatalf("msg = %v, want 请求成功", body["msg"])
	}
	data, ok := body["data"].([]any)
	if !ok {
		t.Fatalf("data = %T, want JSON array", body["data"])
	}
	return data
}

func assertMap(t *testing.T, value any) map[string]any {
	t.Helper()
	got, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %T, want JSON object", value)
	}
	return got
}

func assertKeys(t *testing.T, value map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if _, ok := value[key]; !ok {
			t.Fatalf("key %q missing from %v", key, value)
		}
	}
}

func assertUserInfoDetail(t *testing.T, user map[string]any) {
	t.Helper()
	assertKeys(t, user,
		"dept_id",
		"username",
		"nickname",
		"avatar",
		"email",
		"phone",
		"id",
		"uuid",
		"status",
		"is_superuser",
		"is_staff",
		"is_multi_login",
		"join_time",
		"last_login_time",
	)
}

func assertRefreshCookie(t *testing.T, cookie string) {
	t.Helper()
	lower := strings.ToLower(cookie)
	if !strings.Contains(cookie, "fba_refresh_token=") {
		t.Fatalf("Set-Cookie missing fba_refresh_token: %s", cookie)
	}
	if !strings.Contains(lower, "httponly") {
		t.Fatalf("Set-Cookie missing HttpOnly: %s", cookie)
	}
	if !strings.Contains(lower, "max-age=604800") {
		t.Fatalf("Set-Cookie missing Max-Age=604800: %s", cookie)
	}
}

func registerRoutes(router fiber.Router, routes []plugin.Route) {
	for _, route := range routes {
		router.Add([]string{route.Method}, route.Path, route.Handler)
	}
}
