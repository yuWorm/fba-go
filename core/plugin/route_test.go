package plugin_test

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/plugin"
)

func TestRouteHelpersApplyCommonOptions(t *testing.T) {
	handler := func(c fiber.Ctx) error { return nil }

	route := plugin.GET(
		"/sys/users",
		"List users",
		handler,
		plugin.Auth(),
		plugin.Perm("sys:user:list"),
		plugin.Tags("sys", "users"),
	)

	if route.Method != "GET" {
		t.Fatalf("Method = %q, want GET", route.Method)
	}
	if route.Path != "/sys/users" {
		t.Fatalf("Path = %q, want /sys/users", route.Path)
	}
	if route.Summary != "List users" {
		t.Fatalf("Summary = %q, want List users", route.Summary)
	}
	if !route.AuthRequired {
		t.Fatal("AuthRequired = false, want true")
	}
	if route.Permission != "sys:user:list" {
		t.Fatalf("Permission = %q, want sys:user:list", route.Permission)
	}
	if len(route.Tags) != 2 || route.Tags[0] != "sys" || route.Tags[1] != "users" {
		t.Fatalf("Tags = %v, want [sys users]", route.Tags)
	}
	if route.Handler == nil {
		t.Fatal("Handler is nil")
	}
}

func TestRegisterRoutesRegistersGroupsInOrder(t *testing.T) {
	handler := func(c fiber.Ctx) error { return nil }
	ctx := plugin.NewContext(plugin.ContextOptions{})

	err := plugin.RegisterRoutes(ctx,
		[]plugin.Route{
			plugin.GET("/a", "A", handler),
			plugin.POST("/b", "B", handler),
		},
		[]plugin.Route{
			plugin.PUT("/c", "C", handler),
			plugin.DELETE("/d", "D", handler),
		},
	)
	if err != nil {
		t.Fatalf("RegisterRoutes() error = %v", err)
	}

	routes := ctx.Routes()
	if len(routes) != 4 {
		t.Fatalf("routes = %d, want 4", len(routes))
	}
	for index, want := range []string{"GET /a", "POST /b", "PUT /c", "DELETE /d"} {
		got := routes[index].Method + " " + routes[index].Path
		if got != want {
			t.Fatalf("route[%d] = %s, want %s", index, got, want)
		}
	}
}
