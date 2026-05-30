package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
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
