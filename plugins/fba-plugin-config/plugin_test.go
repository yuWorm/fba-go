package config_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	configplugin "github.com/yuWorm/fba-plugin-config"
	"github.com/yuWorm/fba-plugin-config/repo"
	"gorm.io/gorm"
)

func TestConfigPluginRegistersPythonCompatibleRoutes(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := configplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	got := make(map[string]plugin.Route)
	for _, route := range ctx.Routes() {
		got[route.Method+" "+route.Path] = route
	}
	want := map[string]struct {
		authRequired bool
		permission   string
	}{
		"GET /sys/configs/all": {authRequired: true},
		"GET /sys/configs/:pk": {authRequired: true},
		"GET /sys/configs":     {authRequired: true},
		"POST /sys/configs":    {authRequired: true, permission: "sys:config:add"},
		"PUT /sys/configs":     {authRequired: true, permission: "sys.config.edits"},
		"PUT /sys/configs/:pk": {authRequired: true, permission: "sys:config:edit"},
		"DELETE /sys/configs":  {authRequired: true, permission: "sys:config:del"},
	}
	if len(got) != len(want) {
		t.Fatalf("registered route count = %d, want %d; routes = %v", len(got), len(want), routeKeys(got))
	}
	for key, expected := range want {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, routeKeys(got))
		}
		if route.AuthRequired != expected.authRequired {
			t.Fatalf("%s AuthRequired = %v, want %v", key, route.AuthRequired, expected.authRequired)
		}
		if route.Permission != expected.permission {
			t.Fatalf("%s Permission = %q, want %q", key, route.Permission, expected.permission)
		}
	}
}

func TestConfigPluginRegistersMigrationAndAdminProvider(t *testing.T) {
	container := di.New()
	if err := container.Provide(func() db.Provider {
		return db.NewGORMProvider(&gorm.DB{}, nil)
	}); err != nil {
		t.Fatalf("Provide(db.Provider) error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})

	if err := configplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if len(ctx.Migrations()) != 1 {
		t.Fatalf("migrations = %d, want 1", len(ctx.Migrations()))
	}
	if ctx.Migrations()[0].Scope != "plugin:config" {
		t.Fatalf("migration scope = %q, want plugin:config", ctx.Migrations()[0].Scope)
	}
	var provider adminservice.AdminConfigProvider
	if !ctx.Container().Resolve(&provider) {
		t.Fatal("AdminConfigProvider was not registered")
	}
}

func TestConfigReadWriteEndpointsMatchPython(t *testing.T) {
	app := newConfigApp(t, nil)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/configs/all?type=LOGIN", "")
	assertStatusOK(t, resp)
	all := assertEnvelopeSlice(t, body)
	if len(all) != 2 {
		t.Fatalf("LOGIN config count = %d, want 2", len(all))
	}
	assertConfigDetail(t, assertMap(t, all[0]))

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/configs/17", "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	assertConfigDetail(t, detail)
	if detail["key"] != "LOGIN_CAPTCHA_ENABLED" {
		t.Fatalf("config key = %v, want LOGIN_CAPTCHA_ENABLED", detail["key"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/configs?name=验证码&type=LOGIN", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items := assertSlice(t, page["items"])
	if len(items) != 1 {
		t.Fatalf("filtered configs = %d, want 1", len(items))
	}

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/configs", `{"name":"测试配置","type":"LOGIN","key":"LOGIN_TEST_KEY","value":"true","is_frontend":false,"remark":null}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/configs", `{"name":"重复","type":"LOGIN","key":"LOGIN_TEST_KEY","value":"true","is_frontend":false,"remark":null}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "参数配置 LOGIN_TEST_KEY 已存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/configs", `[{"id":16,"name":"登录配置状态","type":"LOGIN","key":"LOGIN_DUPLICATE","value":"1","is_frontend":false,"remark":null},{"id":17,"name":"验证码开关","type":"LOGIN","key":"LOGIN_DUPLICATE","value":"true","is_frontend":false,"remark":null}]`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "参数配置键名重复")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/configs", `[{"id":17,"name":"验证码开关","type":"LOGIN","key":"LOGIN_CAPTCHA_ENABLED","value":"false","is_frontend":false,"remark":null}]`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/configs", `[17]`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/configs/17", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "参数配置不存在")
}

func TestConfigDeleteNoRowsReturnsPythonBusinessFail(t *testing.T) {
	app := newConfigApp(t, nil)

	resp, body := requestJSON(t, app, "DELETE", "/api/v1/sys/configs", `[999999]`)
	assertStatusOK(t, resp)
	assertBusinessFailEnvelope(t, body)
}

func TestConfigAdminProviderFeedsDynamicPolicy(t *testing.T) {
	custom := repo.SeedData()
	for i := range custom.Configs {
		switch custom.Configs[i].Key {
		case "LOGIN_CAPTCHA_ENABLED":
			custom.Configs[i].Value = "false"
		case "USER_PASSWORD_REQUIRE_SPECIAL_CHAR":
			custom.Configs[i].Value = "true"
		case "USER_PASSWORD_MAX_LENGTH":
			custom.Configs[i].Value = "18"
		}
	}
	container := di.New()
	repository := repo.NewMemoryRepository(custom)
	if err := container.Provide(func() repo.Repository {
		return repository
	}); err != nil {
		t.Fatalf("Provide(repo.Repository) error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})
	if err := configplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	var provider adminservice.AdminConfigProvider
	if !ctx.Container().Resolve(&provider) {
		t.Fatal("AdminConfigProvider was not registered")
	}

	login, err := provider.LoginConfig(context.Background())
	if err != nil {
		t.Fatalf("LoginConfig() error = %v", err)
	}
	if login.CaptchaEnabled {
		t.Fatal("LoginConfig().CaptchaEnabled = true, want false from config")
	}
	security, err := provider.UserSecurityConfig(context.Background())
	if err != nil {
		t.Fatalf("UserSecurityConfig() error = %v", err)
	}
	if !security.RequireSpecialChar || security.MaxLength != 18 {
		t.Fatalf("UserSecurityConfig() = %+v, want special char true and max length 18", security)
	}
}

func newConfigApp(t *testing.T, repository repo.Repository) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	container := di.New()
	if repository != nil {
		if err := container.Provide(func() repo.Repository {
			return repository
		}); err != nil {
			t.Fatalf("Provide(repo.Repository) error = %v", err)
		}
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container, APIGroup: app.Group("/api/v1")})
	if err := configplugin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	for _, route := range ctx.Routes() {
		ctx.APIGroup().Add([]string{route.Method}, route.Path, route.Handler)
	}
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
	if body["code"] != float64(200) || body["msg"] != "请求成功" {
		t.Fatalf("envelope = %v, want success map", body)
	}
	return assertMap(t, body["data"])
}

func assertEnvelopeSlice(t *testing.T, body map[string]any) []any {
	t.Helper()
	if body["code"] != float64(200) || body["msg"] != "请求成功" {
		t.Fatalf("envelope = %v, want success slice", body)
	}
	return assertSlice(t, body["data"])
}

func assertEnvelopeNil(t *testing.T, body map[string]any) {
	t.Helper()
	if body["code"] != float64(200) || body["msg"] != "请求成功" || body["data"] != nil {
		t.Fatalf("envelope = %v, want success null", body)
	}
}

func assertBusinessFailEnvelope(t *testing.T, body map[string]any) {
	t.Helper()
	if body["code"] != float64(400) || body["msg"] != "请求错误" || body["data"] != nil {
		t.Fatalf("envelope = %v, want business fail null", body)
	}
}

func assertErrorEnvelope(t *testing.T, resp *http.Response, body map[string]any, status int, msg string) {
	t.Helper()
	if resp.StatusCode != status {
		t.Fatalf("status = %d, want %d; body = %v", resp.StatusCode, status, body)
	}
	if body["code"] != float64(status) || body["msg"] != msg {
		t.Fatalf("error envelope = %v, want code %d msg %s", body, status, msg)
	}
}

func assertConfigDetail(t *testing.T, data map[string]any) {
	t.Helper()
	for _, key := range []string{"id", "name", "type", "key", "value", "is_frontend", "remark", "created_time", "updated_time"} {
		if _, ok := data[key]; !ok {
			t.Fatalf("config key %q missing from %v", key, data)
		}
	}
}

func assertMap(t *testing.T, value any) map[string]any {
	t.Helper()
	got, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %T, want map", value)
	}
	return got
}

func assertSlice(t *testing.T, value any) []any {
	t.Helper()
	got, ok := value.([]any)
	if !ok {
		t.Fatalf("value = %T, want slice", value)
	}
	return got
}

func routeKeys(routes map[string]plugin.Route) []string {
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	return keys
}
