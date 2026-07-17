# Admin Template

## Source Map

- Default embedded starter: `cmd/fbago/internal/scaffold/templates/admin`
- Embedded dependency metadata: `cmd/fbago/internal/scaffold/templates/admin/fbago-template.yaml`
- Versioned Admin module checkout: `../fba-go-admin`
- Runtime: `../fba-go-admin/internal/runtime`
- Generated module registration: `../fba-go-admin/internal/generated`
- Official module facades: `../fba-go-admin/modules`
- Admin plugin manifest and lock: `../fba-go-admin/plugins.yaml`, `../fba-go-admin/plugins.lock`
- Built-in implementations: `../fba-go-admin/internal/app/admin`, `config`, `dict`, `notice`
- Optional implementations: `../fba-go-admin/plugins/email`, `oauth2`, `task`, `uploadfile`
- Runnable module file: `../fba-go-admin/go.mod`

## Versioned Admin Module vs Embedded Starter

The versioned Admin module owns official implementation source. The default
`fbago init` starter is embedded separately under the core CLI and contains
only project bootstrap files.

Keep this distinction explicit:

- `../fba-go-admin/go.mod` belongs to `github.com/yuWorm/fba-go-admin`.
- `modules/*` are stable public `FBAPlugin` facades over implementation packages.
- The Admin module's `plugins.yaml` is its registration source of truth and
  `internal/generated/fba_plugins.gen.go` is generated from it.
- Embedded `admin/fbago-template.yaml` declares the Admin module and release
  version; it is never emitted.
- Embedded `go.mod.tmpl` renders the project module plus versioned dependencies
  on `fba-go` and `fba-go-admin`.
- Released scaffolds emit semantic module versions without replacements. Local
  integration explicitly uses `--template-replace` or
  `FBAGO_TEMPLATE_REPLACE`; the generated placeholder version is `v0.0.0`.

Generated projects own `internal/app`; official implementation source must not
be copied there. Upgrades happen through the Admin module version.

## Runtime Behavior

`../fba-go-admin/internal/runtime.NewWithOptions` composes:

1. configuration defaults
2. application creation
3. optional database provider
4. plugin registry
5. built-in module registration
6. plugin runtime context
7. plugin route mounting
8. CLI command execution

The default CLI command is `server`. Running `go run ./cmd/api` starts the server.

## Module Layout

Use the existing package shape:

```text
module/
  api/
  dto/
  migration/
  model/
  repo/
  service/
  module.go
  plugin_test.go
```

Not every module needs every package, but avoid mixing handler, repository, and migration logic in one file when the feature grows.

## Official Modules

Official module implementations:

- `admin`: auth, users, roles, menus, departments, data rules, logs, monitor, files, plugin management.
- `config`: system config APIs and admin config provider.
- `dict`: dictionary types and values.
- `notice`: notices and initial notice data.
- `email`: email integration.
- `oauth2`: OAuth2 login.
- `task`: scheduler and task management.
- `uploadfile`: storage, upload, sharing, and cleanup.

Every implementation has a stable facade under `modules/<id>`. Generated
registration imports only these facades, never `internal/*`.

## Managed Manifest

The embedded `admin/fbago.yaml.tmpl` manages only thin bootstrap files such as
`cmd/api/main.go` and the generated-project Makefile.

Rules:

- Never add official app/plugin implementation directories back to `managed`.
- Project business under `internal/app` is project-owned and never overwritten.
- `plugins.yaml` is project-owned; plugin registration is refreshed with
  `fbago plugin sync`, not template update.
- Official feature updates use Go module versions.
- Local forks or Git submodules are selected with `fbago module use --path <checkout> github.com/yuWorm/fba-go-admin`.

## Data Modes

Modules should support memory mode when practical. When database config is present, modules should switch to GORM repositories and register migrations.

This lets the template run tests without external services and still behave like production when configured.

## Verification

Verify the independent Admin module, its generated registration, and the
embedded starter:

```bash
make -C ../fba-go-admin test
GOWORK=off go run ./cmd/fbago plugin sync --dir ../fba-go-admin --check
make verify-template
```

After scaffold changes, generate the default project without `--template`,
confirm it contains no official implementation directories, then run
`fbago plugin sync --check`, `go test ./...`, `go run ./cmd/api --help`, and
`fbago template diff`.
