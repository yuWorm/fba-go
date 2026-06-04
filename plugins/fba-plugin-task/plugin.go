package task

import (
	"os"
	"time"

	"github.com/yuWorm/fba-go/core/db"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/core/realtime"
	"github.com/yuWorm/fba-go/core/redisx"
	coretask "github.com/yuWorm/fba-go/core/task"
	taskapi "github.com/yuWorm/fba-plugin-task/api"
	taskmigration "github.com/yuWorm/fba-plugin-task/migration"
	"github.com/yuWorm/fba-plugin-task/repo"
	"github.com/yuWorm/fba-plugin-task/service"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "task",
		Name:              "Task Plugin",
		Version:           "0.1.0",
		Description:       "Task scheduler compatibility plugin",
		DependsOn:         []plugin.Dependency{{ID: "admin", Optional: true}},
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	var registry coretask.DefinitionRegistry
	_ = ctx.Container().Resolve(&registry)

	repository := repo.Repository(repo.NewMemoryRepository(repo.SeedData()))
	var provider db.Provider
	if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
		repository = repo.NewGORMRepository(provider)
		if err := ctx.Migration(taskmigration.AutoMigrate(provider)); err != nil {
			return err
		}
	}

	executor := coretask.Runtime(coretask.NoopRuntime{})
	_ = ctx.Container().Resolve(&executor)
	var hub realtime.Hub
	_ = ctx.Container().Resolve(&hub)

	leader := service.LeaderLease(service.NoopLeaderLease{})
	var redisClient redisx.RedisClient
	if ctx.Container().Resolve(&redisClient) && redisClient != nil {
		nodeID, _ := os.Hostname()
		if nodeID == "" {
			nodeID = "fba-go"
		}
		ttl := ctx.Config().Task.SchedulerLockTTL
		if ttl <= 0 {
			ttl = 30 * time.Second
		}
		leader = service.NewRedisLeaderLease(redisClient, redisx.NewKeys(ctx.Config().Redis.KeyPrefix).SchedulerLeader(), nodeID, ttl)
	}

	svc := service.New(repository, registry, executor, leader, service.WithRealtimeHub(hub))
	svc.RegisterRealtimeHandlers(hub)
	handler := taskapi.NewHandler(svc)
	return plugin.RegisterRoutes(ctx, taskapi.Routes(handler))
}
