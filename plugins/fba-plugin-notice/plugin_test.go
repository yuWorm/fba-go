package notice_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/plugin"
	notice "github.com/yuWorm/fba-plugin-notice"
	"gorm.io/gorm"
)

func TestNoticePluginRegistersPythonCompatibleRoutes(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := notice.FBAPlugin().Register(ctx); err != nil {
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
		"GET /sys/notices/:pk": {authRequired: true},
		"GET /sys/notices":     {authRequired: true},
		"POST /sys/notices":    {authRequired: true, permission: "sys:notice:add"},
		"PUT /sys/notices/:pk": {authRequired: true, permission: "sys:notice:edit"},
		"DELETE /sys/notices":  {authRequired: true, permission: "sys:notice:del"},
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

func TestNoticePluginRegistersMigrationWhenDBProviderExists(t *testing.T) {
	container := di.New()
	if err := container.Provide(func() db.Provider {
		return db.NewGORMProvider(&gorm.DB{}, nil)
	}); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})

	if err := notice.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	migrations := ctx.Migrations()
	if len(migrations) != 1 {
		t.Fatalf("migrations = %d, want 1", len(migrations))
	}
	if migrations[0].Scope != "plugin:notice" {
		t.Fatalf("migration scope = %q, want plugin:notice", migrations[0].Scope)
	}
}

func TestNoticeReadEndpointsMatchPythonSchemas(t *testing.T) {
	app := newNoticeApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/notices", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	assertPage(t, page, assertNoticeDetail)
	links := assertMap(t, page["links"])
	if !strings.HasPrefix(links["self"].(string), "/api/v1/sys/notices?") {
		t.Fatalf("self link = %v, want /api/v1/sys/notices? prefix", links["self"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/notices/1", "")
	assertStatusOK(t, resp)
	assertNoticeDetail(t, assertEnvelopeMap(t, body))
}

func TestNoticeWriteEndpointsReturnPythonEnvelope(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/sys/notices", noticeBody()},
		{"PUT", "/api/v1/sys/notices/1", noticeBody()},
		{"DELETE", "/api/v1/sys/notices", `{"pks":[1]}`},
	} {
		app := newNoticeApp(t)
		resp, body := requestJSON(t, app, tc.method, tc.path, tc.body)
		assertStatusOK(t, resp)
		assertEnvelopeNull(t, body)
	}
}

func newNoticeApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{APIGroup: app.Group("/api/v1")})
	if err := notice.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	for _, route := range ctx.Routes() {
		ctx.APIGroup().Add([]string{route.Method}, route.Path, route.Handler)
	}
	return app
}

func noticeBody() string {
	return `{"title":"Contract Notice","type":0,"status":1,"content":"hello notice"}`
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

func assertEnvelopeNull(t *testing.T, body map[string]any) {
	t.Helper()
	if body["code"] != float64(200) {
		t.Fatalf("code = %v, want 200; body = %v", body["code"], body)
	}
	if body["msg"] != "请求成功" {
		t.Fatalf("msg = %v, want 请求成功", body["msg"])
	}
	if body["data"] != nil {
		t.Fatalf("data = %v, want nil", body["data"])
	}
}

func assertPage(t *testing.T, page map[string]any, itemAssert func(*testing.T, map[string]any)) {
	t.Helper()
	assertKeys(t, page, "items", "total", "page", "size", "total_pages", "links")
	items, ok := page["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("items = %T len %d, want non-empty JSON array", page["items"], len(items))
	}
	itemAssert(t, assertMap(t, items[0]))
}

func assertNoticeDetail(t *testing.T, item map[string]any) {
	t.Helper()
	assertKeys(t, item, "id", "title", "type", "status", "content", "created_time", "updated_time")
	if _, ok := item["id"].(float64); !ok {
		t.Fatalf("id = %T, want JSON number", item["id"])
	}
	if _, ok := item["title"].(string); !ok {
		t.Fatalf("title = %T, want string", item["title"])
	}
	if _, ok := item["type"].(float64); !ok {
		t.Fatalf("type = %T, want JSON number", item["type"])
	}
	if _, ok := item["status"].(float64); !ok {
		t.Fatalf("status = %T, want JSON number", item["status"])
	}
	if _, ok := item["content"].(string); !ok {
		t.Fatalf("content = %T, want string", item["content"])
	}
	if _, ok := item["created_time"].(string); !ok {
		t.Fatalf("created_time = %T, want string", item["created_time"])
	}
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

func routeKeys(routes map[string]plugin.Route) []string {
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	return keys
}
