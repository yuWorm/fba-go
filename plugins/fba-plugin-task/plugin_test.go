package task_test

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
	coretask "github.com/yuWorm/fba-go/core/task"
	task "github.com/yuWorm/fba-plugin-task"
	"gorm.io/gorm"
)

func TestTaskPluginRegistersPythonCompatibleRoutes(t *testing.T) {
	ctx := plugin.NewContext(plugin.ContextOptions{})

	if err := task.FBAPlugin().Register(ctx); err != nil {
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
		"GET /tasks/registered":         {authRequired: true},
		"DELETE /tasks/:task_id/cancel": {authRequired: true, permission: "sys:task:revoke"},
		"GET /task-results/:pk":         {authRequired: true},
		"GET /task-results":             {authRequired: true},
		"DELETE /task-results":          {authRequired: true, permission: "sys:task:del"},
		"GET /schedulers/all":           {authRequired: true},
		"GET /schedulers/:pk":           {authRequired: true},
		"GET /schedulers":               {authRequired: true},
		"POST /schedulers":              {authRequired: true, permission: "sys:task:add"},
		"PUT /schedulers/:pk":           {authRequired: true, permission: "sys:task:edit"},
		"PUT /schedulers/:pk/status":    {authRequired: true, permission: "sys:task:edit"},
		"DELETE /schedulers/:pk":        {authRequired: true, permission: "sys:task:del"},
		"POST /schedulers/:pk/execute":  {authRequired: true, permission: "sys:task:exec"},
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

func TestRegisteredTasksUseCoreRegistryAndMatchPythonSchema(t *testing.T) {
	registry := coretask.NewRegistry()
	if err := registry.Add(coretask.Definition{Type: "task_demo", Name: "任务演示", Queue: "default"}); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	app := newTaskApp(t, registry)

	resp, body := requestJSON(t, app, "GET", "/api/v1/tasks/registered", "")

	assertStatusOK(t, resp)
	data := assertEnvelopeSlice(t, body)
	if len(data) != 1 {
		t.Fatalf("registered task count = %d, want 1", len(data))
	}
	item := assertMap(t, data[0])
	assertKeys(t, item, "name", "task")
	if item["task"] != "task_demo" {
		t.Fatalf("task = %v, want task_demo", item["task"])
	}
}

func TestTaskPluginConsumesCoreRuntimeContract(t *testing.T) {
	runtime := &recordingRuntime{}
	container := di.New()
	if err := container.Provide(func() coretask.Runtime { return runtime }); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	app := newTaskAppWithContainer(t, container)

	resp, body := requestJSON(t, app, "POST", "/api/v1/schedulers/1/execute", "")
	assertStatusOK(t, resp)
	assertEnvelopeNull(t, body)
	if runtime.executedTask != "task_demo" {
		t.Fatalf("executed task = %q, want task_demo", runtime.executedTask)
	}

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/tasks/task-123/cancel", "")
	assertStatusOK(t, resp)
	assertEnvelopeNull(t, body)
	if runtime.canceledTaskID != "task-123" {
		t.Fatalf("canceled task ID = %q, want task-123", runtime.canceledTaskID)
	}
}

func TestTaskPluginRegistersMigrationWhenDBProviderExists(t *testing.T) {
	container := di.New()
	if err := container.Provide(func() db.Provider {
		return db.NewGORMProvider(&gorm.DB{}, nil)
	}); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	ctx := plugin.NewContext(plugin.ContextOptions{Container: container})

	if err := task.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	migrations := ctx.Migrations()
	if len(migrations) != 1 {
		t.Fatalf("migrations = %d, want 1", len(migrations))
	}
	if migrations[0].Scope != "plugin:task" {
		t.Fatalf("migration scope = %q, want plugin:task", migrations[0].Scope)
	}
}

func TestSchedulersMatchPythonSchemas(t *testing.T) {
	app := newTaskApp(t, nil)

	resp, body := requestJSON(t, app, "GET", "/api/v1/schedulers/all", "")
	assertStatusOK(t, resp)
	all := assertEnvelopeSlice(t, body)
	if len(all) == 0 {
		t.Fatal("schedulers/all data is empty")
	}
	assertSchedulerDetail(t, assertMap(t, all[0]))

	resp, body = requestJSON(t, app, "GET", "/api/v1/schedulers", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	assertPage(t, page, assertSchedulerDetail)

	resp, body = requestJSON(t, app, "GET", "/api/v1/schedulers/1", "")
	assertStatusOK(t, resp)
	assertSchedulerDetail(t, assertEnvelopeMap(t, body))
}

func TestSchedulerWriteEndpointsReturnPythonEnvelope(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/api/v1/schedulers", schedulerBody()},
		{"PUT", "/api/v1/schedulers/1", schedulerBody()},
		{"PUT", "/api/v1/schedulers/1/status", ""},
		{"DELETE", "/api/v1/schedulers/1", ""},
		{"POST", "/api/v1/schedulers/1/execute", ""},
	} {
		app := newTaskApp(t, nil)
		resp, body := requestJSON(t, app, tc.method, tc.path, tc.body)
		assertStatusOK(t, resp)
		assertEnvelopeNull(t, body)
	}
}

func TestTaskResultsMatchPythonSchemas(t *testing.T) {
	app := newTaskApp(t, nil)

	resp, body := requestJSON(t, app, "GET", "/api/v1/task-results", "")
	assertStatusOK(t, resp)
	page := assertEnvelopeMap(t, body)
	assertPage(t, page, assertTaskResultDetail)

	resp, body = requestJSON(t, app, "GET", "/api/v1/task-results/1", "")
	assertStatusOK(t, resp)
	result := assertEnvelopeMap(t, body)
	assertTaskResultDetail(t, result)
	if result["status"] != "STARTED" {
		t.Fatalf("status = %v, want STARTED", result["status"])
	}
}

func TestTaskControlWriteEndpointsReturnPythonEnvelope(t *testing.T) {
	app := newTaskApp(t, nil)

	resp, body := requestJSON(t, app, "DELETE", "/api/v1/tasks/task-1/cancel", "")
	assertStatusOK(t, resp)
	assertEnvelopeNull(t, body)

	resp, body = requestJSON(t, app, "DELETE", "/api/v1/task-results", `{"pks":[1]}`)
	assertStatusOK(t, resp)
	assertEnvelopeNull(t, body)
}

func TestTaskMissingMutationsMatchPython(t *testing.T) {
	app := newTaskApp(t, nil)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{"PUT", "/api/v1/schedulers/999999", schedulerBody()},
		{"PUT", "/api/v1/schedulers/999999/status", ""},
		{"DELETE", "/api/v1/schedulers/999999", ""},
	} {
		resp, body := requestJSON(t, app, tc.method, tc.path, tc.body)
		assertErrorEnvelope(t, resp, body, fiber.StatusNotFound, "任务调度不存在")
	}

	resp, body := requestJSON(t, app, "DELETE", "/api/v1/task-results", `{"pks":[999999]}`)
	assertStatusOK(t, resp)
	assertBusinessFailEnvelope(t, body)
}

func TestTaskValidationErrorsMatchPython(t *testing.T) {
	app := newTaskApp(t, nil)

	resp, body := requestJSON(t, app, "GET", "/api/v1/schedulers/not-int", "")
	assertErrorEnvelope(t, resp, body, fiber.StatusUnprocessableEntity, "请求参数非法: pk 输入应为有效的整数，无法将字符串解析为整数，输入：not-int")
}

func newTaskApp(t *testing.T, registry *coretask.Registry) *fiber.App {
	t.Helper()
	container := di.New()
	if registry != nil {
		if err := container.Provide(func() *coretask.Registry { return registry }); err != nil {
			t.Fatalf("Provide() error = %v", err)
		}
	}
	return newTaskAppWithContainer(t, container)
}

func newTaskAppWithContainer(t *testing.T, container *di.Container) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	ctx := plugin.NewContext(plugin.ContextOptions{
		Container: container,
		APIGroup:  app.Group("/api/v1"),
	})
	if err := task.FBAPlugin().Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	for _, route := range ctx.Routes() {
		ctx.APIGroup().Add([]string{route.Method}, route.Path, route.Handler)
	}
	return app
}

func schedulerBody() string {
	return `{"name":"Fixture","task":"task_demo","args":null,"kwargs":null,"queue":null,"exchange":null,"routing_key":null,"start_time":null,"expire_time":null,"expire_seconds":null,"type":0,"interval_every":10,"interval_period":"seconds","crontab":"* * * * *","one_off":false,"remark":null}`
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

func assertBusinessFailEnvelope(t *testing.T, body map[string]any) {
	t.Helper()
	if body["code"] != float64(400) {
		t.Fatalf("code = %v, want 400; body = %v", body["code"], body)
	}
	if body["msg"] != "请求错误" {
		t.Fatalf("msg = %v, want 请求错误", body["msg"])
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

func assertPage(t *testing.T, page map[string]any, itemAssert func(*testing.T, map[string]any)) {
	t.Helper()
	assertKeys(t, page, "items", "total", "page", "size", "total_pages", "links")
	items, ok := page["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("items = %T len %d, want non-empty JSON array", page["items"], len(items))
	}
	itemAssert(t, assertMap(t, items[0]))
	links := assertMap(t, page["links"])
	assertKeys(t, links, "first", "last", "self", "next", "prev")
}

func assertSchedulerDetail(t *testing.T, data map[string]any) {
	t.Helper()
	assertKeys(t, data,
		"name", "task", "args", "kwargs", "queue", "exchange", "routing_key",
		"start_time", "expire_time", "expire_seconds", "type", "interval_every",
		"interval_period", "crontab", "one_off", "remark", "id", "enabled",
		"total_run_count", "last_run_time", "created_time", "updated_time",
	)
}

func assertTaskResultDetail(t *testing.T, data map[string]any) {
	t.Helper()
	assertKeys(t, data,
		"task_id", "status", "result", "date_done", "traceback", "name",
		"args", "kwargs", "worker", "retries", "queue", "id",
	)
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

type recordingRuntime struct {
	reloadCalls    int
	executedTask   string
	canceledTaskID string
}

func (r *recordingRuntime) Reload(context.Context) error {
	r.reloadCalls++
	return nil
}

func (r *recordingRuntime) Execute(_ context.Context, task string, _ any, _ any) error {
	r.executedTask = task
	return nil
}

func (r *recordingRuntime) Cancel(_ context.Context, taskID string) error {
	r.canceledTaskID = taskID
	return nil
}
