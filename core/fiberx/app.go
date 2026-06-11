package fiberx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
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
	opts.Fiber.ErrorHandler = middleware.NewErrorHandler(opts)

	app := fiber.New(opts.Fiber)
	if opts.Middleware.RequestID.Enabled {
		app.Use(middleware.RequestID())
	}
	if opts.Middleware.AccessLog.Enabled || opts.Middleware.ErrorLog.Enabled {
		app.Use(middleware.HTTPLogger(opts))
	}
	if opts.Middleware.Recover.Enabled {
		app.Use(middleware.Recover(middleware.RecoverConfig{
			EnableStackTrace: opts.Middleware.Recover.EnableStackTrace,
		}))
	}
	if opts.CORS.Enabled {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     opts.CORS.AllowedOrigins,
			AllowCredentials: opts.CORS.AllowCredentials,
			AllowMethods:     opts.CORS.AllowMethods,
			AllowHeaders:     opts.CORS.AllowHeaders,
			ExposeHeaders:    opts.CORS.ExposeHeaders,
		}))
	}
	return &CoreApp{
		App:     app,
		API:     app.Group(opts.App.APIBasePath),
		Options: opts,
	}
}
