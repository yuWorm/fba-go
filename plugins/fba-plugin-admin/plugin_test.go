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
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	admin "github.com/yuWorm/fba-plugin-admin"
	"gorm.io/gorm"
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
		{"PUT", "/api/v1/sys/users/1/permissions?type=multi_login", ""},
		{"PUT", "/api/v1/sys/users/me/password", `{"old_password":"old-password","new_password":"new-password","confirm_password":"new-password"}`},
		{"PUT", "/api/v1/sys/users/1/password", `{"password":"new-password"}`},
		{"PUT", "/api/v1/sys/users/me/nickname", `{"nickname":"Admin"}`},
		{"PUT", "/api/v1/sys/users/me/avatar", `{"avatar":"https://example.invalid/avatar.png"}`},
		{"PUT", "/api/v1/sys/users/me/email", `{"captcha":"123456","email":"admin@example.com"}`},
		{"DELETE", "/api/v1/sys/users/2", ""},
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
		{"DELETE", "/api/v1/sys/depts/2", ""},
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

func TestAdminPluginRegistersMigrationWhenDBProviderExists(t *testing.T) {
	container := di.New()
	if err := container.Provide(func() db.Provider {
		return db.NewGORMProvider(&gorm.DB{}, nil)
	}); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})

	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	migrations := ctx.Migrations()
	if len(migrations) != 1 {
		t.Fatalf("migrations = %d, want 1", len(migrations))
	}
	if migrations[0].Scope != "plugin:admin" {
		t.Fatalf("migration scope = %q, want plugin:admin", migrations[0].Scope)
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
		"PUT /sys/roles/:pk/scopes":      "sys:role:scope:edit",
		"DELETE /sys/roles":              "sys:role:del",
		"POST /sys/menus":                "sys:menu:add",
		"PUT /sys/menus/:pk":             "sys:menu:edit",
		"DELETE /sys/menus/:pk":          "sys:menu:del",
		"POST /sys/depts":                "sys:dept:add",
		"PUT /sys/depts/:pk":             "sys:dept:edit",
		"DELETE /sys/depts/:pk":          "sys:dept:del",
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
		"PUT /sys/users/:pk/password",
		"GET /sys/plugins",
		"GET /sys/plugins/changed",
		"POST /sys/plugins",
		"DELETE /sys/plugins/:plugin",
		"PUT /sys/plugins/:plugin/status",
		"GET /sys/plugins/:plugin",
		"GET /monitors/server",
		"GET /monitors/sessions",
		"DELETE /monitors/sessions/:pk",
	} {
		route, ok := got[key]
		if !ok {
			t.Fatalf("route %s not registered; registered routes: %v", key, maps.Keys(got))
		}
		if !route.SuperuserRequired {
			t.Fatalf("%s SuperuserRequired = false, want true", key)
		}
	}

	for _, key := range []string{
		"PUT /sys/users/me/password",
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

func TestAuthEndpointsAreStatefulAndValidateUsers(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/auth/captcha", "")
	assertStatusOK(t, resp)
	captcha := assertEnvelopeMap(t, body)
	if captcha["uuid"] == "fixture-captcha" {
		t.Fatalf("captcha uuid = %v, want dynamic uuid", captcha["uuid"])
	}
	if captcha["image"] == "" {
		t.Fatal("captcha image is empty")
	}

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"auth_user","password":"secret","nickname":"Auth User","email":null,"phone":null,"dept_id":1,"roles":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeMap(t, body)

	loginBody := `{"username":"auth_user","password":"secret","uuid":"` + captcha["uuid"].(string) + `","captcha":"1234"}`
	resp, body = requestJSON(t, app, "POST", "/api/v1/auth/login", loginBody)
	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	if data["access_token"] == "fixture-access-token" {
		t.Fatal("login returned fixture access token")
	}
	if data["session_uuid"] == "fixture-session" {
		t.Fatal("login returned fixture session uuid")
	}
	user := assertMap(t, data["user"])
	if user["username"] != "auth_user" {
		t.Fatalf("login user = %v, want auth_user", user["username"])
	}
	sessionUUID := data["session_uuid"].(string)
	refreshCookie := requireCookie(t, resp.Header.Get("Set-Cookie"), "fba_refresh_token")

	resp, body = requestJSON(t, app, "GET", "/api/v1/monitors/sessions?username=auth_user", "")
	assertStatusOK(t, resp)
	sessions := assertEnvelopeSlice(t, body)
	if len(sessions) != 1 {
		t.Fatalf("auth_user sessions = %d, want 1", len(sessions))
	}
	session := assertMap(t, sessions[0])
	if session["session_uuid"] != sessionUUID {
		t.Fatalf("session uuid = %v, want %s", session["session_uuid"], sessionUUID)
	}

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	req.AddCookie(refreshCookie)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /auth/refresh error = %v", err)
	}
	defer resp.Body.Close()
	assertStatusOK(t, resp)
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode refresh response: %v", err)
	}
	refreshed := assertEnvelopeMap(t, body)
	if refreshed["session_uuid"] != sessionUUID {
		t.Fatalf("refreshed session uuid = %v, want %s", refreshed["session_uuid"], sessionUUID)
	}
	if refreshed["access_token"] == data["access_token"] {
		t.Fatal("refresh returned the same access token")
	}

	req = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+refreshed["access_token"].(string))
	req.AddCookie(refreshCookie)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("POST /auth/logout error = %v", err)
	}
	defer resp.Body.Close()
	assertStatusOK(t, resp)

	resp, body = requestJSON(t, app, "GET", "/api/v1/monitors/sessions?username=auth_user", "")
	assertStatusOK(t, resp)
	sessions = assertEnvelopeSlice(t, body)
	if len(sessions) != 0 {
		t.Fatalf("auth_user sessions after logout = %d, want 0", len(sessions))
	}

	resp, _ = requestRaw(t, app, "POST", "/api/v1/auth/login", `{"username":"auth_user","password":"wrong","uuid":"fixture-captcha","captcha":"1234"}`)
	if resp.StatusCode == fiber.StatusOK {
		t.Fatal("login with wrong password returned 200")
	}
}

func TestAdminRuntimeAuthUsesTokenUserAndRBAC(t *testing.T) {
	app := newAdminRuntimeApp(t)

	resp, _ := requestRaw(t, app, "GET", "/api/v1/sys/users/me", "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("GET /sys/users/me without token status = %d, want 401", resp.StatusCode)
	}

	adminToken := loginForAccessToken(t, app, "admin", "admin")
	resp, body := requestJSONAuth(t, app, "POST", "/api/v1/sys/users", `{"username":"viewer","password":"secret","nickname":"Viewer","email":null,"phone":null,"dept_id":1,"roles":[1]}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeMap(t, body)

	viewerToken := loginForAccessToken(t, app, "viewer", "secret")
	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/sys/users/me", "", viewerToken)
	assertStatusOK(t, resp)
	current := assertEnvelopeMap(t, body)
	if current["username"] != "viewer" {
		t.Fatalf("current username = %v, want viewer", current["username"])
	}

	resp, _ = requestRawAuth(t, app, "POST", "/api/v1/sys/roles", `{"name":"Blocked","status":1,"is_filter_scopes":false,"remark":null}`, viewerToken)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("viewer POST /sys/roles status = %d, want 403", resp.StatusCode)
	}

	resp, _ = requestRawAuth(t, app, "GET", "/api/v1/sys/plugins", "", viewerToken)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("viewer GET /sys/plugins status = %d, want 403", resp.StatusCode)
	}
}

func TestAdminRuntimeDeptTreeAppliesPythonDataPermission(t *testing.T) {
	app := newAdminRuntimeApp(t)
	adminToken := loginForAccessToken(t, app, "admin", "admin")

	resp, body := requestJSONAuth(t, app, "POST", "/api/v1/sys/depts", `{"name":"Platform","parent_id":1,"sort":1,"leader":"Alex","phone":"13900000000","email":"platform@example.com","status":1}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSONAuth(t, app, "POST", "/api/v1/sys/depts", `{"name":"ExternalRoot","parent_id":null,"sort":2,"leader":"Lee","phone":"13600000000","email":"external@example.com","status":1}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSONAuth(t, app, "POST", "/api/v1/sys/data-rules", `{"name":"OwnDeptChildren","model":"dept","column":"parent_id","operator":0,"expression":0,"value":"${dept_id}"}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/sys/data-rules?name=OwnDeptChildren", "", adminToken)
	assertStatusOK(t, resp)
	rulePage := assertEnvelopeMap(t, body)
	ruleItems := assertSlice(t, rulePage["items"])
	if len(ruleItems) != 1 {
		t.Fatalf("data rule items = %d, want 1", len(ruleItems))
	}
	ruleID := int(assertMap(t, ruleItems[0])["id"].(float64))

	resp, body = requestJSONAuth(t, app, "POST", "/api/v1/sys/data-scopes", `{"name":"OwnDeptScope","status":1}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/sys/data-scopes?name=OwnDeptScope", "", adminToken)
	assertStatusOK(t, resp)
	scopePage := assertEnvelopeMap(t, body)
	scopeItems := assertSlice(t, scopePage["items"])
	if len(scopeItems) != 1 {
		t.Fatalf("data scope items = %d, want 1", len(scopeItems))
	}
	scopeID := int(assertMap(t, scopeItems[0])["id"].(float64))
	resp, body = requestJSONAuth(t, app, "PUT", "/api/v1/sys/data-scopes/"+itoa(scopeID)+"/rules", `{"rules":[`+itoa(ruleID)+`]}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSONAuth(t, app, "POST", "/api/v1/sys/roles", `{"name":"DeptLimited","status":1,"is_filter_scopes":true,"remark":null}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/sys/roles?name=DeptLimited", "", adminToken)
	assertStatusOK(t, resp)
	rolePage := assertEnvelopeMap(t, body)
	roleItems := assertSlice(t, rolePage["items"])
	if len(roleItems) != 1 {
		t.Fatalf("role items = %d, want 1", len(roleItems))
	}
	roleID := int(assertMap(t, roleItems[0])["id"].(float64))
	resp, body = requestJSONAuth(t, app, "PUT", "/api/v1/sys/roles/"+itoa(roleID)+"/scopes", `{"scopes":[`+itoa(scopeID)+`]}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSONAuth(t, app, "POST", "/api/v1/sys/users", `{"username":"dept_limited","password":"secret","nickname":"Dept Limited","email":null,"phone":null,"dept_id":1,"roles":[`+itoa(roleID)+`]}`, adminToken)
	assertStatusOK(t, resp)
	assertEnvelopeMap(t, body)
	limitedToken := loginForAccessToken(t, app, "dept_limited", "secret")

	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/sys/depts", "", limitedToken)
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)
	platform := findNodeInTree(t, tree, "Platform")
	if platform["parent_id"] != float64(1) {
		t.Fatalf("Platform parent_id = %v, want 1", platform["parent_id"])
	}
	if _, ok := findNodeInTreeValue(t, tree, "ExternalRoot"); ok {
		t.Fatalf("ExternalRoot is visible in data-filtered dept tree: %v", tree)
	}
}

func TestAdminRuntimeSwaggerTokenAuthorizesRoutes(t *testing.T) {
	app := newAdminRuntimeApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/auth/login/swagger", "")
	assertStatusOK(t, resp)
	token, ok := body["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("swagger access_token = %v, want non-empty string", body["access_token"])
	}

	resp, body = requestJSONAuth(t, app, "GET", "/api/v1/auth/codes", "", token)
	assertStatusOK(t, resp)
	assertEnvelopeSlice(t, body)
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

func TestUserEndpointsApplyPythonCRUDGuards(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_a","password":"secret","nickname":null,"email":"guard-a@example.com","phone":null,"dept_id":1,"roles":[1]}`)
	assertStatusOK(t, resp)
	guardA := assertEnvelopeMap(t, body)
	guardAID := int(guardA["id"].(float64))

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_a","password":"secret","nickname":null,"email":"guard-a-copy@example.com","phone":null,"dept_id":1,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "用户名已注册")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_email","password":"secret","nickname":null,"email":"guard-a@example.com","phone":null,"dept_id":1,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "邮箱已被绑定")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_empty_password","password":"","nickname":null,"email":null,"phone":null,"dept_id":1,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusBadRequest, "密码不允许为空")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_missing_dept","password":"secret","nickname":null,"email":null,"phone":null,"dept_id":999999,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "部门不存在")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_missing_role","password":"secret","nickname":null,"email":null,"phone":null,"dept_id":1,"roles":[999999]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "角色不存在")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_b","password":"secret","nickname":"Guard B","email":"guard-b@example.com","phone":null,"dept_id":1,"roles":[1]}`)
	assertStatusOK(t, resp)
	guardB := assertEnvelopeMap(t, body)
	guardBID := int(guardB["id"].(float64))

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/999999", `{"dept_id":1,"username":"missing","nickname":"Missing","avatar":null,"email":null,"phone":null,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "用户不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(guardBID), `{"dept_id":1,"username":"guard_a","nickname":"Guard B","avatar":null,"email":"guard-b@example.com","phone":null,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "用户名已注册")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(guardBID), `{"dept_id":1,"username":"guard_b","nickname":"Guard B","avatar":null,"email":"guard-a@example.com","phone":null,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "邮箱已被绑定")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(guardBID), `{"dept_id":999999,"username":"guard_b","nickname":"Guard B","avatar":null,"email":"guard-b@example.com","phone":null,"roles":[1]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "部门不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(guardBID), `{"dept_id":1,"username":"guard_b","nickname":"Guard B","avatar":null,"email":"guard-b@example.com","phone":null,"roles":[999999]}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "角色不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/999999/permissions?type=status", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "用户不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/1/permissions?type=status", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusForbidden, "禁止修改自身权限")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(guardBID)+"/permissions?type=missing", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusBadRequest, "权限类型不存在")

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/users/999999", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "用户不存在")

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/users/"+itoa(guardAID), "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"guard_a","password":"secret","nickname":null,"email":"guard-a@example.com","phone":null,"dept_id":1,"roles":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeMap(t, body)
}

func TestCurrentUserProfileEndpointsAreStateful(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "PUT", "/api/v1/sys/users/me/nickname", `{"nickname":"Admin Updated"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/me/avatar", `{"avatar":"https://example.invalid/avatar.png"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/me/email", `{"captcha":"123456","email":"admin-updated@example.com"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/1/password", `{"password":"reset-password"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/users/me/password", `{"old_password":"reset-password","new_password":"new-password","confirm_password":"new-password"}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/users/me", "")
	assertStatusOK(t, resp)
	current := assertEnvelopeMap(t, body)
	if current["nickname"] != "Admin Updated" {
		t.Fatalf("current user nickname = %v, want Admin Updated", current["nickname"])
	}
	if current["avatar"] != "https://example.invalid/avatar.png" {
		t.Fatalf("current user avatar = %v, want updated avatar", current["avatar"])
	}
	if current["email"] != "admin-updated@example.com" {
		t.Fatalf("current user email = %v, want admin-updated@example.com", current["email"])
	}
	roles, ok := current["roles"].([]any)
	if !ok || len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("current user roles = %v, want [admin]", current["roles"])
	}
}

func TestPluginEndpointsAreStatefulAndPythonCompatible(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/plugins", "")
	assertStatusOK(t, resp)
	plugins := assertEnvelopeSlice(t, body)
	dictPlugin := assertMap(t, findPluginByName(t, plugins, "dict"))
	dictInfo := assertMap(t, dictPlugin["plugin"])
	if dictInfo["enable"] != "1" {
		t.Fatalf("dict plugin enable = %v, want 1", dictInfo["enable"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/plugins/changed", "")
	assertStatusOK(t, resp)
	if body["data"] != false {
		t.Fatalf("plugin changed = %v, want false", body["data"])
	}

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/plugins/dict/status", "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/plugins", "")
	assertStatusOK(t, resp)
	plugins = assertEnvelopeSlice(t, body)
	dictPlugin = assertMap(t, findPluginByName(t, plugins, "dict"))
	dictInfo = assertMap(t, dictPlugin["plugin"])
	if dictInfo["enable"] != "0" {
		t.Fatalf("dict plugin enable after toggle = %v, want 0", dictInfo["enable"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/plugins/changed", "")
	assertStatusOK(t, resp)
	if body["data"] != true {
		t.Fatalf("plugin changed after status toggle = %v, want true", body["data"])
	}

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/plugins?type=git&repo_url=https://example.invalid/analytics.git", "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/plugins", "")
	assertStatusOK(t, resp)
	plugins = assertEnvelopeSlice(t, body)
	analytics := assertMap(t, findPluginByName(t, plugins, "analytics"))
	analyticsInfo := assertMap(t, analytics["plugin"])
	if analyticsInfo["enable"] != "1" {
		t.Fatalf("installed plugin enable = %v, want 1", analyticsInfo["enable"])
	}

	resp, raw := requestRaw(t, app, "GET", "/api/v1/sys/plugins/analytics", "")
	assertStatusOK(t, resp)
	if strings.Contains(raw, `"code"`) {
		t.Fatalf("download plugin response is enveloped JSON: %s", raw)
	}
	if !strings.Contains(raw, "analytics") {
		t.Fatalf("download plugin body = %q, want plugin name", raw)
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/plugins/analytics", "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/plugins", "")
	assertStatusOK(t, resp)
	plugins = assertEnvelopeSlice(t, body)
	if hasPluginByName(plugins, "analytics") {
		t.Fatal("analytics plugin still present after uninstall")
	}
}

func TestLogEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/logs/login?username=admin&status=1&ip=127", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	items, ok := page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered login log items = %T len %d, want one item", page["items"], len(items))
	}
	loginLog := assertMap(t, items[0])
	assertKeys(t, loginLog, "id", "user_uuid", "username", "status", "ip", "country", "region", "city", "user_agent", "browser", "os", "device", "msg", "login_time", "created_time")
	if loginLog["username"] != "admin" {
		t.Fatalf("login log username = %v, want admin", loginLog["username"])
	}
	loginLogID := int(loginLog["id"].(float64))

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/logs/login", `{"pks":[`+itoa(loginLogID)+`]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/logs/login?username=admin", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted login log items = %T len %d, want empty list", page["items"], len(items))
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/logs/opera?username=admin&status=1&ip=127", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("filtered opera log items = %T len %d, want one item", page["items"], len(items))
	}
	operaLog := assertMap(t, items[0])
	assertKeys(t, operaLog, "id", "trace_id", "username", "method", "title", "path", "ip", "country", "region", "city", "user_agent", "browser", "os", "device", "args", "status", "code", "msg", "cost_time", "opera_time", "created_time")
	if operaLog["path"] != "/api/v1/sys/users" {
		t.Fatalf("opera log path = %v, want /api/v1/sys/users", operaLog["path"])
	}
	operaLogID := int(operaLog["id"].(float64))

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/logs/opera", `{"pks":[`+itoa(operaLogID)+`]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/logs/opera?username=admin", "")
	assertStatusOK(t, resp)
	page = assertEnvelopeMap(t, body)
	items, ok = page["items"].([]any)
	if !ok || len(items) != 0 {
		t.Fatalf("deleted opera log items = %T len %d, want empty list", page["items"], len(items))
	}
}

func TestUploadFileUsesMultipartFilenameLikePython(t *testing.T) {
	app := newAdminApp(t)

	uploadBody := "--fba-upload\r\nContent-Disposition: form-data; name=\"file\"; filename=\"audit.txt\"\r\nContent-Type: text/plain\r\n\r\nhello\r\n--fba-upload--\r\n"
	req := httptest.NewRequest("POST", "/api/v1/sys/files/upload", strings.NewReader(uploadBody))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=fba-upload")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /sys/files/upload error = %v", err)
	}
	defer resp.Body.Close()
	assertStatusOK(t, resp)
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode upload body: %v", err)
	}
	data := assertEnvelopeMap(t, body)
	if data["url"] != "/static/upload/audit.txt" {
		t.Fatalf("upload url = %v, want /static/upload/audit.txt", data["url"])
	}
}

func TestMonitorEndpointsUseServiceData(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/monitors/server", "")
	assertStatusOK(t, resp)
	server := assertEnvelopeMap(t, body)
	cpu := assertMap(t, server["cpu"])
	if cpu["logical_num"].(float64) <= 0 {
		t.Fatalf("server cpu logical_num = %v, want > 0", cpu["logical_num"])
	}
	mem := assertMap(t, server["mem"])
	if mem["total"].(float64) <= 0 {
		t.Fatalf("server mem total = %v, want > 0", mem["total"])
	}
	service := assertMap(t, server["service"])
	if service["name"] != "fba-go" {
		t.Fatalf("server service name = %v, want fba-go", service["name"])
	}
	if service["home"] == "" {
		t.Fatal("server service home is empty")
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/monitors/redis", "")
	assertStatusOK(t, resp)
	redis := assertEnvelopeMap(t, body)
	info := assertMap(t, redis["info"])
	assertKeys(t, info, "redis_version", "redis_mode", "role", "tcp_port", "uptime", "connected_clients", "blocked_clients", "used_memory_human", "used_memory_rss_human", "maxmemory_human", "mem_fragmentation_ratio", "instantaneous_ops_per_sec", "total_commands_processed", "rejected_connections", "keys_num")
	if info["redis_mode"] == "" {
		t.Fatal("redis monitor redis_mode is empty")
	}
	stats, ok := redis["stats"].([]any)
	if !ok || len(stats) == 0 {
		t.Fatalf("redis monitor stats = %T len %d, want non-empty list", redis["stats"], len(stats))
	}
}

func TestSessionEndpointsAreStatefulAndFilterLikePython(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/monitors/sessions?username=admin", "")
	assertStatusOK(t, resp)
	sessions := assertEnvelopeSlice(t, body)
	if len(sessions) != 1 {
		t.Fatalf("filtered sessions count = %d, want 1", len(sessions))
	}
	session := assertMap(t, sessions[0])
	assertKeys(t, session, "id", "session_uuid", "username", "nickname", "ip", "os", "browser", "device", "status", "last_login_time", "expire_time")
	if session["session_uuid"] != "fixture-session" {
		t.Fatalf("session uuid = %v, want fixture-session", session["session_uuid"])
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/monitors/sessions/1?session_uuid=fixture-session", "")
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/monitors/sessions?username=admin", "")
	assertStatusOK(t, resp)
	sessions = assertEnvelopeSlice(t, body)
	if len(sessions) != 0 {
		t.Fatalf("sessions after delete = %d, want 0", len(sessions))
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

func TestSeedMenusIncludeOfficialPluginMenusAndPermissions(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/menus", "")
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)
	system := findNodeInTree(t, tree, "System")
	systemID := int(system["id"].(float64))

	for _, tc := range []struct {
		name     string
		title    string
		parentID int
		perms    []string
	}{
		{
			name:     "PluginDict",
			title:    "dict.menu",
			parentID: systemID,
			perms:    []string{"dict:type:add", "dict:type:edit", "dict:type:del", "dict:data:add", "dict:data:edit", "dict:data:del"},
		},
		{
			name:     "PluginNotice",
			title:    "notice.menu",
			parentID: systemID,
			perms:    []string{"sys:notice:add", "sys:notice:edit", "sys:notice:del"},
		},
		{
			name:     "Scheduler",
			title:    "page.menu.scheduler",
			parentID: 0,
			perms:    []string{"sys:task:add", "sys:task:edit", "sys:task:del", "sys:task:exec", "sys:task:revoke"},
		},
	} {
		menu := findNodeInTree(t, tree, tc.name)
		if menu["title"] != tc.title {
			t.Fatalf("%s title = %v, want %s", tc.name, menu["title"], tc.title)
		}
		if tc.parentID == 0 {
			if menu["parent_id"] != nil {
				t.Fatalf("%s parent_id = %v, want nil", tc.name, menu["parent_id"])
			}
		} else if menu["parent_id"] != float64(tc.parentID) {
			t.Fatalf("%s parent_id = %v, want %d", tc.name, menu["parent_id"], tc.parentID)
		}
		for _, perm := range tc.perms {
			node := findMenuNodeWithPerm(t, tree, perm)
			if node["perms"] != perm {
				t.Fatalf("permission node %s perms = %v, want %s", perm, node["perms"], perm)
			}
		}
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/roles/1/menus", "")
	assertStatusOK(t, resp)
	roleMenus := assertEnvelopeSlice(t, body)
	for _, name := range []string{"PluginDict", "PluginNotice", "Scheduler", "AddDictType", "AddNotice", "AddScheduler"} {
		assertFlatMenuContains(t, roleMenus, name)
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/auth/codes", "")
	assertStatusOK(t, resp)
	codes := assertEnvelopeSlice(t, body)
	for _, perm := range []string{"dict:type:add", "dict:data:del", "sys:notice:add", "sys:task:add", "sys:task:exec", "sys:task:revoke"} {
		assertStringSliceContains(t, codes, perm)
	}
}

func TestSeedMenusIncludePythonBaseNavigationAndPermissions(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/menus", "")
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)

	for _, tc := range []struct {
		name       string
		title      string
		parentName string
	}{
		{name: "Dashboard", title: "page.dashboard.title"},
		{name: "Analytics", title: "page.dashboard.analytics", parentName: "Dashboard"},
		{name: "Workspace", title: "page.dashboard.workspace", parentName: "Dashboard"},
		{name: "System", title: "page.menu.system"},
		{name: "SysDept", title: "page.menu.sysDept", parentName: "System"},
		{name: "SysUser", title: "page.menu.sysUser", parentName: "System"},
		{name: "SysRole", title: "page.menu.sysRole", parentName: "System"},
		{name: "SysMenu", title: "page.menu.sysMenu", parentName: "System"},
		{name: "SysDataPermission", title: "page.menu.sysDataPermission", parentName: "System"},
		{name: "SysDataScope", title: "page.menu.sysDataScope", parentName: "SysDataPermission"},
		{name: "SysDataRule", title: "page.menu.sysDataRule", parentName: "SysDataPermission"},
		{name: "SysPlugin", title: "page.menu.sysPlugin", parentName: "System"},
		{name: "Log", title: "page.menu.log"},
		{name: "LoginLog", title: "page.menu.login", parentName: "Log"},
		{name: "OperaLog", title: "page.menu.opera", parentName: "Log"},
		{name: "Monitor", title: "page.menu.monitor"},
		{name: "Online", title: "page.menu.online", parentName: "Monitor"},
		{name: "Redis", title: "page.menu.redis", parentName: "Monitor"},
		{name: "Server", title: "page.menu.server", parentName: "Monitor"},
		{name: "Project", title: "项目"},
		{name: "Document", title: "文档", parentName: "Project"},
		{name: "Github", title: "Github", parentName: "Project"},
		{name: "Apifox", title: "Apifox", parentName: "Project"},
		{name: "Profile", title: "page.menu.profile"},
	} {
		menu := findNodeInTree(t, tree, tc.name)
		if menu["title"] != tc.title {
			t.Fatalf("%s title = %v, want %s", tc.name, menu["title"], tc.title)
		}
		if tc.parentName == "" {
			if menu["parent_id"] != nil {
				t.Fatalf("%s parent_id = %v, want nil", tc.name, menu["parent_id"])
			}
			continue
		}
		parent := findNodeInTree(t, tree, tc.parentName)
		if menu["parent_id"] != parent["id"] {
			t.Fatalf("%s parent_id = %v, want %v (%s)", tc.name, menu["parent_id"], parent["id"], tc.parentName)
		}
	}

	for _, perm := range []string{
		"sys:dept:add",
		"sys:dept:edit",
		"sys:dept:del",
		"sys:user:del",
		"sys:role:add",
		"sys:role:edit",
		"sys:role:menu:edit",
		"sys:role:scope:edit",
		"sys:role:del",
		"sys:menu:add",
		"sys:menu:edit",
		"sys:menu:del",
		"data:scope:add",
		"data:scope:edit",
		"data:scope:rule:edit",
		"data:scope:del",
		"data:rule:add",
		"data:rule:edit",
		"data:rule:del",
		"log:login:del",
		"log:login:clear",
		"log:opera:del",
		"log:opera:clear",
	} {
		node := findMenuNodeWithPerm(t, tree, perm)
		if node["perms"] != perm {
			t.Fatalf("permission node %s perms = %v, want %s", perm, node["perms"], perm)
		}
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus/sidebar", "")
	assertStatusOK(t, resp)
	sidebar := assertEnvelopeSlice(t, body)
	profile := findNodeInTree(t, sidebar, "Profile")
	meta := assertMap(t, profile["meta"])
	if meta["hideInMenu"] != true {
		t.Fatalf("Profile hideInMenu = %v, want true", meta["hideInMenu"])
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/auth/codes", "")
	assertStatusOK(t, resp)
	codes := assertEnvelopeSlice(t, body)
	for _, perm := range []string{"sys:dept:add", "sys:role:scope:edit", "log:login:clear", "log:opera:clear"} {
		assertStringSliceContains(t, codes, perm)
	}
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

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus?title=Reports%20Child", "")
	assertStatusOK(t, resp)
	items := assertEnvelopeSlice(t, body)
	if len(items) != 1 {
		t.Fatalf("created child lookup count = %d, want 1", len(items))
	}
	createdLookup := assertMap(t, items[0])
	id := int(createdLookup["id"].(float64))

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/menus/"+itoa(id), "")
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

func TestDeptEndpointsApplyPythonCRUDGuards(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"总部","parent_id":null,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "部门名称已存在")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Missing Parent","parent_id":999999,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "父级部门不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/depts/999999", `{"name":"Missing Dept","parent_id":null,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "部门不存在")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Guard Parent","parent_id":null,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	guardParentID := findDeptID(t, app, "Guard Parent")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Guard Child","parent_id":`+itoa(guardParentID)+`,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	guardChildID := findDeptID(t, app, "Guard Child")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/depts/"+itoa(guardChildID), `{"name":"总部","parent_id":null,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "部门名称已存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/depts/"+itoa(guardChildID), `{"name":"Guard Child","parent_id":999999,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "父级部门不存在")

	resp, body = requestJSON(t, app, "PUT", "/api/v1/sys/depts/"+itoa(guardChildID), `{"name":"Guard Child","parent_id":`+itoa(guardChildID)+`,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertErrorEnvelope(t, resp, body, fiber.StatusForbidden, "禁止关联自身为父级")

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Occupied Dept","parent_id":null,"sort":1,"leader":null,"phone":null,"email":null,"status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)
	occupiedDeptID := findDeptID(t, app, "Occupied Dept")
	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/users", `{"username":"dept_occupied","password":"secret","nickname":"Dept Occupied","email":null,"phone":null,"dept_id":`+itoa(occupiedDeptID)+`,"roles":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeMap(t, body)

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/depts/"+itoa(occupiedDeptID), "")
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "部门下存在用户，无法删除")

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/depts/"+itoa(guardParentID), "")
	assertErrorEnvelope(t, resp, body, fiber.StatusConflict, "部门下存在子部门，无法删除")

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/sys/depts/999999", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "部门不存在")
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
	platformID := int(created["id"].(float64))

	resp, body = requestJSON(t, app, "POST", "/api/v1/sys/depts", `{"name":"Backend","parent_id":`+itoa(platformID)+`,"sort":1,"leader":"Sam","phone":"13700000000","email":"backend@example.com","status":1}`)
	assertStatusOK(t, resp)
	assertEnvelopeNil(t, body)

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/depts", "")
	assertStatusOK(t, resp)
	tree := assertEnvelopeSlice(t, body)
	child := findNodeInTree(t, tree, "Platform")
	if child["parent_id"] != float64(1) {
		t.Fatalf("created child parent_id = %v, want 1", child["parent_id"])
	}
	grandchild := findNodeInTree(t, tree, "Backend")
	if grandchild["parent_id"] != float64(platformID) {
		t.Fatalf("created grandchild parent_id = %v, want %d", grandchild["parent_id"], platformID)
	}
}

func TestDataRuleMetadataEndpointsMatchPythonConfig(t *testing.T) {
	app := newAdminApp(t)

	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/data-rules/models", "")
	assertStatusOK(t, resp)
	models := assertEnvelopeSlice(t, body)
	if len(models) == 0 || models[0] != "__ALL__" {
		t.Fatalf("data-rule models = %v, want __ALL__ first", models)
	}
	if !hasString(models, "user") || !hasString(models, "role") || !hasString(models, "dept") {
		t.Fatalf("data-rule models = %v, want user/role/dept", models)
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules/models/user/columns", "")
	assertStatusOK(t, resp)
	columns := assertEnvelopeSlice(t, body)
	if hasColumn(columns, "id") {
		t.Fatalf("user columns include excluded id column: %v", columns)
	}
	if !hasColumn(columns, "username") || !hasColumn(columns, "dept_id") || !hasColumn(columns, "__dept_id__") || !hasColumn(columns, "__created_by__") {
		t.Fatalf("user columns = %v, want real columns plus template columns", columns)
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules/models/__ALL__/columns", "")
	assertStatusOK(t, resp)
	columns = assertEnvelopeSlice(t, body)
	if len(columns) != 2 || !hasColumn(columns, "__dept_id__") || !hasColumn(columns, "__created_by__") {
		t.Fatalf("__ALL__ columns = %v, want only template columns", columns)
	}

	resp, body = requestJSON(t, app, "GET", "/api/v1/sys/data-rules/value-template-variables", "")
	assertStatusOK(t, resp)
	variables := assertEnvelopeSlice(t, body)
	if !hasColumn(variables, "${user_id}") || !hasColumn(variables, "${dept_id}") || !hasColumn(variables, "${now}") {
		t.Fatalf("value template variables = %v, want Python default variables", variables)
	}

	resp, _ = requestRaw(t, app, "GET", "/api/v1/sys/data-rules/models/missing/columns", "")
	if resp.StatusCode == fiber.StatusOK {
		t.Fatal("missing data-rule model columns returned 200")
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
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	ctx := plugin.NewContext(plugin.ContextOptions{APIGroup: app.Group("/api/v1")})
	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	registerRoutes(ctx.APIGroup(), ctx.Routes())
	return app
}

func newAdminRuntimeApp(t *testing.T) *fiber.App {
	t.Helper()
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{
		Container: di.New(),
		APIGroup:  app.Group("/api/v1"),
	})
	if err := admin.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	plugin.MountRoutes(ctx.APIGroup(), ctx.Routes(), plugin.WithContainer(ctx.Container()))
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

func requestJSONAuth(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, map[string]any) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
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

func requestRaw(t *testing.T, app *fiber.App, method string, path string, body string) (*http.Response, string) {
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
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	return resp, string(raw)
}

func requestRawAuth(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, string) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s error = %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	return resp, string(raw)
}

func loginForAccessToken(t *testing.T, app *fiber.App, username string, password string) string {
	t.Helper()
	resp, body := requestJSON(t, app, "POST", "/api/v1/auth/login", `{"username":"`+username+`","password":"`+password+`","uuid":"fixture-captcha","captcha":"1234"}`)
	assertStatusOK(t, resp)
	data := assertEnvelopeMap(t, body)
	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("access_token = %v, want non-empty string", data["access_token"])
	}
	return token
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

func assertSlice(t *testing.T, value any) []any {
	t.Helper()
	got, ok := value.([]any)
	if !ok {
		t.Fatalf("value = %T, want JSON array", value)
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
	if found, ok := findNodeInTreeValue(t, items, name); ok {
		return found
	}
	t.Fatalf("menu %q not found in tree %v", name, items)
	return nil
}

func findDeptID(t *testing.T, app *fiber.App, name string) int {
	t.Helper()
	escapedName := strings.ReplaceAll(name, " ", "%20")
	resp, body := requestJSON(t, app, "GET", "/api/v1/sys/depts?name="+escapedName, "")
	assertStatusOK(t, resp)
	items := assertEnvelopeSlice(t, body)
	if len(items) != 1 {
		t.Fatalf("dept %q lookup count = %d, want 1", name, len(items))
	}
	item := assertMap(t, items[0])
	return int(item["id"].(float64))
}

func findNodeInTreeValue(t *testing.T, items []any, name string) (map[string]any, bool) {
	t.Helper()
	for _, raw := range items {
		item := assertMap(t, raw)
		if item["name"] == name || item["title"] == name {
			return item, true
		}
		children, _ := item["children"].([]any)
		if len(children) == 0 {
			continue
		}
		if found, ok := findNodeInTreeValue(t, children, name); ok {
			return found, true
		}
	}
	return nil, false
}

func findMenuNodeWithPerm(t *testing.T, items []any, perm string) map[string]any {
	t.Helper()
	if found, ok := findMenuNodeWithPermValue(t, items, perm); ok {
		return found
	}
	t.Fatalf("permission %q not found in menu tree %v", perm, items)
	return nil
}

func findMenuNodeWithPermValue(t *testing.T, items []any, perm string) (map[string]any, bool) {
	t.Helper()
	for _, raw := range items {
		item := assertMap(t, raw)
		if item["perms"] == perm {
			return item, true
		}
		children, _ := item["children"].([]any)
		if len(children) == 0 {
			continue
		}
		if found, ok := findMenuNodeWithPermValue(t, children, perm); ok {
			return found, true
		}
	}
	return nil, false
}

func assertFlatMenuContains(t *testing.T, items []any, name string) {
	t.Helper()
	for _, raw := range items {
		item := assertMap(t, raw)
		if item["name"] == name || item["title"] == name {
			return
		}
	}
	t.Fatalf("flat menu %q not found in %v", name, items)
}

func assertStringSliceContains(t *testing.T, items []any, want string) {
	t.Helper()
	for _, item := range items {
		if item == want {
			return
		}
	}
	t.Fatalf("string %q not found in %v", want, items)
}

func findPluginByName(t *testing.T, items []any, name string) any {
	t.Helper()
	for _, raw := range items {
		item := assertMap(t, raw)
		pluginInfo := assertMap(t, item["plugin"])
		if pluginInfo["name"] == name {
			return raw
		}
	}
	t.Fatalf("plugin %q not found in %v", name, items)
	return nil
}

func hasPluginByName(items []any, name string) bool {
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		pluginInfo, ok := item["plugin"].(map[string]any)
		if ok && pluginInfo["name"] == name {
			return true
		}
	}
	return false
}

func hasString(items []any, want string) bool {
	for _, raw := range items {
		if raw == want {
			return true
		}
	}
	return false
}

func hasColumn(items []any, key string) bool {
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if item["key"] == key {
			return true
		}
	}
	return false
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

func requireCookie(t *testing.T, header string, name string) *http.Cookie {
	t.Helper()
	if header == "" {
		t.Fatalf("missing Set-Cookie header for %s", name)
	}
	parts := strings.Split(header, ";")
	nameValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(nameValue) != 2 || nameValue[0] != name {
		t.Fatalf("Set-Cookie = %q, want %s cookie", header, name)
	}
	return &http.Cookie{Name: nameValue[0], Value: nameValue[1]}
}

func registerRoutes(router fiber.Router, routes []plugin.Route) {
	for _, route := range routes {
		router.Add([]string{route.Method}, route.Path, route.Handler)
	}
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
