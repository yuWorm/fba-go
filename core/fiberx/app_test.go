package fiberx_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/fiberx"
)

func TestNewCreatesAPIGroupWithDefaultBasePath(t *testing.T) {
	fx := fiberx.New(config.Options{})
	fx.API.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	resp, err := fx.App.Test(httptest.NewRequest("GET", "/api/v1/ping", nil))
	if err != nil {
		t.Fatalf("App.Test() error = %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "pong" {
		t.Fatalf("body = %q, want pong", body)
	}
}

func TestRouteKeepsCompatibilityMetadata(t *testing.T) {
	route := fiberx.Route{
		Method:       "DELETE",
		Path:         "/sys/users/:pk",
		Permission:   "sys:user:del",
		AuthRequired: true,
		Summary:      "删除用户",
		Tags:         []string{"系统用户"},
	}

	if route.Permission != "sys:user:del" {
		t.Fatalf("Permission = %q", route.Permission)
	}
	if !route.AuthRequired {
		t.Fatal("AuthRequired = false, want true")
	}
}
