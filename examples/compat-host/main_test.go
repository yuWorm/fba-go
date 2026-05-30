package main

import (
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
