package plugin_test

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/command"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/migration"
	"github.com/yuWorm/fba-go/core/plugin"
	"go.uber.org/zap"
)

func TestModuleCanRegisterCoreContributions(t *testing.T) {
	app := fiber.New()
	ctx := plugin.NewContext(plugin.ContextOptions{
		Container: di.New(),
		Router:    app,
		APIGroup:  app.Group("/api/v1"),
		Logger:    zap.NewNop(),
	})

	module := testModule{}
	if err := module.Register(ctx); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if len(ctx.Routes()) != 1 {
		t.Fatalf("routes = %d, want 1", len(ctx.Routes()))
	}
	if len(ctx.Migrations()) != 1 {
		t.Fatalf("migrations = %d, want 1", len(ctx.Migrations()))
	}
	if len(ctx.Tasks()) != 1 {
		t.Fatalf("tasks = %d, want 1", len(ctx.Tasks()))
	}
	if len(ctx.Commands()) != 1 {
		t.Fatalf("commands = %d, want 1", len(ctx.Commands()))
	}
	if len(ctx.SwaggerFragments()) != 1 {
		t.Fatalf("swagger fragments = %d, want 1", len(ctx.SwaggerFragments()))
	}

	var service string
	if err := ctx.Container().Invoke(func(value string) {
		service = value
	}); err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if service != "provided" {
		t.Fatalf("service = %q, want provided", service)
	}
}

type testModule struct{}

func (testModule) Meta() plugin.Meta {
	return plugin.Meta{
		ID:      "order",
		Name:    "订单插件",
		Version: "0.1.0",
		DependsOn: []plugin.Dependency{
			{ID: "admin", Version: ">=0.1.0"},
		},
		AutoInjectDefault: true,
	}
}

func (testModule) Register(ctx plugin.Context) error {
	if err := ctx.Provide(func() string { return "provided" }); err != nil {
		return err
	}
	if err := ctx.Route(plugin.Route{
		Method:       "GET",
		Path:         "/orders",
		Summary:      "分页获取订单",
		Tags:         []string{"订单"},
		Permission:   "order:list",
		AuthRequired: true,
		Handler:      func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) },
	}); err != nil {
		return err
	}
	if err := ctx.Migration(migration.Migration{Scope: "plugin:order", Version: "0001"}); err != nil {
		return err
	}
	if err := ctx.Task(plugin.TaskDefinition{Type: "order:close", Name: "订单关闭"}); err != nil {
		return err
	}
	if err := ctx.Command(command.Command{Use: "order close", Short: "Close expired orders"}); err != nil {
		return err
	}
	return ctx.Swagger(plugin.SwaggerFragment{PluginID: "order"})
}
