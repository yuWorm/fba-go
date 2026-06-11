# fba-go

[中文](README.md)

`fba-go` is the Go migration and evolution of [FastAPI Best Architecture](https://github.com/fastapi-practices/fastapi_best_architecture). It provides a reusable backend core as a Go module, using Fiber v3, plugin registration, and project templates to host admin, auth, config, dict, notice, task scheduler, and other business implementations.

## Features

- **Go module core**: application bootstrap, config loading, response envelopes, pagination, auth middleware, RBAC, plugin registration, Swagger, realtime events, and task contracts.
- **Fiber v3**: HTTP is built on Fiber v3.
- **Plugin system**: core keeps plugin registration and scanning; editable business code should live under `internal/app` and `plugins` in generated projects.
- **Compatibility first**: the main repository keeps core contracts; full admin API behavior is owned and verified by the template repository.
- **Templates**: the official template repository is included as a submodule under `templates/fba-go-template`, and remote Git templates are supported.

## Quick Start

Install the CLI:

```bash
go install github.com/yuWorm/fba-go/cmd/fbago@latest
```

Create a project with the CLI:

```bash
fbago init github.com/your-org/my-backend --dir ./my-backend
cd my-backend

go get github.com/yuWorm/fba-go@latest
make tidy
make test
make run
```

You can also use `fba-go` directly in an existing project:

```go
package main

import (
	"context"
	"log"

	fba "github.com/yuWorm/fba-go"
)

func main() {
	opts, err := fba.LoadOptionsFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	app, err := fba.NewApplication(opts)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

Run it:

```bash
go run .
```

## Scaffolding

Create a minimal backend project:

```bash
fbago init github.com/your-org/my-backend --dir ./my-backend

cd my-backend
go get github.com/yuWorm/fba-go@latest
make tidy
make test
make run
```

Create an admin starter backend:

```bash
fbago init github.com/your-org/my-admin \
  --template github.com/yuWorm/fba-go-template/admin@v0.0.1 \
  --dir ./my-admin

cd my-admin
go get github.com/yuWorm/fba-go@latest
make tidy
make test
make run
```

## Repository Layout

| Path | Purpose |
| --- | --- |
| `core/` | Reusable core APIs and contracts |
| `cmd/fbago/` | CLI for scaffolding, plugin scan, Swagger, and contract tests |
| `contracts/` | Core smoke API contract definitions |
| `templates/fba-go-template/` | Official template repository submodule |
| `docs/` | Migration and design docs |
| `sources/fastapi-best-architecture/` | Optional local reference source, usually not published with the repository |

## Common Commands

| Command | Description |
| --- | --- |
| `go install github.com/yuWorm/fba-go/cmd/fbago@latest` | Install the CLI |
| `fbago init <module>` | Create a project |
| `go run ./cmd/fbago template list` | List built-in templates for local development |
| `fbago template diff` | Show `.fbago.yaml` managed template source changes |
| `fbago template update --dry-run` | Preview managed template source updates |
| `make test` | Run core tests |
| `make verify-template` | Verify the official admin template and generated project |

## Local Development

```bash
git clone --recursive https://github.com/yuWorm/fba-go.git
cd fba-go

# If cloned without --recursive
git submodule update --init --recursive

make test
```

Local development builds of `fbago` automatically write `replace github.com/yuWorm/fba-go => <local-path>` into generated projects. Installed builds or other layouts can use `--core-replace /path/to/fba-go` or `FBAGO_CORE_REPLACE`.

## More Docs

- [Migration and design doc](docs/fba_go_module_migration_ha_design.md)
- [Template repository](https://github.com/yuWorm/fba-go-template)
