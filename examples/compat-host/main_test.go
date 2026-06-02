package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/db"
	adminmodel "github.com/yuWorm/fba-plugin-admin/model"
	dictmodel "github.com/yuWorm/fba-plugin-dict/model"
	noticemodel "github.com/yuWorm/fba-plugin-notice/model"
	taskmodel "github.com/yuWorm/fba-plugin-task/model"
)

func TestCompatHostRegistersFixturePlugin(t *testing.T) {
	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}

	resp, err := app.HTTP().Test(httptest.NewRequest("GET", "/api/v1/orders", nil))
	if err != nil {
		t.Fatalf("GET /api/v1/orders error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "orders fixture" {
		t.Fatalf("body = %q, want orders fixture", body)
	}
}

func TestCompatHostRegistersOfficialPlugins(t *testing.T) {
	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}
	token := compatAccessToken(t, app.HTTP())

	for _, tc := range []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/auth/captcha"},
		{"POST", "/api/v1/auth/login/swagger"},
		{"GET", "/api/v1/dict-datas/type-codes/sys_status"},
		{"GET", "/api/v1/tasks/registered"},
		{"GET", "/api/v1/schedulers"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.HTTP().Test(req)
		if err != nil {
			t.Fatalf("%s %s error = %v", tc.method, tc.path, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("%s %s status = %d body = %s", tc.method, tc.path, resp.StatusCode, body)
		}
	}
}

func TestCompatHostOfficialPluginPriorityResponses(t *testing.T) {
	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}
	token := compatAccessToken(t, app.HTTP())

	req := httptest.NewRequest("GET", "/api/v1/dict-datas/type-codes/sys_status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.HTTP().Test(req)
	if err != nil {
		t.Fatalf("GET /dict-datas/type-codes/sys_status error = %v", err)
	}
	defer resp.Body.Close()

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if body["code"] != float64(200) {
		t.Fatalf("code = %v, want 200", body["code"])
	}
	data, ok := body["data"].([]any)
	if !ok || len(data) != 2 {
		t.Fatalf("data = %T len %d, want 2 dict rows", body["data"], len(data))
	}
}

func TestCompatHostDictSeedsPythonMainOptions(t *testing.T) {
	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}
	token := compatAccessToken(t, app.HTTP())

	assertCompatDictDataLen(t, app.HTTP(), token, "task_period_type", 5)
	assertCompatDictDataLen(t, app.HTTP(), token, "notice", 2)
}

func TestCompatHostSQLiteModeRunsPluginMigrationsAndSeeds(t *testing.T) {
	t.Setenv("FBA_COMPAT_DB", "sqlite")
	t.Setenv("FBA_COMPAT_SQLITE_DSN", compatSQLiteTestDSN(t))

	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}

	var provider db.Provider
	if ok := app.Container().Resolve(&provider); !ok {
		t.Fatalf("db.Provider was not registered in sqlite compat mode")
	}

	adminSeed := adminmodel.SeedData()
	assertCompatTableCount(t, provider, &adminmodel.User{}, len(adminSeed.Users))
	assertCompatTableCount(t, provider, &adminmodel.Role{}, len(adminSeed.Roles))
	assertCompatTableCount(t, provider, &adminmodel.Plugin{}, len(adminSeed.Plugins))
	assertCompatTableCount(t, provider, &dictmodel.DictType{}, len(dictmodel.SeedDictTypes()))
	assertCompatTableCount(t, provider, &dictmodel.DictData{}, len(dictmodel.SeedDictData()))
	assertCompatTableCount(t, provider, &noticemodel.Notice{}, len(noticemodel.SeedNotices()))
	assertCompatTableCount(t, provider, &taskmodel.TaskScheduler{}, len(taskmodel.SeedSchedulers()))
	assertCompatTableCount(t, provider, &taskmodel.TaskResult{}, len(taskmodel.SeedTaskResults()))

	token := compatAccessToken(t, app.HTTP())
	resp, body := compatRequestJSON(t, app.HTTP(), "GET", "/api/v1/schedulers", "", token)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GET /schedulers status = %d body = %v, want 200", resp.StatusCode, body)
	}
}

func TestCompatHostEnforcesBusinessPluginPermissions(t *testing.T) {
	app, err := newApplication()
	if err != nil {
		t.Fatalf("newApplication() error = %v", err)
	}
	adminToken := compatAccessToken(t, app.HTTP())

	resp, _ := compatRequestRaw(t, app.HTTP(), "GET", "/api/v1/dict-datas/type-codes/sys_status", "", "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("GET /dict-datas/type-codes/sys_status without token status = %d, want 401", resp.StatusCode)
	}

	viewerToken := createCompatUserToken(t, app.HTTP(), adminToken, "viewer", 2)
	resp, _ = compatRequestJSON(t, app.HTTP(), "GET", "/api/v1/dict-types/all", "", viewerToken)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("viewer GET /dict-types/all status = %d, want 200", resp.StatusCode)
	}
	resp, _ = compatRequestRaw(t, app.HTTP(), "POST", "/api/v1/dict-types", `{"name":"Viewer Type","code":"viewer_type","remark":null}`, viewerToken)
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("viewer POST /dict-types status = %d, want 403", resp.StatusCode)
	}

	writerToken := createCompatUserToken(t, app.HTTP(), adminToken, "writer", 1)
	resp, body := compatRequestJSON(t, app.HTTP(), "POST", "/api/v1/dict-types", `{"name":"Writer Type","code":"writer_type","remark":null}`, writerToken)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("writer POST /dict-types status = %d body = %v, want 200", resp.StatusCode, body)
	}
}

func compatAccessToken(t *testing.T, app *fiber.App) string {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"admin","uuid":"fixture-captcha","captcha":"1234"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("POST /auth/login error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST /auth/login status = %d body = %s", resp.StatusCode, body)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode(login) error = %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("login data = %T, want object", payload["data"])
	}
	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("access_token = %v, want non-empty string", data["access_token"])
	}
	return token
}

func createCompatUserToken(t *testing.T, app *fiber.App, adminToken string, username string, roleID int) string {
	t.Helper()
	body := `{"username":"` + username + `","password":"secret","nickname":"` + username + `","email":null,"phone":null,"dept_id":1,"roles":[` + itoa(roleID) + `]}`
	resp, payload := compatRequestJSON(t, app, "POST", "/api/v1/sys/users", body, adminToken)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("POST /sys/users status = %d body = %v, want 200", resp.StatusCode, payload)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("create user data = %T, want object", payload["data"])
	}
	id, ok := data["id"].(float64)
	if !ok {
		t.Fatalf("created user id = %v, want number", data["id"])
	}
	resp, payload = compatRequestJSON(t, app, "PUT", "/api/v1/sys/users/"+itoa(int(id))+"/permissions?type=staff", "", adminToken)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("PUT /sys/users/%d/permissions status = %d body = %v, want 200", int(id), resp.StatusCode, payload)
	}
	return compatAccessTokenForUser(t, app, username, "secret")
}

func compatAccessTokenForUser(t *testing.T, app *fiber.App, username string, password string) string {
	t.Helper()
	resp, payload := compatRequestJSON(t, app, "POST", "/api/v1/auth/login", `{"username":"`+username+`","password":"`+password+`","uuid":"fixture-captcha","captcha":"1234"}`, "")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("POST /auth/login status = %d body = %v, want 200", resp.StatusCode, payload)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("login data = %T, want object", payload["data"])
	}
	token, ok := data["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("access_token = %v, want non-empty string", data["access_token"])
	}
	return token
}

func compatRequestJSON(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, map[string]any) {
	t.Helper()
	resp, raw := compatRequestRaw(t, app, method, path, body, token)
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode %s %s response: %v", method, path, err)
	}
	return resp, payload
}

func compatRequestRaw(t *testing.T, app *fiber.App, method string, path string, body string, token string) (*http.Response, string) {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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

func itoa(value int) string {
	return strconv.Itoa(value)
}

func compatSQLiteTestDSN(t *testing.T) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return "file:" + name + "?mode=memory&cache=shared"
}

func assertCompatTableCount(t *testing.T, provider db.Provider, table any, want int) {
	t.Helper()
	var got int64
	if err := provider.Read().WithContext(context.Background()).Model(table).Count(&got).Error; err != nil {
		t.Fatalf("count %T error = %v", table, err)
	}
	if got != int64(want) {
		t.Fatalf("count %T = %d, want %d", table, got, want)
	}
}

func assertCompatDictDataLen(t *testing.T, app *fiber.App, token string, code string, want int) {
	t.Helper()
	resp, body := compatRequestJSON(t, app, "GET", "/api/v1/dict-datas/type-codes/"+code, "", token)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GET /dict-datas/type-codes/%s status = %d body = %v, want 200", code, resp.StatusCode, body)
	}
	data, ok := body["data"].([]any)
	if !ok {
		t.Fatalf("GET /dict-datas/type-codes/%s data = %T, want array", code, body["data"])
	}
	if len(data) != want {
		t.Fatalf("GET /dict-datas/type-codes/%s len = %d, want %d", code, len(data), want)
	}
}
