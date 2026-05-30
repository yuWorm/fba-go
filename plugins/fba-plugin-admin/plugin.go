package admin

import (
	"github.com/yuWorm/fba-go/core/plugin"
	adminapi "github.com/yuWorm/fba-plugin-admin/api"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "admin",
		Name:              "Admin Plugin",
		Version:           "0.1.0",
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	handler := adminapi.NewHandler()

	for _, route := range []plugin.Route{
		{Method: "GET", Path: "/auth/captcha", Summary: "Get captcha", Handler: handler.Captcha},
		{Method: "POST", Path: "/auth/login/swagger", Summary: "Swagger login", Handler: handler.LoginSwagger},
		{Method: "POST", Path: "/auth/login", Summary: "Login", Handler: handler.Login},
		{Method: "POST", Path: "/auth/refresh", Summary: "Refresh token", Handler: handler.Refresh},
		{Method: "POST", Path: "/auth/logout", Summary: "Logout", Handler: handler.Logout},
		{Method: "GET", Path: "/auth/codes", Summary: "Permission codes", AuthRequired: true, Handler: handler.Codes},
		{Method: "GET", Path: "/sys/users/me", Summary: "Current user", AuthRequired: true, Handler: handler.CurrentUser},
		{Method: "GET", Path: "/sys/menus/sidebar", Summary: "Sidebar menus", AuthRequired: true, Handler: handler.SidebarMenus},
	} {
		if err := ctx.Route(route); err != nil {
			return err
		}
	}

	return nil
}
