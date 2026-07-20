package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	coreauth "github.com/yuWorm/fba-go/core/auth"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/fiberx"
	"github.com/yuWorm/fba-go/core/observability"
	"github.com/yuWorm/fba-go/core/realtime"
	"github.com/yuWorm/fba-go/core/redisx"
	coretask "github.com/yuWorm/fba-go/core/task"
)

type Application interface {
	HTTP() *fiber.App
	Container() *di.Container
	Run(ctx context.Context) error
	RunHTTP(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type application struct {
	container    *di.Container
	http         *fiber.App
	opts         config.Options
	shutdownOnce sync.Once
	shutdownErr  error
}

func New(opts config.Options) (Application, error) {
	opts = opts.WithDefaults()
	if err := config.ValidateCORSOptions(opts.CORS); err != nil {
		return nil, fmt.Errorf("CORS configuration: %w", err)
	}
	var tokenService coreauth.TokenService
	if err := coreauth.ValidateJWTOptions(opts.Auth); err == nil {
		tokenService = coreauth.NewJWTService(opts.Auth)
	} else if !opts.Realtime.Disabled {
		return nil, fmt.Errorf("realtime authentication configuration: %w", err)
	}
	fx := fiberx.New(opts)
	observability.RegisterCoreRoutes(fx.App, observability.NewReadiness())
	container := di.New()
	var redisClient redisx.RedisClient
	if applicationRedisConfigured(opts) {
		redisClient = redisx.NewUniversalClient(opts.Redis)
		if err := container.Provide(func() redisx.RedisClient { return redisClient }); err != nil {
			_ = redisClient.Close()
			return nil, err
		}
		opts.Hooks.OnShutdown = append(opts.Hooks.OnShutdown, func(context.Context) error {
			return redisClient.Close()
		})
	}
	taskRegistry := coretask.NewRegistry()
	if err := container.Provide(func() coretask.DefinitionRegistry { return taskRegistry }); err != nil {
		return nil, err
	}
	taskRuntime := coretask.Runtime(coretask.NoopRuntime{})
	if opts.Task.Enabled {
		taskRedisClient := coretask.BackendRedisClient(redisx.NewUniversalClient(taskRedisOptions(opts)))
		asynqRuntime, err := coretask.NewAsynqRuntime(taskRedisClient, taskRegistry, coretask.AsynqRuntimeOptions{
			Concurrency: opts.Task.Concurrency,
			Queues:      opts.Task.Queues,
		})
		if err != nil {
			return nil, fmt.Errorf("task runtime: %w", err)
		}
		taskRuntime = asynqRuntime
		if err := container.Provide(func() coretask.BackendRedisClient { return taskRedisClient }); err != nil {
			return nil, err
		}
		opts.Hooks.OnStart = append(opts.Hooks.OnStart, asynqRuntime.Start)
		opts.Hooks.OnShutdown = append(opts.Hooks.OnShutdown, asynqRuntime.Shutdown)
	}
	if err := container.Provide(func() coretask.Runtime { return taskRuntime }); err != nil {
		return nil, err
	}
	if tokenService != nil {
		if err := container.Provide(func() coreauth.TokenService { return tokenService }); err != nil {
			return nil, err
		}
	}
	onlineStore := realtime.OnlineStore(realtime.NewMemoryOnlineStore())
	hubOptions := []realtime.SocketIOHubOption{}
	if !opts.Realtime.Disabled && opts.Realtime.MultiInstance.Enabled {
		keys := redisx.NewKeys(opts.Redis.KeyPrefix)
		onlineStore = realtime.NewRedisOnlineStore(redisClient, keys)
		nodeID := opts.Realtime.MultiInstance.NodeID
		if nodeID != "" {
			hubOptions = append(hubOptions, realtime.WithNodeID(nodeID))
		}
		hubOptions = append(hubOptions, realtime.WithBroadcaster(realtime.NewRedisBroadcaster(redisClient, opts.Realtime.MultiInstance.Channel)))
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
			Authenticator: realtime.NewJWTAuthenticator(tokenService, opts, realtime.WithAccessSessionValidatorResolver(container)),
		})
		server.Mount(fx.App)
		// Realtime JWTs must remain coupled to a revocable server-side session.
		// Resolve after plugins register so Admin or another session owner can
		// provide the validator without introducing a core-to-plugin dependency.
		opts.Hooks.OnStart = append(opts.Hooks.OnStart, func(context.Context) error {
			var validator realtime.AccessSessionValidator
			if !container.Resolve(&validator) || validator == nil {
				return fmt.Errorf("realtime access session validator is required")
			}
			return nil
		})
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

func (a *application) RunHTTP(ctx context.Context) (runErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	// Every exit path, including a failed start hook or listener, must unwind
	// resources that earlier hooks may already have started.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		runErr = errors.Join(runErr, a.Shutdown(shutdownCtx))
	}()
	for _, hook := range a.opts.Hooks.OnStart {
		if err := hook(ctx); err != nil {
			return err
		}
	}
	return a.http.Listen(":8000", fiber.ListenConfig{GracefulContext: ctx})
}

func (a *application) Shutdown(ctx context.Context) error {
	a.shutdownOnce.Do(func() {
		var shutdownErrs []error
		for i := len(a.opts.Hooks.OnShutdown) - 1; i >= 0; i-- {
			shutdownErrs = append(shutdownErrs, a.opts.Hooks.OnShutdown[i](ctx))
		}
		shutdownErrs = append(shutdownErrs, a.http.ShutdownWithContext(ctx))
		a.shutdownErr = errors.Join(shutdownErrs...)
	})
	return a.shutdownErr
}

func taskRedisOptions(opts config.Options) config.RedisOptions {
	redisOptions := opts.Redis
	if opts.Task.RedisMode != "" {
		redisOptions.Mode = opts.Task.RedisMode
	}
	if len(opts.Task.RedisAddrs) > 0 {
		redisOptions.Addrs = append([]string(nil), opts.Task.RedisAddrs...)
		redisOptions.Addr = ""
	} else if opts.Task.RedisAddr != "" {
		redisOptions.Addr = opts.Task.RedisAddr
		redisOptions.Addrs = nil
	}
	if opts.Task.RedisPassword != "" {
		redisOptions.Password = opts.Task.RedisPassword
	}
	if opts.Task.RedisMasterName != "" {
		redisOptions.MasterName = opts.Task.RedisMasterName
	}
	redisOptions.DB = opts.Task.RedisDB
	return redisOptions
}

func applicationRedisConfigured(opts config.Options) bool {
	redisOptions := opts.Redis
	return redisOptions.Mode != "" ||
		redisOptions.Addr != "" ||
		len(redisOptions.Addrs) > 0 ||
		redisOptions.Username != "" ||
		redisOptions.Password != "" ||
		redisOptions.MasterName != "" ||
		redisOptions.DB != 0 ||
		(!opts.Realtime.Disabled && opts.Realtime.MultiInstance.Enabled)
}
