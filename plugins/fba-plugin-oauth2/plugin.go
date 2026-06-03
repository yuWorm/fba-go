package oauth2

import (
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	adminrepo "github.com/yuWorm/fba-plugin-admin/repo"
	adminservice "github.com/yuWorm/fba-plugin-admin/service"
	oauth2api "github.com/yuWorm/fba-plugin-oauth2/api"
	oauth2migration "github.com/yuWorm/fba-plugin-oauth2/migration"
	"github.com/yuWorm/fba-plugin-oauth2/repo"
	"github.com/yuWorm/fba-plugin-oauth2/service"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "oauth2",
		Name:              "OAuth2 Plugin",
		Version:           "0.0.11",
		Description:       "支持 GitHub、Google 等社交平台登录",
		Author:            "wu-clan",
		Tags:              []string{"auth"},
		DependsOn:         []plugin.Dependency{{ID: "admin", Optional: true}},
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	settings := service.DefaultSettings()
	repository := repo.Repository(repo.NewMemoryRepository(repo.SeedData()))
	var injectedRepository repo.Repository
	if ctx.Container().Resolve(&injectedRepository) && injectedRepository != nil {
		repository = injectedRepository
	} else {
		var provider db.Provider
		if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
			repository = repo.NewGORMRepository(provider)
			if err := ctx.Migration(oauth2migration.AutoMigrate(provider)); err != nil {
				return err
			}
		}
	}

	adminRepository := adminrepo.Repository(adminrepo.NewMemoryRepository(adminrepo.SeedData()))
	var injectedAdminRepository adminrepo.Repository
	if ctx.Container().Resolve(&injectedAdminRepository) && injectedAdminRepository != nil {
		adminRepository = injectedAdminRepository
	}
	var redisClient adminservice.RedisClient
	_ = ctx.Container().Resolve(&redisClient)
	stateStore := service.StateStore(service.NewMemoryStateStore())
	var injectedStateStore service.StateStore
	if ctx.Container().Resolve(&injectedStateStore) && injectedStateStore != nil {
		stateStore = injectedStateStore
	} else if redisClient != nil {
		stateStore = service.NewRedisStateStore(redisClient, settings.StateRedisPrefix)
	}

	svc := service.New(service.Options{
		Repository: repository,
		AdminRepo:  adminRepository,
		StateStore: stateStore,
		Settings:   settings,
	})
	return plugin.RegisterRoutes(ctx, oauth2api.Routes(oauth2api.NewHandler(svc)))
}
