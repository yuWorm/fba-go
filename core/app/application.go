package app

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/fiberx"
	"github.com/yuWorm/fba-go/core/observability"
	"github.com/yuWorm/fba-go/core/realtime"
	"github.com/yuWorm/fba-go/core/redisx"
)

type Application interface {
	HTTP() *fiber.App
	Container() *di.Container
	Run(ctx context.Context) error
	RunHTTP(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type application struct {
	container *di.Container
	http      *fiber.App
	opts      config.Options
}

func New(opts config.Options) (Application, error) {
	opts = opts.WithDefaults()
	fx := fiberx.New(opts)
	observability.RegisterCoreRoutes(fx.App, observability.NewReadiness())
	container := di.New()
	onlineStore := realtime.OnlineStore(realtime.NewMemoryOnlineStore())
	hubOptions := []realtime.SocketIOHubOption{}
	if opts.Realtime.MultiInstance.Enabled {
		redisClient := redisx.RedisClient(redisx.NewUniversalClient(opts.Redis))
		keys := redisx.NewKeys(opts.Redis.KeyPrefix)
		onlineStore = realtime.NewRedisOnlineStore(redisClient, keys)
		nodeID := opts.Realtime.MultiInstance.NodeID
		if nodeID != "" {
			hubOptions = append(hubOptions, realtime.WithNodeID(nodeID))
		}
		hubOptions = append(hubOptions, realtime.WithBroadcaster(realtime.NewRedisBroadcaster(redisClient, opts.Realtime.MultiInstance.Channel)))
		if err := container.Provide(func() redisx.RedisClient { return redisClient }); err != nil {
			return nil, err
		}
	}
	hub := realtime.NewSocketIOHub(onlineStore, hubOptions...)
	if err := container.Provide(func() realtime.Hub { return hub }); err != nil {
		return nil, err
	}
	if err := container.Provide(func() realtime.OnlineStore { return onlineStore }); err != nil {
		return nil, err
	}
	if !opts.Realtime.Disabled {
		server := realtime.NewSocketIOServer(hub, realtime.SocketIOServerOptions{
			Config:        opts,
			Authenticator: realtime.NewJWTAuthenticator(nil, opts),
		})
		server.Mount(fx.App)
		if opts.Realtime.MultiInstance.Enabled {
			opts.Hooks.OnStart = append(opts.Hooks.OnStart, hub.StartBroadcaster)
			opts.Hooks.OnShutdown = append(opts.Hooks.OnShutdown, hub.ShutdownBroadcaster)
		}
		opts.Hooks.OnShutdown = append(opts.Hooks.OnShutdown, realtime.Shutdown)
	}
	return &application{
		container: container,
		http:      fx.App,
		opts:      opts,
	}, nil
}

func (a *application) HTTP() *fiber.App {
	return a.http
}

func (a *application) Container() *di.Container {
	return a.container
}

func (a *application) Run(ctx context.Context) error {
	return a.RunHTTP(ctx)
}

func (a *application) RunHTTP(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, hook := range a.opts.Hooks.OnStart {
		if err := hook(ctx); err != nil {
			return err
		}
	}
	return a.http.Listen(":8000")
}

func (a *application) Shutdown(ctx context.Context) error {
	for i := len(a.opts.Hooks.OnShutdown) - 1; i >= 0; i-- {
		if err := a.opts.Hooks.OnShutdown[i](ctx); err != nil {
			return err
		}
	}
	return a.http.ShutdownWithContext(ctx)
}
