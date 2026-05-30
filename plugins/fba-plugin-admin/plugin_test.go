package admin_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		{"GET", "/api/v1/sys/users/1"},
		{"GET", "/api/v1/sys/users/1/roles"},
		{"GET", "/api/v1/sys/users"},
		{"GET", "/api/v1/sys/roles/all"},
		{"GET", "/api/v1/sys/roles/1/menus"},
		{"GET", "/api/v1/sys/roles/1/scopes"},
		{"GET", "/api/v1/sys/roles/1"},
		{"GET", "/api/v1/sys/roles"},
		{"GET", "/api/v1/sys/menus/sidebar"},
		{"GET", "/api/v1/sys/menus/1"},
		{"GET", "/api/v1/sys/menus"},
		{"GET", "/api/v1/sys/depts/1"},
		{"GET", "/api/v1/sys/depts"},
		{"GET", "/api/v1/sys/data-rules/models"},
		{"GET", "/api/v1/sys/data-rules/models/user/columns"},
		{"GET", "/api/v1/sys/data-rules/value-template-variables"},
		{"GET", "/api/v1/sys/data-rules/all"},
		{"GET", "/api/v1/sys/data-rules/1"},
		{"GET", "/api/v1/sys/data-rules"},
		{"GET", "/api/v1/sys/data-scopes/all"},
		{"GET", "/api/v1/sys/data-scopes/1"},
		{"GET", "/api/v1/sys/data-scopes/1/rules"},
		{"GET", "/api/v1/sys/data-scopes"},
		{"GET", "/api/v1/sys/plugins"},
		{"GET", "/api/v1/sys/plugins/changed"},
		{"GET", "/api/v1/sys/plugins/dict"},
		{"GET", "/api/v1/logs/login"},
		{"GET", "/api/v1/logs/opera"},
		{"GET", "/api/v1/monitors/server"},
		{"GET", "/api/v1/monitors/redis"},
		{"GET", "/api/v1/monitors/sessions"},
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

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/sys/users", `{"username":"contract_user","password":"Passw0rd!","nickname":"Contract User","email":null,"phone":null,"dept_id":1,"roles":[1]}`},
		{"PUT", "/api/v1/sys/users/1", `{"dept_id":null,"username":"admin","nickname":"Admin","avatar":null,"email":null,"phone":null,"roles":[1]}`},
		{"PUT", "/api/v1/sys/users/1/permissions?type=status", ""},
		{"PUT", "/api/v1/sys/users/me/password", `{"old_password":"old-password","new_password":"new-password","confirm_password":"new-password"}`},
		{"PUT", "/api/v1/sys/users/1/password", `{"password":"new-password"}`},
		{"PUT", "/api/v1/sys/users/me/nickname", `{"nickname":"Admin"}`},
		{"PUT", "/api/v1/sys/users/me/avatar", `{"avatar":"https://example.invalid/avatar.png"}`},
		{"PUT", "/api/v1/sys/users/me/email", `{"captcha":"123456","email":"admin@example.com"}`},
		{"DELETE", "/api/v1/sys/users/999999", ""},
		{"POST", "/api/v1/sys/roles", `{"name":"Contract Role","status":1,"is_filter_scopes":true,"remark":null}`},
		{"PUT", "/api/v1/sys/roles/1", `{"name":"admin","status":1,"is_filter_scopes":true,"remark":null}`},
		{"PUT", "/api/v1/sys/roles/1/menus", `{"menus":[1]}`},
		{"PUT", "/api/v1/sys/roles/1/scopes", `{"scopes":[1]}`},
		{"DELETE", "/api/v1/sys/roles", `{"pks":[999999]}`},
		{"POST", "/api/v1/sys/menus", `{"title":"Contract Menu","name":"ContractMenu","path":"/contract","parent_id":null,"sort":0,"icon":null,"type":1,"component":"Layout","perms":null,"status":1,"display":1,"cache":1,"link":null,"remark":null}`},
		{"PUT", "/api/v1/sys/menus/1", `{"title":"仪表盘","name":"Dashboard","path":"/dashboard","parent_id":null,"sort":0,"icon":"lucide:layout-dashboard","type":1,"component":"Layout","perms":null,"status":1,"display":1,"cache":1,"link":null,"remark":null}`},
		{"DELETE", "/api/v1/sys/menus/1", ""},
		{"POST", "/api/v1/sys/depts", `{"name":"Contract Dept","parent_id":null,"sort":0,"leader":null,"phone":null,"email":null,"status":1}`},
		{"PUT", "/api/v1/sys/depts/1", `{"name":"总部","parent_id":null,"sort":0,"leader":null,"phone":null,"email":null,"status":1}`},
		{"DELETE", "/api/v1/sys/depts/1", ""},
		{"POST", "/api/v1/sys/data-rules", `{"name":"Contract Rule","model":"user","column":"id","operator":0,"expression":0,"value":"{{ user_id }}"}`},
		{"PUT", "/api/v1/sys/data-rules/1", `{"name":"本人数据","model":"user","column":"id","operator":0,"expression":0,"value":"{{ user_id }}"}`},
		{"DELETE", "/api/v1/sys/data-rules", `{"pks":[999999]}`},
		{"POST", "/api/v1/sys/data-scopes", `{"name":"Contract Scope","status":1}`},
		{"PUT", "/api/v1/sys/data-scopes/1", `{"name":"本人数据范围","status":1}`},
		{"PUT", "/api/v1/sys/data-scopes/1/rules", `{"rules":[1]}`},
		{"DELETE", "/api/v1/sys/data-scopes", `{"pks":[999999]}`},
		{"POST", "/api/v1/sys/plugins?type=git&repo_url=https://example.invalid/plugin.git", ""},
		{"DELETE", "/api/v1/sys/plugins/dict", ""},
		{"PUT", "/api/v1/sys/plugins/dict/status", ""},
		{"DELETE", "/api/v1/logs/login", `{"pks":[999999]}`},
		{"DELETE", "/api/v1/logs/login/all", ""},
		{"DELETE", "/api/v1/logs/opera", `{"pks":[999999]}`},
		{"DELETE", "/api/v1/logs/opera/all", ""},
		{"DELETE", "/api/v1/monitors/sessions/1?session_uuid=fixture-session", ""},
	} {
		var reqBody io.Reader
		if tc.body != "" {
			reqBody = strings.NewReader(tc.body)
		}
		req := httptest.NewRequest(tc.method, tc.path, reqBody)
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
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

	uploadBody := "--fba-contract\r\nContent-Disposition: form-data; name=\"file\"; filename=\"contract.txt\"\r\nContent-Type: text/plain\r\n\r\ncontract\r\n--fba-contract--\r\n"
	req := httptest.NewRequest("POST", "/api/v1/sys/files/upload", strings.NewReader(uploadBody))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=fba-contract")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /api/v1/sys/files/upload error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /api/v1/sys/files/upload status = %d body = %s", resp.StatusCode, body)
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
		"GET /auth/captcha":                            false,
		"POST /auth/login/swagger":                     false,
		"POST /auth/login":                             false,
		"POST /auth/refresh":                           false,
		"POST /auth/logout":                            false,
		"GET /auth/codes":                              true,
		"GET /sys/users/me":                            true,
		"GET /sys/users/:pk":                           true,
		"GET /sys/users/:pk/roles":                     true,
		"GET /sys/users":                               true,
		"POST /sys/users":                              true,
		"PUT /sys/users/:pk":                           true,
		"PUT /sys/users/:pk/permissions":               true,
		"PUT /sys/users/me/password":                   true,
		"PUT /sys/users/:pk/password":                  true,
		"PUT /sys/users/me/nickname":                   true,
		"PUT /sys/users/me/avatar":                     true,
		"PUT /sys/users/me/email":                      true,
		"DELETE /sys/users/:pk":                        true,
		"GET /sys/roles/all":                           true,
		"GET /sys/roles/:pk/menus":                     true,
		"GET /sys/roles/:pk/scopes":                    true,
		"GET /sys/roles/:pk":                           true,
		"GET /sys/roles":                               true,
		"POST /sys/roles":                              true,
		"PUT /sys/roles/:pk":                           true,
		"PUT /sys/roles/:pk/menus":                     true,
		"PUT /sys/roles/:pk/scopes":                    true,
		"DELETE /sys/roles":                            true,
		"GET /sys/menus/sidebar":                       true,
		"GET /sys/menus/:pk":                           true,
		"GET /sys/menus":                               true,
		"POST /sys/menus":                              true,
		"PUT /sys/menus/:pk":                           true,
		"DELETE /sys/menus/:pk":                        true,
		"GET /sys/depts/:pk":                           true,
		"GET /sys/depts":                               true,
		"POST /sys/depts":                              true,
		"PUT /sys/depts/:pk":                           true,
		"DELETE /sys/depts/:pk":                        true,
		"GET /sys/data-rules/models":                   true,
		"GET /sys/data-rules/models/:model/columns":    true,
		"GET /sys/data-rules/value-template-variables": true,
		"GET /sys/data-rules/all":                      true,
		"GET /sys/data-rules/:pk":                      true,
		"GET /sys/data-rules":                          true,
		"POST /sys/data-rules":                         true,
		"PUT /sys/data-rules/:pk":                      true,
		"DELETE /sys/data-rules":                       true,
		"GET /sys/data-scopes/all":                     true,
		"GET /sys/data-scopes/:pk":                     true,
		"GET /sys/data-scopes/:pk/rules":               true,
		"GET /sys/data-scopes":                         true,
		"POST /sys/data-scopes":                        true,
		"PUT /sys/data-scopes/:pk":                     true,
		"PUT /sys/data-scopes/:pk/rules":               true,
		"DELETE /sys/data-scopes":                      true,
		"POST /sys/files/upload":                       true,
		"GET /sys/plugins":                             true,
		"GET /sys/plugins/changed":                     true,
		"GET /sys/plugins/:plugin":                     true,
		"POST /sys/plugins":                            true,
		"DELETE /sys/plugins/:plugin":                  true,
		"PUT /sys/plugins/:plugin/status":              true,
		"GET /logs/login":                              true,
		"DELETE /logs/login":                           true,
		"DELETE /logs/login/all":                       true,
		"GET /logs/opera":                              true,
		"DELETE /logs/opera":                           true,
		"DELETE /logs/opera/all":                       true,
		"GET /monitors/server":                         true,
		"GET /monitors/redis":                          true,
		"GET /monitors/sessions":                       true,
		"DELETE /monitors/sessions/:pk":                true,
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

	wantPermissions := map[string]string{
		"POST /sys/roles":                "sys:role:add",
		"PUT /sys/roles/:pk":             "sys:role:edit",
		"PUT /sys/roles/:pk/menus":       "sys:role:menu:edit",
		"DELETE /sys/roles":              "sys:role:del",
		"POST /sys/menus":                "sys:menu:add",
		"PUT /sys/menus/:pk":             "sys:menu:edit",
		"DELETE /sys/menus/:pk":          "sys:menu:del",
		"POST /sys/data-rules":           "data:rule:add",
		"PUT /sys/data-rules/:pk":        "data:rule:edit",
		"DELETE /sys/data-rules":         "data:rule:del",
		"POST /sys/data-scopes":          "data:scope:add",
		"PUT /sys/data-scopes/:pk":       "data:scope:edit",
		"PUT /sys/data-scopes/:pk/rules": "data:scope:rule:edit",
		"DELETE /sys/data-scopes":        "data:scope:del",
		"DELETE /sys/users/:pk":          "sys:user:del",
		"POST /sys/files/upload":         "sys:file:upload",
		"DELETE /logs/login":             "log:login:del",
		"DELETE /logs/login/all":         "log:login:clear",
		"DELETE /logs/opera":             "log:opera:del",
		"DELETE /logs/opera/all":         "log:opera:clear",
	}
	for key, permission := range wantPermissions {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, maps.Keys(got))
		}
		if route.Permission != permission {
			t.Fatalf("%s Permission = %q, want %q", key, route.Permission, permission)
		}
	}

	for _, key := range []string{
		"POST /sys/users",
		"PUT /sys/users/:pk",
		"PUT /sys/users/:pk/permissions",
		"PUT /sys/users/me/password",
		"PUT /sys/users/:pk/password",
		"PUT /sys/users/me/nickname",
		"PUT /sys/users/me/avatar",
		"PUT /sys/users/me/email",
	} {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, maps.Keys(got))
		}
		if route.Permission != "" {
			t.Fatalf("%s Permission = %q, want empty", key, route.Permission)
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

func TestUserEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"operator","password":"Passw0rd!","nickname":"Operator","email":"operator@example.com","phone":"13900000000","dept_id":1,"roles":[1]}`)
	assertStatusOK(t, resp)
	created := assertEnvelopeMap(t, body)
	if created["username"] != "operator" {
		t.Fatalf("created user username = %v, want operator", created["username"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users?dept=1&username=oper&phone=139&status=1", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items, ok := page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered user items = %T len %d, want one item", page["items"], len(items))
	}
	user := assertMap(t, items[0])
	if user["username"] != "operator" {
		t.Fatalf("filtered user username = %v, want operator", user["username"])
	}
	id := int(user["id"].(float64))

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users/"+itoa(id)+"/roles", "")
	assertStatusOK(t, resp)
	roles := assertEnvelopeSlice(t, body)
	if len(roles) != 1 {
		t.Fatalf("user roles count = %d, want 1", len(roles))
	}
	role := assertMap(t, roles[0])
	if role["id"] != float64(1) {
		t.Fatalf("user role id = %v, want 1", role["id"])
	}

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(id), `{"dept_id":1,"username":"operator","nickname":"Operator Updated","avatar":null,"email":"operator-updated@example.com","phone":"13900000001","roles":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["nickname"] != "Operator Updated" {
		t.Fatalf("updated user nickname = %v, want Operator Updated", detail["nickname"])
	}

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(id)+"/permissions?type=status", "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail = assertEnvelopeMap(t, body)
	if detail["status"] != float64(0) {
		t.Fatalf("toggled user status = %v, want 0", detail["status"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/users/"+itoa(id), "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users?username=operator", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted user items = %T len %d, want empty list", page["items"], len(items))
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

func TestMenuEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/menus", `{"title":"Reports","name":"Reports","path":"/reports","parent_id":null,"sort":9,"icon":"lucide:file-bar-chart","type":1,"component":"/reports/index","perms":null,"status":0,"display":1,"cache":1,"link":null,"remark":"report menu"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus?title=Report&status=0", "")
	assertStatusOK(t, resp)
	items := assertEnvelopeSlice(t, body)
	if len(items) != 1 {
		t.Fatalf("filtered menu count = %d, want 1", len(items))
	}
	report := assertMap(t, items[0])
	if report["title"] != "Reports" {
		t.Fatalf("filtered menu title = %v, want Reports", report["title"])
	}
	if report["status"] != float64(0) {
		t.Fatalf("filtered menu status = %v, want 0", report["status"])
	}
	id := int(report["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/menus/"+itoa(id), `{"title":"Reports Updated","name":"Reports","path":"/reports","parent_id":null,"sort":9,"icon":"lucide:file-bar-chart","type":1,"component":"/reports/index","perms":null,"status":1,"display":1,"cache":1,"link":null,"remark":null}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["title"] != "Reports Updated" {
		t.Fatalf("updated menu title = %v, want Reports Updated", detail["title"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/menus/"+itoa(id), "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus?title=Reports", "")
	assertStatusOK(t, resp)
	items = assertEnvelopeSlice(t, body)
	if len(items) != 0 {
		t.Fatalf("deleted menu count = %d, want 0", len(items))
	}
}

func TestMenuTreeAndSidebarReflectCreatedChildren(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/menus", `{"title":"Reports Child","name":"ReportsChild","path":"/dashboard/reports","parent_id":1,"sort":1,"icon":"lucide:file-bar-chart","type":1,"component":"/reports/index","perms":null,"status":1,"display":1,"cache":1,"link":null,"remark":null}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus/2", "")
	assertStatusOK(t, resp)
	created := assertEnvelopeMap(t, body)
	if created["title"] != "Reports Child" {
		t.Fatalf("created menu title = %v, want Reports Child", created["title"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus", "")
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)
	child := findNodeInTree(t, tree, "Reports Child")
	if child["parent_id"] != float64(1) {
		t.Fatalf("created child parent_id = %v, want 1", child["parent_id"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus/sidebar", "")
	assertStatusOK(t, resp)
	sidebar := assertEnvelopeSlice(t, body)
	sidebarChild := findNodeInTree(t, sidebar, "ReportsChild")
	meta := assertMap(t, sidebarChild["meta"])
	if meta["title"] != "Reports Child" {
		t.Fatalf("sidebar child title = %v, want Reports Child", meta["title"])
	}
}

func TestDeptEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Engineering","parent_id":null,"sort":9,"leader":"Jane","phone":"13800000000","email":"eng@example.com","status":0}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts?name=Engineer&leader=Jane&phone=138&status=0", "")
	assertStatusOK(t, resp)
	items := assertEnvelopeSlice(t, body)
	if len(items) != 1 {
		t.Fatalf("filtered dept count = %d, want 1", len(items))
	}
	dept := assertMap(t, items[0])
	if dept["name"] != "Engineering" {
		t.Fatalf("filtered dept name = %v, want Engineering", dept["name"])
	}
	if dept["status"] != float64(0) {
		t.Fatalf("filtered dept status = %v, want 0", dept["status"])
	}
	id := int(dept["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/depts/"+itoa(id), `{"name":"Engineering Updated","parent_id":null,"sort":10,"leader":"Jane","phone":"13800000000","email":"eng@example.com","status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["name"] != "Engineering Updated" {
		t.Fatalf("updated dept name = %v, want Engineering Updated", detail["name"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/depts/"+itoa(id), "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts?name=Engineering", "")
	assertStatusOK(t, resp)
	items = assertEnvelopeSlice(t, body)
	if len(items) != 0 {
		t.Fatalf("deleted dept count = %d, want 0", len(items))
	}
}

func TestDeptTreeReflectsCreatedChildren(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Platform","parent_id":1,"sort":1,"leader":"Alex","phone":"13900000000","email":"platform@example.com","status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts/2", "")
	assertStatusOK(t, resp)
	created := assertEnvelopeMap(t, body)
	if created["name"] != "Platform" {
		t.Fatalf("created dept name = %v, want Platform", created["name"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts", "")
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)
	child := findNodeInTree(t, tree, "Platform")
	if child["parent_id"] != float64(1) {
		t.Fatalf("created child parent_id = %v, want 1", child["parent_id"])
	}
}

func TestDataRuleEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/data-rules", `{"name":"Engineering Rule","model":"user","column":"dept_id","operator":0,"expression":0,"value":"{{ dept_id }}"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules?name=Engineering", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items, ok := page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered data-rule items = %T len %d, want one item", page["items"], len(items))
	}
	rule := assertMap(t, items[0])
	if rule["name"] != "Engineering Rule" {
		t.Fatalf("filtered data-rule name = %v, want Engineering Rule", rule["name"])
	}
	if rule["column"] != "dept_id" {
		t.Fatalf("filtered data-rule column = %v, want dept_id", rule["column"])
	}
	id := int(rule["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/data-rules/"+itoa(id), `{"name":"Engineering Rule Updated","model":"user","column":"id","operator":1,"expression":1,"value":"{{ user_id }}"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["name"] != "Engineering Rule Updated" {
		t.Fatalf("updated data-rule name = %v, want Engineering Rule Updated", detail["name"])
	}
	if detail["operator"] != float64(1) {
		t.Fatalf("updated data-rule operator = %v, want 1", detail["operator"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/data-rules", `{"pks":[`+itoa(id)+`]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules?name=Engineering", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted data-rule items = %T len %d, want empty list", page["items"], len(items))
	}
}

func TestDataScopeEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/data-scopes", `{"name":"Platform Scope","status":0}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-scopes?name=Platform&status=0", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items, ok := page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered data-scope items = %T len %d, want one item", page["items"], len(items))
	}
	scope := assertMap(t, items[0])
	if scope["name"] != "Platform Scope" {
		t.Fatalf("filtered data-scope name = %v, want Platform Scope", scope["name"])
	}
	if scope["status"] != float64(0) {
		t.Fatalf("filtered data-scope status = %v, want 0", scope["status"])
	}
	id := int(scope["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/data-scopes/"+itoa(id), `{"name":"Platform Scope Updated","status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-scopes/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["name"] != "Platform Scope Updated" {
		t.Fatalf("updated data-scope name = %v, want Platform Scope Updated", detail["name"])
	}

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/data-scopes/"+itoa(id)+"/rules", `{"rules":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-scopes/"+itoa(id)+"/rules", "")
	assertStatusOK(t, resp)
	withRules := assertEnvelopeMap(t, body)
	rules, ok := withRules["rules"].([]any)
	if !ok || len(rules) != 1 {
		t.Fatalf("data-scope rules = %T len %d, want one rule", withRules["rules"], len(rules))
	}
	rule := assertMap(t, rules[0])
	if rule["id"] != float64(1) {
		t.Fatalf("data-scope rule id = %v, want 1", rule["id"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/data-scopes", `{"pks":[`+itoa(id)+`]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-scopes?name=Platform", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted data-scope items = %T len %d, want empty list", page["items"], len(items))
	}
}

func TestRoleEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/roles", `{"name":"Auditor","status":0,"is_filter_scopes":false,"remark":"read only"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles?name=Audit&status=0", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items, ok := page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered role items = %T len %d, want one item", page["items"], len(items))
	}
	auditor := assertMap(t, items[0])
	if auditor["name"] != "Auditor" {
		t.Fatalf("filtered role name = %v, want Auditor", auditor["name"])
	}
	if auditor["status"] != float64(0) {
		t.Fatalf("filtered role status = %v, want 0", auditor["status"])
	}
	id := int(auditor["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/roles/"+itoa(id), `{"name":"Auditor Updated","status":1,"is_filter_scopes":true,"remark":null}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles/"+itoa(id), "")
	assertStatusOK(t, resp)
	detail := assertEnvelopeMap(t, body)
	if detail["name"] != "Auditor Updated" {
		t.Fatalf("updated role name = %v, want Auditor Updated", detail["name"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/roles", `{"pks":[`+itoa(id)+`]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles?name=Auditor", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted role items = %T len %d, want empty list", page["items"], len(items))
	}
}

func TestRoleRelationEndpointsAreStateful(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "PUT", "/api/v1/sys/roles/1/menus", `{"menus":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles/1/menus", "")
	assertStatusOK(t, resp)
	menus := assertEnvelopeSlice(t, body)
	if len(menus) != 1 {
		t.Fatalf("role menu count = %d, want 1", len(menus))
	}
	menu := assertMap(t, menus[0])
	if menu["id"] != float64(1) {
		t.Fatalf("role menu id = %v, want 1", menu["id"])
	}

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/roles/1/scopes", `{"scopes":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles/1/scopes", "")
	assertStatusOK(t, resp)
	scopes := assertEnvelopeSlice(t, body)
	if len(scopes) != 1 || scopes[0] != float64(1) {
		t.Fatalf("role scopes = %v, want [1]", scopes)
	}
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

func assertEnvelopeNil(t *testing.T, body map[string]any) {
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

func findNodeInTree(t *testing.T, items []any, name string) map[string]any {
	t.Helper()
	for _, raw := range items {
		item := assertMap(t, raw)
		if item["name"] == name || item["title"] == name {
			return item
		}
		children, _ := item["children"].([]any)
		if len(children) == 0 {
			continue
		}
		if found := findNodeInTree(t, children, name); found != nil {
			return found
		}
	}
	t.Fatalf("menu %q not found in tree %v", name, items)
	return nil
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

func itoa(value int) string {
	return strconv.Itoa(value)
}
