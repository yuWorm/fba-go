package app

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/fiberx"
	"github.com/yuWorm/fba-go/core/observability"
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
	return &application{
		container: di.New(),
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
