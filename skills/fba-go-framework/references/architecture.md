# FBA Go Architecture

## Source Map

- Public facade: `fba.go`
- Application runtime: `core/app/application.go`
- HTTP and middleware setup: `core/fiberx/app.go`, `core/middleware/*`
- Dependency injection: `core/di/container.go`
- Plugin contracts: `core/plugin/module.go`, `core/plugin/context.go`, `core/plugin/registry.go`
- Route mounting and RBAC bridge: `core/plugin/route.go`, `core/plugin/mount.go`, `core/rbac/*`
- Database provider: `core/db/*`
- Migrations: `core/migration/*`
- Realtime: `core/realtime/*`
- Official Admin runtime composition: `../fba-go-admin/internal/runtime/runtime.go`
- Official Admin module facade: `../fba-go-admin/admin.go`
- Official module facades: `../fba-go-admin/modules/*`
- Admin generated registration: `../fba-go-admin/internal/generated`

## Runtime Shape

`fba.go` intentionally exposes a small facade:

- `LoadOptionsFromEnv`
- `LoadOptionsFromEnvFile`
- `NewApplication`
- aliases for `Application`, `Options`, and hooks

Core application creation lives in `core/app.New`. It builds:

- a Fiber app via `fiberx.New`
- core observability routes
- a DI container
- realtime hub and online store
- optional Redis-backed realtime broadcaster
- startup and shutdown hooks

The core application does not know which business modules exist. It owns infrastructure and exposes `Application.HTTP()` and `Application.Container()`.

## Versioned Admin Runtime Composition

The independent `github.com/yuWorm/fba-go-admin` repository is a standalone Go
module. Generated projects depend on it by version and call its public
`admin.Run` facade; official implementation source is not copied into projects.

The Admin runtime:

1. Loads options from `.env` and environment variables.
2. Creates `fba.Application`.
3. Opens the database if configured and provides `db.Provider` into DI.
4. Runs the project `Configure` hook so DI overrides exist before modules.
5. Builds a `plugin.Registry`.
6. Calls generated `RegisterPlugins` from `plugins.yaml`.
7. Creates `plugin.RuntimeContext`.
8. Calls `registry.RegisterAll`.
9. Mounts collected routes onto `cfg.App.APIBasePath`.
10. Executes CLI commands with default command `server`.

Plugins declare capabilities, runtime decides when to mount routes, run
migrations, and execute commands. Project business stays under the generated
project's `internal/app` and joins the same generated registry.

## Plugin Context Contract

Plugins interact through `plugin.Context`:

- `Container()` resolves or provides services.
- `Router()` exposes the root Fiber router.
- `APIGroup()` exposes the versioned API group.
- `Config()` exposes resolved config.
- `Provide()` registers constructors into DI.
- `Route()` collects a route declaration.
- `Task()` collects a task declaration.
- `Migration()` collects a migration.
- `Command()` collects a CLI command.
- `Swagger()` collects an OpenAPI fragment.

`RuntimeContext` stores these declarations and returns defensive copies through `Routes`, `Tasks`, `Migrations`, `Commands`, and `SwaggerFragments`.

## Plugin Registry

`plugin.Registry` stores `plugin.Module` entries by ID and mode.

Modes:

- `ModeAuto`: register automatically.
- `ModeDisabled`: skip registration and fail required dependents.
- `ModePureDependency`: available as dependency metadata but not auto-registered by `RegisterAll`.

`Resolve` topologically sorts dependencies and reports missing dependencies, disabled required dependencies, and cycles.

## Routing and RBAC

Routes are declared with helpers:

- `plugin.GET`
- `plugin.POST`
- `plugin.PUT`
- `plugin.DELETE`

Auth metadata is attached with:

- `plugin.Auth()`
- `plugin.Perm("permission:code")`
- `plugin.Superuser()`

`MountRoutes` wraps protected routes with an authenticator resolved from DI or provided explicitly. Auth-only routes require a current user but do not apply permission checks unless the route declares RBAC metadata.

## Migrations

Core migration contracts live under `core/migration`. The admin runtime uses a `GORMStore` and `Runner` over migrations collected from plugins. A migration is registered by calling `ctx.Migration(...)` during plugin registration.

Database-aware plugins should register migrations only when a usable `db.Provider` is available. Memory-only fallback should not register GORM migrations.

## Extension Boundary

Use core packages for stable, cross-project contracts. Use template runtime for composition. Use plugins and app modules for business features.

Do not add business-specific behavior to core unless multiple generated projects need the same stable contract.
