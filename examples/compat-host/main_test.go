package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
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
		resp, err := app.HTTP().Test(httptest.NewRequest(tc.method, tc.path, nil))
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

	resp, err := app.HTTP().Test(httptest.NewRequest("GET", "/api/v1/dict-datas/type-codes/sys_status", nil))
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
