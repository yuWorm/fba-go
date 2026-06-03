package dict_test

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
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	dict "github.com/yuWorm/fba-plugin-dict"
	"gorm.io/gorm"
)

func TestDictPluginRegistersPythonCompatibleRoutes(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := dict.FBAPlugin().Register(ctx); err != nil {
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
		"GET /dict-types/all":              {authRequired: true},
		"GET /dict-types/:pk":              {authRequired: true},
		"GET /dict-types":                  {authRequired: true},
		"POST /dict-types":                 {authRequired: true, permission: "dict:type:add"},
		"PUT /dict-types/:pk":              {authRequired: true, permission: "dict:type:edit"},
		"DELETE /dict-types":               {authRequired: true, permission: "dict:type:del"},
		"GET /dict-datas/all":              {authRequired: true},
		"GET /dict-datas/:pk":              {authRequired: true},
		"GET /dict-datas/type-codes/:code": {authRequired: true},
		"GET /dict-datas":                  {authRequired: true},
		"POST /dict-datas":                 {authRequired: true, permission: "dict:data:add"},
		"PUT /dict-datas/:pk":              {authRequired: true, permission: "dict:data:edit"},
		"DELETE /dict-datas":               {authRequired: true, permission: "dict:data:del"},
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

func TestDictPluginRegistersMigrationWhenDBProviderExists(t *testing.T) {
	container := di.New()
	if err := container.Provide(func() db.Provider {
		return db.NewGORMProvider(&gorm.DB{}, nil)
	}); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})

	if err := dict.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	migrations := ctx.Migrations()
	if len(migrations) != 1 {
		t.Fatalf("migrations = %d, want 1", len(migrations))
	}
	if migrations[0].Scope != "plugin:dict" {
		t.Fatalf("migration scope = %q, want plugin:dict", migrations[0].Scope)
	}
}

func TestDictDataByTypeCodeMatchesPythonSchema(t *testing.T) {
	app := newDictApp(t)
	resp, body := requestJSON(t, app, "GET", "/api/v1/dict-datas/type-codes/sys_status", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeSlice(t, body)
	if len(data) != 2 {
		t.Fatalf("dict data count = %d, want 2", len(data))
	}
	item := assertMap(t, data[0])
	assertDictDataDetail(t, item)
	if item["type_code"] != "sys_status" {
		t.Fatalf("type_code = %v, want sys_status", item["type_code"])
	}
}

func TestDictTypesAllMatchesPythonSchema(t *testing.T) {
	app := newDictApp(t)
	resp, body := requestJSON(t, app, "GET", "/api/v1/dict-types/all", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeSlice(t, body)
	if len(data) == 0 {
		t.Fatal("dict types data is empty")
	}
	assertDictTypeDetail(t, assertMap(t, data[0]))
}

func TestDictPaginatedEndpointsMatchPythonPageSchema(t *testing.T) {
	app := newDictApp(t)
	for _, tc := range []struct {
		path      string
		itemTest  func(*testing.T, map[string]any)
		baseRoute string
	}{
		{path: "/api/v1/dict-types", itemTest: assertDictTypeDetail, baseRoute: "/api/v1/dict-types"},
		{path: "/api/v1/dict-datas", itemTest: assertDictDataDetail, baseRoute: "/api/v1/dict-datas"},
	} {
		resp, body := requestJSON(t, app, "GET", tc.path, "")
		assertStatusOK(t, resp)
		page := assertEnvelopeMap(t, body)
		assertKeys(t, page, "items", "total", "page", "size", "total_pages", "links")
		items, ok := page["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("%s items = %T len %d, want non-empty JSON array", tc.path, page["items"], len(items))
		}
		tc.itemTest(t, assertMap(t, items[0]))
		links := assertMap(t, page["links"])
		assertKeys(t, links, "first", "last", "self", "next", "prev")
		if !strings.HasPrefix(links["self"].(string), tc.baseRoute+"?") {
			t.Fatalf("%s self link = %v, want prefix %s?", tc.path, links["self"], tc.baseRoute)
		}
	}
}

func TestDictWriteEndpointsReturnPythonEnvelope(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/dict-types", `{"name":"fixture","code":"fixture","remark":null}`},
		{"PUT", "/api/v1/dict-types/1", `{"name":"fixture","code":"fixture","remark":null}`},
		{"DELETE", "/api/v1/dict-types", `{"pks":[1]}`},
		{"POST", "/api/v1/dict-datas", `{"type_id":1,"label":"Fixture","value":"fixture","color":null,"sort":0,"status":1,"remark":null}`},
		{"PUT", "/api/v1/dict-datas/1", `{"type_id":1,"label":"Fixture","value":"fixture","color":null,"sort":0,"status":1,"remark":null}`},
		{"DELETE", "/api/v1/dict-datas", `{"pks":[1]}`},
	} {
		app := newDictApp(t)
		resp, body := requestJSON(t, app, tc.method, tc.path, tc.body)
		assertStatusOK(t, resp)
		assertEnvelopeNull(t, body)
	}
}

func TestDictValidationErrorsMatchPython(t *testing.T) {
	app := newDictApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/dict-types/not-int", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusUnprocessableEntity, "请求参数非法: pk 输入应为有效的整数，无法将字符串解析为整数，输入：not-int")
}

func newDictApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	ctx := plugin.NewContext(plugin.ContextOptions{APIGroup: app.Group("/api/v1")})
	if err := dict.FBAPlugin().Register(ctx); err != nil {
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

func assertErrorEnvelope(t *testing.T, resp *http.Response, body map[string]any, status int, msg string) {
	t.Helper()
	if resp.StatusCode != status {
		t.Fatalf("status = %d, want %d; body = %v", resp.StatusCode, status, body)
	}
	if body["code"] != float64(status) {
		t.Fatalf("code = %v, want %d; body = %v", body["code"], status, body)
	}
	if body["msg"] != msg {
		t.Fatalf("msg = %v, want %s", body["msg"], msg)
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

func assertDictDataDetail(t *testing.T, data map[string]any) {
	t.Helper()
	assertKeys(t, data, "type_id", "label", "value", "color", "sort", "status", "remark", "id", "type_code", "created_time", "updated_time")
}

func assertDictTypeDetail(t *testing.T, data map[string]any) {
	t.Helper()
	assertKeys(t, data, "name", "code", "remark", "id", "created_time", "updated_time")
}

func routeKeys(routes map[string]plugin.Route) []string {
	keys := make([]string, 0, len(routes))
	for key := range routes {
		keys = append(keys, key)
	}
	return keys
}
