package dict

import (
	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/core/redisx"
	dictapi "github.com/yuWorm/fba-plugin-dict/api"
	dictmigration "github.com/yuWorm/fba-plugin-dict/migration"
	"github.com/yuWorm/fba-plugin-dict/repo"
	"github.com/yuWorm/fba-plugin-dict/service"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "dict",
		Name:              "Dict Plugin",
		Version:           "0.0.8",
		Description:       "Dictionary data plugin",
		Author:            "wu-clan",
		Tags:              []string{"other"},
		DependsOn:         []plugin.Dependency{{ID: "admin", Optional: true}},
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	repository := repo.Repository(repo.NewMemoryRepository(repo.SeedData()))
	var provider db.Provider
	if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
		repository = repo.NewGORMRepository(provider)
		if err := ctx.Migration(dictmigration.AutoMigrate(provider)); err != nil {
			return err
		}
	}

	invalidator := service.CacheInvalidator(service.NoopInvalidator{})
	var redisClient redisx.RedisClient
	if ctx.Container().Resolve(&redisClient) && redisClient != nil {
		keys := redisx.NewKeys(ctx.Config().Redis.KeyPrefix)
		invalidator = service.NewRedisInvalidator(redisClient, keys.CacheInvalidateChannel(), keys.DictCache())
	}

	handler := dictapi.NewHandler(service.New(repository, invalidator))

	for _, route := range []plugin.Route{
		{Method: "GET", Path: "/dict-types/all", Summary: "Get all dict types", AuthRequired: true, Handler: handler.GetAllDictTypes},
		{Method: "GET", Path: "/dict-types/:pk", Summary: "Get dict type", AuthRequired: true, Handler: handler.GetDictType},
		{Method: "GET", Path: "/dict-types", Summary: "List dict types", AuthRequired: true, Handler: handler.ListDictTypes},
		{Method: "POST", Path: "/dict-types", Summary: "Create dict type", AuthRequired: true, Permission: "dict:type:add", Handler: handler.CreateDictType},
		{Method: "PUT", Path: "/dict-types/:pk", Summary: "Update dict type", AuthRequired: true, Permission: "dict:type:edit", Handler: handler.UpdateDictType},
		{Method: "DELETE", Path: "/dict-types", Summary: "Delete dict types", AuthRequired: true, Permission: "dict:type:del", Handler: handler.DeleteDictTypes},
		{Method: "GET", Path: "/dict-datas/all", Summary: "Get all dict data", AuthRequired: true, Handler: handler.GetAllDictData},
		{Method: "GET", Path: "/dict-datas/:pk", Summary: "Get dict data", AuthRequired: true, Handler: handler.GetDictData},
		{Method: "GET", Path: "/dict-datas/type-codes/:code", Summary: "Get dict data by type code", AuthRequired: true, Handler: handler.GetDictDataByTypeCode},
		{Method: "GET", Path: "/dict-datas", Summary: "List dict data", AuthRequired: true, Handler: handler.ListDictData},
		{Method: "POST", Path: "/dict-datas", Summary: "Create dict data", AuthRequired: true, Permission: "dict:data:add", Handler: handler.CreateDictData},
		{Method: "PUT", Path: "/dict-datas/:pk", Summary: "Update dict data", AuthRequired: true, Permission: "dict:data:edit", Handler: handler.UpdateDictData},
		{Method: "DELETE", Path: "/dict-datas", Summary: "Delete dict data", AuthRequired: true, Permission: "dict:data:del", Handler: handler.DeleteDictData},
	} {
		if err := ctx.Route(route); err != nil {
			return err
		}
	}

	return nil
}
