package plugin_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	fbaerrors "github.com/yuWorm/fba-go/core/errors"
	"github.com/yuWorm/fba-go/core/middleware"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/core/rbac"
)

func TestRouteHelpersApplyCommonOptions(t *testing.T) {
	handler := func(c fiber.Ctx) error { return nil }

	route := plugin.GET(
		"/sys/users",
		"List users",
		handler,
		plugin.Auth(),
		plugin.Perm("sys:user:list"),
		plugin.Superuser(),
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
	if !route.SuperuserRequired {
		t.Fatal("SuperuserRequired = false, want true")
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

func TestMountRoutesRejectsProtectedRouteWithoutAuthenticator(t *testing.T) {
	app := fiber.New()
	routes := []plugin.Route{
		plugin.GET("/secure", "Secure", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Auth()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes)

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/secure", nil))
	if err != nil {
		t.Fatalf("GET /secure error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestMountRoutesAuthFailureIncludesTraceID(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestID())
	routes := []plugin.Route{
		plugin.GET("/secure", "Secure", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Auth()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes)

	req := httptest.NewRequest("GET", "/api/v1/secure", nil)
	req.Header.Set(middleware.RequestIDHeader, "request-1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("GET /secure error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if !strings.Contains(string(body), `"trace_id":"request-1"`) {
		t.Fatalf("body = %s, want trace_id request-1", body)
	}
}

func TestMountRoutesAllowsAuthOnlyRouteWithoutRBACRoles(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{ID: 2},
	}
	routes := []plugin.Route{
		plugin.GET("/profile", "Profile", func(c fiber.Ctx) error {
			user, ok := c.Locals(plugin.CurrentUserLocalKey).(*rbac.CurrentUser)
			if !ok || user.ID != 2 {
				t.Fatalf("current user local = %#v, want user 2", c.Locals(plugin.CurrentUserLocalKey))
			}
			return c.SendString("ok")
		}, plugin.Auth()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/profile", nil))
	if err != nil {
		t.Fatalf("GET /profile error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GET /profile status = %d, want 200", resp.StatusCode)
	}
}

func TestMountRoutesPreservesAuthenticatorAppError(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		err: fbaerrors.New(fiber.StatusForbidden, fiber.StatusForbidden, "用户所属部门已被锁定，请联系系统管理员", nil),
	}
	routes := []plugin.Route{
		plugin.GET("/secure", "Secure", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Auth()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/secure", nil))
	if err != nil {
		t.Fatalf("GET /secure error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want 403; body = %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"msg":"用户所属部门已被锁定，请联系系统管理员"`) {
		t.Fatalf("body = %s, want authenticator message", body)
	}
}

func TestMountRoutesAuthorizesWithAuthenticatorAndPermission(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{
			ID:      2,
			IsStaff: true,
			Roles: []rbac.Role{
				{ID: 1, Enabled: true, MenuCount: 1, Permissions: []string{"sys:user:add"}},
			},
		},
	}
	routes := []plugin.Route{
		plugin.POST("/users", "Create user", func(c fiber.Ctx) error {
			user, ok := c.Locals(plugin.CurrentUserLocalKey).(*rbac.CurrentUser)
			if !ok || user.ID != 2 {
				t.Fatalf("current user local = %#v, want user 2", c.Locals(plugin.CurrentUserLocalKey))
			}
			return c.SendString("ok")
		}, plugin.Auth(), plugin.Perm("sys:user:add")),
		plugin.DELETE("/users", "Delete user", func(c fiber.Ctx) error {
			return c.SendString("deleted")
		}, plugin.Auth(), plugin.Perm("sys:user:del")),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("POST", "/api/v1/users", nil))
	if err != nil {
		t.Fatalf("POST /users error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("POST /users status = %d, want 200", resp.StatusCode)
	}

	resp, err = app.Test(httptest.NewRequest("DELETE", "/api/v1/users", nil))
	if err != nil {
		t.Fatalf("DELETE /users error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("DELETE /users status = %d, want 403", resp.StatusCode)
	}
}

func TestMountRoutesReportsPythonStaffGuardMessage(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{
			ID: 2,
			Roles: []rbac.Role{
				{ID: 1, Enabled: true, MenuCount: 1, Permissions: []string{"sys:role:add"}},
			},
		},
	}
	routes := []plugin.Route{
		plugin.POST("/roles", "Create role", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Auth(), plugin.Perm("sys:role:add")),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("POST", "/api/v1/roles", nil))
	if err != nil {
		t.Fatalf("POST /roles error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("POST /roles status = %d, want 403; body = %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"msg":"用户已被禁止后台管理操作，请联系系统管理员"`) {
		t.Fatalf("body = %s, want staff guard message", body)
	}
}

func TestMountRoutesReportsPythonNoEnabledRoleMessage(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{
			ID:      2,
			IsStaff: true,
			Roles: []rbac.Role{
				{ID: 1, Enabled: false, MenuCount: 1, Permissions: []string{"sys:role:view"}},
			},
		},
	}
	routes := []plugin.Route{
		plugin.GET("/roles", "List roles", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Auth(), plugin.Perm("sys:role:view")),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/roles", nil))
	if err != nil {
		t.Fatalf("GET /roles error = %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("GET /roles status = %d, want 403; body = %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), `"msg":"用户所属角色已被锁定，请联系系统管理员"`) {
		t.Fatalf("body = %s, want no enabled role message", body)
	}
}

func TestMountRoutesRejectsSuperuserRouteForNonSuperuser(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{
			ID:      2,
			IsStaff: true,
			Roles: []rbac.Role{
				{ID: 1, Enabled: true},
			},
		},
	}
	routes := []plugin.Route{
		plugin.GET("/plugins", "Plugins", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Superuser()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/plugins", nil))
	if err != nil {
		t.Fatalf("GET /plugins error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("GET /plugins status = %d, want 403", resp.StatusCode)
	}
}

func TestMountRoutesAllowsSuperuserRouteForSuperAdmin(t *testing.T) {
	app := fiber.New()
	authenticator := fakeAuthenticator{
		user: &rbac.CurrentUser{
			ID:           1,
			IsSuperAdmin: true,
		},
	}
	routes := []plugin.Route{
		plugin.GET("/plugins", "Plugins", func(c fiber.Ctx) error {
			return c.SendString("ok")
		}, plugin.Superuser()),
	}

	plugin.MountRoutes(app.Group("/api/v1"), routes, plugin.WithAuthenticator(authenticator))

	resp, err := app.Test(httptest.NewRequest("GET", "/api/v1/plugins", nil))
	if err != nil {
		t.Fatalf("GET /plugins error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GET /plugins status = %d, want 200", resp.StatusCode)
	}
}

type fakeAuthenticator struct {
	user *rbac.CurrentUser
	err  error
}

func (f fakeAuthenticator) Authenticate(fiber.Ctx) (*rbac.CurrentUser, error) {
	return f.user, f.err
}
