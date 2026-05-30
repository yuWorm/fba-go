package main

import (
	"context"
	"log"

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
	plugin.MountRoutes(pluginContext.APIGroup(), pluginContext.Routes(), plugin.WithContainer(pluginContext.Container()))

	return app, nil
}
