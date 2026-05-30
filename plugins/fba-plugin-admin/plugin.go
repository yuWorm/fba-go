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

	return plugin.RegisterRoutes(ctx,
		adminapi.AuthRoutes(handler),
		adminapi.UserRoutes(handler),
		adminapi.RoleRoutes(handler),
		adminapi.MenuRoutes(handler),
		adminapi.DeptRoutes(handler),
		adminapi.DataRuleRoutes(handler),
		adminapi.DataScopeRoutes(handler),
		adminapi.FileRoutes(handler),
		adminapi.PluginRoutes(handler),
		adminapi.LogRoutes(handler),
		adminapi.MonitorRoutes(handler),
	)
}
