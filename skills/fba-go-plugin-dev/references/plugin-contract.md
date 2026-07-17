# Plugin Contract

## Source Map

- Module contract: `core/plugin/module.go`
- Runtime context: `core/plugin/context.go`
- Registry and dependency resolution: `core/plugin/registry.go`
- Plugin modes: `core/plugin/mode.go`
- Generated registration: `cmd/fbago/internal/plugin/generate.go`
- Official registration example: `../fba-go-admin/internal/generated/fba_plugins.gen.go`
- Module examples: `../fba-go-admin/internal/app/admin/module.go`, `../fba-go-admin/plugins/task/module.go`

## Required Shape

Every plugin or app module exposes:

```go
func FBAPlugin() plugin.Module {
	return Module{}
}

type Module struct{}

func (Module) Meta() plugin.Meta {
	return plugin.Meta{ID: "task", Name: "Task Plugin", Version: "0.1.0"}
}

func (Module) Register(ctx plugin.Context) error {
	return nil
}
```

`FBAPlugin` is the scanner and generator entry point. Keep it simple and deterministic.

## Meta Fields

Use `plugin.Meta` to describe:

- `ID`: stable unique plugin ID.
- `Name`: human-readable name.
- `Version`: plugin version.
- `Description`: short purpose.
- `Author`: author or vendor.
- `Tags`: searchable capabilities.
- `DependsOn`: plugin dependencies.
- `Provides`: capabilities exported for discovery.
- `AutoInjectDefault`: whether templates should auto-inject by default.
- `PureDependencyDefault`: whether the plugin is normally only a dependency.

Dependencies use:

```go
plugin.Dependency{ID: "admin", Version: ">=0.1.0", Optional: true}
```

Current registry validates ID presence, duplicate IDs, dependency presence, disabled dependencies, and cycles. It does not currently enforce semver ranges, so version is metadata until enforcement is added.

## Registration Rules

During `Register`, declare capabilities through the context:

- `ctx.Provide`: register constructors into DI.
- `ctx.Route`: register route declarations.
- `ctx.Migration`: register migrations.
- `ctx.Command`: register CLI commands.
- `ctx.Task`: register task declarations.
- `ctx.Swagger`: register OpenAPI fragments.

Do not mount routes directly unless the plugin is explicitly lower-level infrastructure. Normal application routes should be declared and mounted by runtime.

## Repository Selection

Use the admin pattern:

1. Start with memory repository seed data when possible.
2. Resolve an injected repository override if the module supports tests or custom wiring.
3. Resolve `db.Provider`.
4. If a write DB exists, switch to GORM repository and register migrations.

This keeps generated projects runnable without a database while still enabling production persistence.

## Registration Order

The runtime calls `registry.RegisterAll`. The registry resolves dependencies first and only registers `ModeAuto` entries.

Do not assume source import order. If a plugin needs another plugin's `ctx.Provide` result, declare the dependency in `Meta().DependsOn`.

## Generated Registration and Version Lock

Projects declare official, third-party, and local plugin package paths in
`plugins.yaml`. Run:

```bash
fbago plugin sync
```

The command:

1. validates IDs, modes, and package paths
2. writes deterministic `internal/generated/fba_plugins.gen.go`
3. runs `go mod tidy`
4. resolves each package to its owning Go module
5. writes `plugins.lock` with module versions and local replacement commits

Use `fbago plugin sync --check` in CI. The runtime consumes
`generated.RegisterPlugins`; do not maintain a second handwritten registry.

`fbago plugin scan` remains the lower-level merger for manifest, local
`plugin.yaml`, and blank-import discovery.

Official modules are normal Go module dependencies. A local fork or Git
submodule is selected with:

```bash
fbago module use --path ../fba-go-admin github.com/yuWorm/fba-go-admin
```

Keep plugin package paths importable and side-effect free. Public module facades
may wrap internal implementations, but generated projects must never import
another module's implementation internals directly.
