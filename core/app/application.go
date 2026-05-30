package app

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
)

type Application interface {
	HTTP() *fiber.App
	Run(ctx context.Context) error
	RunHTTP(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type application struct {
	http *fiber.App
	opts config.Options
}

func New(opts config.Options) (Application, error) {
	opts = opts.WithDefaults()
	return &application{
		http: fiber.New(opts.Fiber),
		opts: opts,
	}, nil
}

func (a *application) HTTP() *fiber.App {
	return a.http
}

func (a *application) Run(ctx context.Context) error {
	return a.RunHTTP(ctx)
}

func (a *application) RunHTTP(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return a.http.Listen(":8000")
}

func (a *application) Shutdown(ctx context.Context) error {
	return a.http.ShutdownWithContext(ctx)
}
