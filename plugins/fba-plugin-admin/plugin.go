package admin

import (
	"context"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	adminapi "github.com/yuWorm/fba-plugin-admin/api"
	adminmigration "github.com/yuWorm/fba-plugin-admin/migration"
	"github.com/yuWorm/fba-plugin-admin/repo"
	"github.com/yuWorm/fba-plugin-admin/service"
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
	var injectedRepository repo.Repository
	var provider db.Provider
	if ctx.Container().Resolve(&injectedRepository) && injectedRepository != nil {
		repository = injectedRepository
	} else if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
		repository = repo.NewGORMRepository(provider)
		if err := ctx.Migration(adminmigration.AutoMigrate(provider)); err != nil {
			return err
		}
		if err := ctx.Migration(adminmigration.PasswordSecurityMigration(provider)); err != nil {
			return err
		}
	}
	var redisClient service.RedisClient
	_ = ctx.Container().Resolve(&redisClient)
	if err := ctx.Provide(func() repo.Repository {
		return repository
	}); err != nil {
		return err
	}

	handler := adminapi.NewHandlerWithAdminOptions(repository, adminapi.HandlerOptions{
		Config:         ctx.Config(),
		ConfigProvider: deferredAdminConfigProvider{resolver: ctx.Container()},
		Redis:          redisClient,
	})
	if err := ctx.Provide(func() plugin.Authenticator {
		return handler
	}); err != nil {
		return err
	}

	apiBasePath := ctx.Config().App.APIBasePath
	return plugin.RegisterRoutes(ctx,
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.AuthRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.UserRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.RoleRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.MenuRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.DeptRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.DataRuleRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.DataScopeRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.FileRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.PluginRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.LogRoutes(handler)),
		adminapi.WithOperationLogging(handler, apiBasePath, adminapi.MonitorRoutes(handler)),
	)
}

type adminConfigProviderResolver interface {
	Resolve(any) bool
}

type deferredAdminConfigProvider struct {
	resolver adminConfigProviderResolver
}

func (p deferredAdminConfigProvider) LoginConfig(ctx context.Context) (service.LoginConfig, error) {
	return p.current().LoginConfig(ctx)
}

func (p deferredAdminConfigProvider) UserSecurityConfig(ctx context.Context) (service.UserSecurityConfig, error) {
	return p.current().UserSecurityConfig(ctx)
}

func (p deferredAdminConfigProvider) current() service.AdminConfigProvider {
	var provider service.AdminConfigProvider
	if p.resolver != nil && p.resolver.Resolve(&provider) && provider != nil {
		return provider
	}
	// The config plugin may register after admin because it extends admin routes.
	// Until then, admin must keep the same static defaults as the Python runtime.
	return service.DefaultAdminConfigProvider{}
}
