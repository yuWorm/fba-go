package plugin

import (
	"github.com/gofiber/fiber/v3"
	"github.com/yuWorm/fba-go/core/command"
	"github.com/yuWorm/fba-go/core/config"
	"github.com/yuWorm/fba-go/core/di"
	"github.com/yuWorm/fba-go/core/migration"
	"go.uber.org/zap"
)

type Context interface {
	Container() *di.Container
	Router() fiber.Router
	APIGroup() fiber.Router
	Logger() *zap.Logger
	Config() config.Options

	Provide(constructor any) error
	Route(route Route) error
	Task(task TaskDefinition) error
	Migration(m migration.Migration) error
	Command(command.Command) error
	Swagger(fragment SwaggerFragment) error
}

type ContextOptions struct {
	Container *di.Container
	Router    fiber.Router
	APIGroup  fiber.Router
	Logger    *zap.Logger
	Config    config.Options
}

type RuntimeContext struct {
	container *di.Container
	router    fiber.Router
	apiGroup  fiber.Router
	logger    *zap.Logger
	config    config.Options

	routes           []Route
	tasks            []TaskDefinition
	migrations       []migration.Migration
	commands         []command.Command
	swaggerFragments []SwaggerFragment
}

func NewContext(opts ContextOptions) *RuntimeContext {
	if opts.Container == nil {
		opts.Container = di.New()
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	return &RuntimeContext{
		container: opts.Container,
		router:    opts.Router,
		apiGroup:  opts.APIGroup,
		logger:    opts.Logger,
		config:    opts.Config,
	}
}

func (c *RuntimeContext) Container() *di.Container {
	return c.container
}

func (c *RuntimeContext) Router() fiber.Router {
	return c.router
}

func (c *RuntimeContext) APIGroup() fiber.Router {
	return c.apiGroup
}

func (c *RuntimeContext) Logger() *zap.Logger {
	return c.logger
}

func (c *RuntimeContext) Config() config.Options {
	return c.config
}

func (c *RuntimeContext) Provide(constructor any) error {
	return c.container.Provide(constructor)
}

func (c *RuntimeContext) Route(route Route) error {
	c.routes = append(c.routes, route)
	return nil
}

func (c *RuntimeContext) Task(task TaskDefinition) error {
	c.tasks = append(c.tasks, task)
	return nil
}

func (c *RuntimeContext) Migration(m migration.Migration) error {
	c.migrations = append(c.migrations, m)
	return nil
}

func (c *RuntimeContext) Command(command command.Command) error {
	c.commands = append(c.commands, command)
	return nil
}

func (c *RuntimeContext) Swagger(fragment SwaggerFragment) error {
	c.swaggerFragments = append(c.swaggerFragments, fragment)
	return nil
}

func (c *RuntimeContext) Routes() []Route {
	return append([]Route(nil), c.routes...)
}

func (c *RuntimeContext) Tasks() []TaskDefinition {
	return append([]TaskDefinition(nil), c.tasks...)
}

func (c *RuntimeContext) Migrations() []migration.Migration {
	return append([]migration.Migration(nil), c.migrations...)
}

func (c *RuntimeContext) Commands() []command.Command {
	return append([]command.Command(nil), c.commands...)
}

func (c *RuntimeContext) SwaggerFragments() []SwaggerFragment {
	return append([]SwaggerFragment(nil), c.swaggerFragments...)
}
