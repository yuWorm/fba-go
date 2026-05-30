package main

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3"
	fba "github.com/yuWorm/fba-go"
	"github.com/yuWorm/fba-go/core/plugin"
	"github.com/yuWorm/fba-go/examples/compat-host/internal/generated"
)

func main() {
	app, err := newApplication()
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func newApplication() (fba.Application, error) {
	app, err := fba.NewApplication(fba.Options{})
	if err != nil {
		return nil, err
	}

	registry := plugin.NewRegistry()
	if err := generated.RegisterPlugins(registry); err != nil {
		return nil, err
	}

	pluginContext := plugin.NewContext(plugin.ContextOptions{
		Container: app.Container(),
		Router:    app.HTTP(),
		APIGroup:  app.HTTP().Group("/api/v1"),
	})
	if err := registry.RegisterAll(pluginContext); err != nil {
		return nil, err
	}
	registerRoutes(pluginContext.APIGroup(), pluginContext.Routes())

	return app, nil
}

func registerRoutes(router fiber.Router, routes []plugin.Route) {
	for _, route := range routes {
		router.Add([]string{strings.ToUpper(route.Method)}, route.Path, route.Handler)
	}
}
