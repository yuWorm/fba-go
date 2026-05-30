package order

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/plugin"
)

func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{
		ID:                "order",
		Name:              "Order Fixture",
		Version:           "0.1.0",
		AutoInjectDefault: true,
	}
}

func (Module) Register(ctx plugin.Context) error {
	return ctx.Route(plugin.Route{
		Method:       "GET",
		Path:         "/orders",
		Summary:      "Fixture orders endpoint",
		Tags:         []string{"fixture"},
		AuthRequired: false,
		Handler: func(c fiber.Ctx) error {
			return c.SendString("orders fixture")
		},
	})
}
