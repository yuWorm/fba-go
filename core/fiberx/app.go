package fiberx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/middleware"
)

type CoreApp struct {
	App     *fiber.App
	API     fiber.Router
	Options config.Options
}

func New(opts config.Options) *CoreApp {
	opts = opts.WithDefaults()
	opts.Fiber.ErrorHandler = middleware.ErrorHandler

	app := fiber.New(opts.Fiber)
	return &CoreApp{
		App:     app,
		API:     app.Group(opts.App.APIBasePath),
		Options: opts,
	}
}
