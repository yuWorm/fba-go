package admin

import (
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	adminapi "github.com/yuWorm/fba-plugin-admin/api"
	adminmigration "github.com/yuWorm/fba-plugin-admin/migration"
	"github.com/yuWorm/fba-plugin-admin/repo"
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
	repository := repo.Repository(repo.NewMemoryRepository(repo.SeedData()))
	var provider db.Provider
	if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
		repository = repo.NewGORMRepository(provider)
		if err := ctx.Migration(adminmigration.AutoMigrate(provider)); err != nil {
			return err
		}
	}

	handler := adminapi.NewHandlerWithOptions(repository, ctx.Config())
	if err := ctx.Provide(func() plugin.Authenticator {
		return handler
	}); err != nil {
		return err
	}

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
