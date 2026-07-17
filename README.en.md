# fba-go

[中文](README.md)

`fba-go` is the Go migration and evolution of [FastAPI Best Architecture](https://github.com/fastapi-practices/fastapi_best_architecture). It provides a reusable backend core as a Go module, using Fiber v3, plugin registration, and project templates to host admin, auth, config, dict, notice, task scheduler, and other business implementations.

## Features

- **Go module core**: application bootstrap, config loading, response envelopes, pagination, auth middleware, RBAC, plugin registration, Swagger, realtime events, and task contracts.
- **Fiber v3**: HTTP is built on Fiber v3.
- **Plugin system**: official, third-party, and project-local modules are registered through `plugins.yaml`; `plugins.lock` records resolved Go module versions and local replacements.
- **Embedded Admin template**: `fbago` carries the thin Admin starter directly; the complete feature set is supplied by the independently versioned [`fba-go-admin`](https://github.com/yuWorm/fba-go-admin) module.
- **Editable source**: Go modules are the default upgrade path; use `--template-replace` during initialization or `fbago module use` in an existing project to select a local fork or Git submodule.

## Quick Start

Install the CLI:

```bash
go install github.com/yuWorm/fba-go/cmd/fbago@latest
```

Create the default Admin project:

```bash
fbago init github.com/your-org/my-admin --dir ./my-admin
cd my-admin

fbago plugin sync --check
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

Create the default Admin starter:

```bash
fbago init github.com/your-org/my-admin --dir ./my-admin

cd my-admin
fbago plugin sync --check
make test
make run
```

Create a minimal backend:

```bash
fbago init github.com/your-org/my-backend \
  --template basic \
  --dir ./my-backend

cd my-backend
make tidy
make test
make run
```

## Repository Layout

| Path | Purpose |
| --- | --- |
| `core/` | Reusable core APIs and contracts |
| `cmd/fbago/` | CLI plus the embedded `basic` and `admin` scaffolds |
| `contracts/` | Core smoke API contract definitions |
| `templates/fba-go-template/` | AI engineering skills and scaffold integration assets |
| `docs/` | Migration and design docs |
| `sources/fastapi-best-architecture/` | Optional local reference source, usually not published with the repository |

## Common Commands

| Command | Description |
| --- | --- |
| `go install github.com/yuWorm/fba-go/cmd/fbago@latest` | Install the CLI |
| `fbago init <module>` | Create a project with the embedded Admin template |
| `fbago init <module> --template-replace ../fba-go-admin` | Initialize with a local Admin checkout |
| `fbago init <module> --template basic` | Create a minimal backend project |
| `go run ./cmd/fbago template list` | List embedded templates during local development |
| `fbago template diff` | Show `.fbago.yaml` managed template source changes |
| `fbago template update --dry-run` | Preview managed template source updates |
| `fbago plugin sync` | Generate registration, tidy dependencies, and write the version-aware plugin lock |
| `fbago plugin sync --check` | Verify registration, `go.mod`, `go.sum`, and the plugin lock are current |
| `fbago module use --path ../fba-go-admin github.com/yuWorm/fba-go-admin` | Use a local Admin module checkout |
| `fbago module reset github.com/yuWorm/fba-go-admin` | Remove the local override and return to the selected version |
| `make test` | Run core tests |
| `make verify-template` | Verify the embedded Admin starter against the independent Admin module |

## Local Development

```bash
git clone --recursive https://github.com/yuWorm/fba-go.git
git clone https://github.com/yuWorm/fba-go-admin.git
cd fba-go

# If cloned without --recursive
git submodule update --init --recursive

make test
make verify-template
```

Released projects use fixed semantic module versions. To integrate unpublished
Admin source, pass `--template-replace <fba-go-admin checkout>` to `fbago init`;
after generation, use `fbago module use` to select a local checkout explicitly.

## More Docs

- [Migration and design doc](docs/fba_go_module_migration_ha_design.md)
- [Admin module repository](https://github.com/yuWorm/fba-go-admin)
