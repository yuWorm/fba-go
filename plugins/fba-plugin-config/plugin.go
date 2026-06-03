package config

import (
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	configapi "github.com/yuWorm/fba-plugin-config/api"
	configmigration "github.com/yuWorm/fba-plugin-config/migration"
	"github.com/yuWorm/fba-plugin-config/repo"
	"github.com/yuWorm/fba-plugin-config/service"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "config",
		Name:              "Config Plugin",
		Version:           "0.0.2",
		Description:       "System config plugin",
		Author:            "wu-clan",
		Tags:              []string{"other"},
		DependsOn:         []plugin.Dependency{{ID: "admin", Optional: true}},
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	repository := repo.Repository(repo.NewMemoryRepository(repo.SeedData()))
	if ctx.Container().Resolve(&repository) && repository != nil {
		// Host applications may provide a custom repository for tests or tenancy;
		// keep it authoritative instead of replacing it with the DB-backed default.
	} else {
		var provider db.Provider
		if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
			repository = repo.NewGORMRepository(provider)
			if err := ctx.Migration(configmigration.AutoMigrate(provider)); err != nil {
				return err
			}
		}
	}

	if err := ctx.Container().Provide(func() adminservice.AdminConfigProvider {
		return service.NewAdminConfigProvider(repository)
	}); err != nil {
		return err
	}

	handler := configapi.NewHandler(service.New(repository))
	return plugin.RegisterRoutes(ctx, configapi.Routes(handler))
}
