# FBA Go Core Module And Business Plugin Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `github.com/yuWorm/fba-go` as the reusable core Go module, with every business capability delivered as a build-time registered Go module plugin.

**Architecture:** The root repository is core-only: app lifecycle, config, Fiber integration, DI, response/errors, logger, Redis/DB providers, auth primitives, RBAC primitives, task primitives, migration, Swagger, plugin registry, contracts, and `fbago`. Frontend-compatible business routes such as `/api/v1/auth/*`, `/api/v1/sys/*`, `/api/v1/dict-*`, `/api/v1/tasks/*`, `/api/v1/logs/*`, and `/api/v1/monitors/*` are implemented by separate business plugin modules and compiled into a host app through generated registration code.

**Tech Stack:** Go 1.23.5, Fiber v3, Dig, Zap, go-redis, GORM-compatible DB abstraction, Asynq, Pond, Prometheus, OpenAPI/Swagger fragments, YAML manifests, build-time code generation.

---

## Boundary Decision

This plan updates the source design in `docs/fba_go_module_migration_ha_design.md` with one important rule:

```text
fba-go repository = reusable core Go module only
business modules = Go module plugins
```

Core packages may expose framework services, interfaces, helpers, middleware, and code generation. Core packages must not contain concrete admin, dict, task, monitor, or log business handlers.

Allowed core-owned HTTP endpoints:

```text
GET /healthz
GET /readyz
GET /metrics
GET /docs
GET /openapi
GET /swagger/doc.json
```

Business plugin-owned compatible endpoints:

```text
/api/v1/auth/*
/api/v1/sys/*
/api/v1/dict-types*
/api/v1/dict-datas*
/api/v1/logs/*
/api/v1/monitors/*
/api/v1/tasks/*
/api/v1/task-results*
/api/v1/schedulers*
```

The first official plugin set should be delivered as separate Go modules:

```text
github.com/yuWorm/fba-plugin-admin
github.com/yuWorm/fba-plugin-dict
github.com/yuWorm/fba-plugin-task
```

This repository may contain test fixtures and examples, but not production business plugin code.

---

## File Structure

Create or modify these core files in this repository:

```text
fba.go                                      # public NewApplication entrypoint
go.mod                                     # core module dependencies
contracts/api.contract.yaml                # frozen frontend API contract
contracts/response.contract.yaml           # response envelope contract
contracts/redis.contract.yaml              # compatible Redis key contract
core/app/application.go                     # application lifecycle
core/app/hooks.go                           # lifecycle hooks
core/config/options.go                      # root Options and defaults
core/di/container.go                        # Dig wrapper
core/fiberx/app.go                          # Fiber factory and route helpers
core/fiberx/route.go                        # framework route metadata
core/response/response.go                   # success/error envelope
core/pagination/pagination.go               # PageData and links
core/datetime/datetime.go                   # compatible date JSON format
core/errors/errors.go                       # typed framework errors
core/middleware/*.go                        # recover, request id, auth adapter, operation log hook
core/logger/logger.go                       # Zap builder
core/redisx/client.go                       # Redis universal client factory
core/redisx/keys.go                         # compatible key helpers
core/db/provider.go                         # DB provider abstraction
core/db/gorm.go                             # initial GORM provider
core/migration/*.go                         # migration runner and lock interface
core/auth/*.go                              # token/session/password interfaces and implementations
core/rbac/*.go                              # permission and data-scope primitives
core/task/*.go                              # Asynq client/server abstractions
core/pool/*.go                              # Pond provider
core/observability/*.go                     # health/readiness/metrics
core/swagger/*.go                           # OpenAPI aggregation and UI handler
core/plugin/*.go                            # plugin SDK, registry, dependency graph
cmd/fbago/main.go                          # CLI root
cmd/fbago/internal/scaffold/*.go           # backend project scaffold init
cmd/fbago/internal/plugin/*.go             # scan, manifest, graph, codegen
cmd/fbago/internal/swagger/*.go            # swagger aggregation
cmd/fbago/internal/contract/*.go           # snapshot and contract test
examples/compat-host/*                      # minimal host app consuming generated plugins
internal/testplugin/*                       # test-only plugin fixtures
```

External plugin modules are planned separately and should use this structure:

```text
fba-plugin-admin/
  go.mod
  plugin.yaml
  plugin.go
  api/
  dto/
  service/
  repo/
  model/
  migration/
  docs/swagger.yaml

fba-plugin-dict/
  same layout

fba-plugin-task/
  same layout
```

---

## Chunk 1: Core Skeleton And Contracts

### Task 1: Freeze core-only module shape

**Files:**
- Modify: `go.mod`
- Create: `fba.go`
- Create: `core/config/options.go`
- Create: `core/app/application.go`
- Test: `core/app/application_test.go`

- [ ] **Step 1: Write the failing public API test**

```go
package app_test

import (
    "context"
    "testing"

    fba "github.com/yuWorm/fba-go"
)

func TestNewApplicationBuildsCoreApp(t *testing.T) {
    app, err := fba.NewApplication(fba.Options{})
    if err != nil {
        t.Fatalf("NewApplication() error = %v", err)
    }
    if app.HTTP() == nil {
        t.Fatal("HTTP() returned nil")
    }
    if err := app.Shutdown(context.Background()); err != nil {
        t.Fatalf("Shutdown() error = %v", err)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core/app`

Expected: FAIL because `fba.NewApplication`, `Options`, or `Application` does not exist.

- [ ] **Step 3: Add minimal public API**

Create `fba.go`:

```go
package fba

import (
    "github.com/yuWorm/fba-go/core/app"
    "github.com/yuWorm/fba-go/core/config"
)

type Options = config.Options
type Application = app.Application

func NewApplication(opts Options) (Application, error) {
    return app.New(opts)
}
```

Create `core/config/options.go`:

```go
package config

type Options struct {
    App AppOptions
}

type AppOptions struct {
    Name        string
    Version     string
    Environment string
    APIBasePath string
    Timezone    string
}
```

Create `core/app/application.go` with a minimal implementation that constructs a Fiber app and supports `HTTP`, `Run`, and `Shutdown`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./core/app`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add go.mod fba.go core/config/options.go core/app/application.go core/app/application_test.go
git commit -m "feat: scaffold core fba module"
```

### Task 2: Add compatibility contracts

**Files:**
- Create: `contracts/api.contract.yaml`
- Create: `contracts/response.contract.yaml`
- Create: `contracts/redis.contract.yaml`

- [ ] **Step 1: Create contract fixtures from the migration design**

Include first-priority routes:

```yaml
routes:
  - method: GET
    path: /api/v1/auth/captcha
  - method: POST
    path: /api/v1/auth/login
  - method: POST
    path: /api/v1/auth/refresh
  - method: POST
    path: /api/v1/auth/logout
  - method: GET
    path: /api/v1/auth/codes
  - method: GET
    path: /api/v1/sys/users/me
  - method: GET
    path: /api/v1/sys/menus/sidebar
  - method: GET
    path: /api/v1/dict-datas/type-codes/{code}
```

- [ ] **Step 2: Add response and Redis key contracts**

Response contract must require `{code,msg,data}` and error `trace_id`.

Redis contract must include token, refresh token, user cache, captcha, limiter, plugin, migration, and scheduler leader keys.

- [ ] **Step 3: Validate YAML parses**

Run: `go test ./cmd/fbago/internal/contract`

Expected initially: FAIL until the contract package exists. Mark this as pending until Chunk 8.

- [ ] **Step 4: Commit**

Run:

```bash
git add contracts
git commit -m "docs: freeze frontend compatibility contracts"
```

---

## Chunk 2: Response, Error, Pagination, DateTime

### Task 3: Implement response envelope and pagination

**Files:**
- Create: `core/response/response.go`
- Create: `core/errors/errors.go`
- Create: `core/pagination/pagination.go`
- Test: `core/errors/errors_test.go`
- Test: `core/response/response_test.go`
- Test: `core/pagination/pagination_test.go`

- [ ] **Step 1: Write failing response tests**

Test success JSON equals:

```json
{"code":200,"msg":"成功","data":null}
```

Test error JSON includes:

```json
{"code":400,"msg":"请求参数非法","data":null,"trace_id":"trace-1"}
```

- [ ] **Step 2: Run tests**

Run: `go test ./core/errors ./core/response ./core/pagination`

Expected: FAIL because packages do not exist.

- [ ] **Step 3: Implement minimal structs**

```go
type Response[T any] struct {
    Code int    `json:"code"`
    Msg  string `json:"msg"`
    Data T      `json:"data"`
}

type ErrorResponse struct {
    Code    int    `json:"code"`
    Msg     string `json:"msg"`
    Data    any    `json:"data"`
    TraceID string `json:"trace_id,omitempty"`
}
```

Implement `pagination.PageData[T]` with `items`, `total`, `page`, `size`, `total_pages`, and `links`.

Implement `errors.AppError` with HTTP status, compatible response code, public message, internal cause, and optional trace ID extraction for Fiber error handling.

- [ ] **Step 4: Run tests**

Run: `go test ./core/errors ./core/response ./core/pagination`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/errors core/response core/pagination
git commit -m "feat: add compatible response models"
```

### Task 4: Implement compatible DateTime JSON

**Files:**
- Create: `core/datetime/datetime.go`
- Test: `core/datetime/datetime_test.go`

- [ ] **Step 1: Write failing DateTime test**

Assert JSON marshaling uses `2006-01-02 15:04:05` in configured timezone.

- [ ] **Step 2: Run test**

Run: `go test ./core/datetime`

Expected: FAIL.

- [ ] **Step 3: Implement DateTime**

Expose:

```go
type DateTime time.Time

func SetLocation(loc *time.Location)
func (d DateTime) MarshalJSON() ([]byte, error)
```

- [ ] **Step 4: Run test**

Run: `go test ./core/datetime`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/datetime
git commit -m "feat: add compatible datetime serialization"
```

---

## Chunk 3: App, DI, Fiber, Logger

### Task 5: Wire Dig and lifecycle hooks

**Files:**
- Create: `core/di/container.go`
- Create: `core/app/hooks.go`
- Modify: `core/app/application.go`
- Test: `core/di/container_test.go`
- Test: `core/app/application_test.go`

- [ ] **Step 1: Write failing DI test**

Test `Provide` and `Invoke` wrap Dig errors with readable context.

- [ ] **Step 2: Run tests**

Run: `go test ./core/di ./core/app`

Expected: FAIL.

- [ ] **Step 3: Implement container wrapper and lifecycle hooks**

Expose `di.Provider`, `di.Container`, `app.Hooks`, `OnStart`, `OnShutdown`, and deterministic shutdown ordering.

- [ ] **Step 4: Run tests**

Run: `go test ./core/di ./core/app`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/di core/app
git commit -m "feat: add di and app lifecycle"
```

### Task 6: Build Fiber factory and middleware chain

**Files:**
- Create: `core/fiberx/app.go`
- Create: `core/fiberx/route.go`
- Create: `core/middleware/recover.go`
- Create: `core/middleware/request_id.go`
- Modify: `core/app/application.go`
- Test: `core/fiberx/app_test.go`
- Test: `core/middleware/middleware_test.go`

- [ ] **Step 1: Write failing route tests**

Verify `APIBasePath` defaults to `/api/v1`, route metadata keeps permission/auth flags, and request ID is copied into context.

- [ ] **Step 2: Run tests**

Run: `go test ./core/fiberx ./core/middleware`

Expected: FAIL.

- [ ] **Step 3: Implement Fiber factory**

Use `fiber.New(fiber.Config{ErrorHandler: ...})`, expose `fiberx.New`, and ensure no business routes are registered by default.

- [ ] **Step 4: Run tests**

Run: `go test ./core/fiberx ./core/middleware`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/fiberx core/middleware core/app
git commit -m "feat: add fiber core integration"
```

### Task 7: Add Zap logger provider

**Files:**
- Create: `core/logger/logger.go`
- Modify: `core/config/options.go`
- Test: `core/logger/logger_test.go`

- [ ] **Step 1: Write failing logger tests**

Assert default level is `info`, default encoding is `json`, and invalid levels return errors.

- [ ] **Step 2: Run test**

Run: `go test ./core/logger`

Expected: FAIL.

- [ ] **Step 3: Implement logger options and constructor**

Expose `logger.New(config.LoggerOptions) (*zap.Logger, error)`.

- [ ] **Step 4: Run test**

Run: `go test ./core/logger`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/logger core/config
git commit -m "feat: add zap logger provider"
```

---

## Chunk 4: Redis, DB, Migration

### Task 8: Implement Redis universal client factory and key helpers

**Files:**
- Create: `core/redisx/client.go`
- Create: `core/redisx/keys.go`
- Modify: `core/config/options.go`
- Test: `core/redisx/keys_test.go`

- [ ] **Step 1: Write failing key compatibility tests**

Assert helpers produce:

```text
fba:token:10001:session-1
fba:refresh_token:10001:session-1
fba:user:10001
fba:login:captcha:uuid-1
fba:task:scheduler:leader
fba:migration:lock
```

- [ ] **Step 2: Run test**

Run: `go test ./core/redisx`

Expected: FAIL.

- [ ] **Step 3: Implement Redis options, factory, and key helpers**

Support `single`, `sentinel`, and `cluster` option mapping. Do not call Redis in unit tests.

- [ ] **Step 4: Run test**

Run: `go test ./core/redisx`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/redisx core/config
git commit -m "feat: add redis provider and compatible keys"
```

### Task 9: Add DB provider abstraction and migration lock

**Files:**
- Create: `core/db/provider.go`
- Create: `core/db/gorm.go`
- Create: `core/migration/migration.go`
- Create: `core/migration/runner.go`
- Create: `core/migration/lock.go`
- Test: `core/migration/runner_test.go`

- [ ] **Step 1: Write failing migration runner tests**

Use in-memory fake lock and fake store. Assert runner records `scope`, `version`, `checksum`, success, error, and execution time.

- [ ] **Step 2: Run test**

Run: `go test ./core/db ./core/migration`

Expected: FAIL.

- [ ] **Step 3: Implement provider interfaces**

Expose:

```go
type Provider interface {
    Write() any
    Read() any
    Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}
```

Use GORM in the first concrete provider, but keep plugin code consuming the abstraction unless it needs GORM-specific features.

- [ ] **Step 4: Implement migration runner**

Use `fba_schema_migrations` and a lock key compatible with `fba:migration:lock`.

- [ ] **Step 5: Run tests**

Run: `go test ./core/db ./core/migration`

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
git add core/db core/migration core/config
git commit -m "feat: add db and migration primitives"
```

---

## Chunk 5: Plugin SDK And Registry

### Task 10: Implement plugin interfaces

**Files:**
- Create: `core/plugin/module.go`
- Create: `core/plugin/context.go`
- Create: `core/plugin/route.go`
- Create: `core/plugin/mode.go`
- Test: `core/plugin/module_test.go`

- [ ] **Step 1: Write failing API compile test**

Create a test plugin implementing:

```go
func FBAPlugin() plugin.Module
```

Assert it can register a route, provider, task, migration, and swagger fragment through `plugin.Context`.

- [ ] **Step 2: Run test**

Run: `go test ./core/plugin`

Expected: FAIL.

- [ ] **Step 3: Implement SDK interfaces**

Use the design document names: `Module`, `Meta`, `Dependency`, `Context`, `Route`, `ModeAuto`, `ModeDisabled`, `ModePureDependency`.

- [ ] **Step 4: Run test**

Run: `go test ./core/plugin`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/plugin
git commit -m "feat: define plugin sdk"
```

### Task 11: Implement plugin registry and dependency sort

**Files:**
- Create: `core/plugin/registry.go`
- Create: `core/plugin/graph.go`
- Test: `core/plugin/registry_test.go`
- Test: `core/plugin/graph_test.go`

- [ ] **Step 1: Write failing graph tests**

Cover:

```text
admin -> dict optional
order -> admin required
cycle: order -> payment -> order
missing required dependency
disabled dependency does not satisfy required dependency
pure_dependency satisfies dependency but Register is not called
```

- [ ] **Step 2: Run tests**

Run: `go test ./core/plugin`

Expected: FAIL.

- [ ] **Step 3: Implement registry**

Registry must deduplicate by plugin ID, sort dependencies, expose readable graph errors, and call `Register(ctx)` only for `auto` plugins.

- [ ] **Step 4: Run tests**

Run: `go test ./core/plugin`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/plugin
git commit -m "feat: add plugin registry dependency graph"
```

---

## Chunk 6: Auth, RBAC, Task, Observability Core

### Task 12: Add auth primitives without business routes

**Files:**
- Create: `core/auth/token.go`
- Create: `core/auth/password.go`
- Create: `core/auth/session.go`
- Modify: `core/config/options.go`
- Test: `core/auth/token_test.go`

- [ ] **Step 1: Write failing token tests**

Assert JWT payload contains `session_uuid`, `exp`, and string `sub`; Redis session keys use `redisx` helpers.

- [ ] **Step 2: Run test**

Run: `go test ./core/auth`

Expected: FAIL.

- [ ] **Step 3: Implement token service interfaces and default JWT service**

Do not implement `/api/v1/auth/login` handlers here. The admin plugin consumes these services.

- [ ] **Step 4: Run test**

Run: `go test ./core/auth`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/auth core/config
git commit -m "feat: add auth session primitives"
```

### Task 13: Add RBAC and data permission primitives

**Files:**
- Create: `core/rbac/permission.go`
- Create: `core/rbac/current_user.go`
- Create: `core/rbac/data_permission.go`
- Test: `core/rbac/permission_test.go`

- [ ] **Step 1: Write failing RBAC tests**

Cover whitelist, unauthenticated, super admin, non-staff write denial, missing role menus, and permission code mismatch.

- [ ] **Step 2: Run test**

Run: `go test ./core/rbac`

Expected: FAIL.

- [ ] **Step 3: Implement RBAC evaluator**

Keep DB-backed user/role/menu loading outside core; core accepts `CurrentUser` and route permission metadata.

- [ ] **Step 4: Run test**

Run: `go test ./core/rbac`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/rbac
git commit -m "feat: add rbac primitives"
```

### Task 14: Add task, scheduler, and pool primitives

**Files:**
- Create: `core/task/definition.go`
- Create: `core/task/client.go`
- Create: `core/task/server.go`
- Create: `core/task/scheduler.go`
- Create: `core/pool/provider.go`
- Test: `core/task/definition_test.go`
- Test: `core/pool/provider_test.go`

- [ ] **Step 1: Write failing task registry test**

Assert duplicate task types are rejected and definitions preserve `type`, `name`, `queue`, and handler metadata.

- [ ] **Step 2: Run test**

Run: `go test ./core/task ./core/pool`

Expected: FAIL.

- [ ] **Step 3: Implement Asynq abstractions**

Core owns enqueue, worker registration, leader lock primitive, status mapping, and named Pond pools. The task business plugin owns `/api/v1/tasks`, `/api/v1/task-results`, and `/api/v1/schedulers` handlers.

- [ ] **Step 4: Run test**

Run: `go test ./core/task ./core/pool`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/task core/pool core/config
git commit -m "feat: add task and pool primitives"
```

### Task 15: Add readiness, health, metrics hooks

**Files:**
- Create: `core/observability/health.go`
- Create: `core/observability/readiness.go`
- Create: `core/observability/metrics.go`
- Modify: `core/app/application.go`
- Test: `core/observability/readiness_test.go`

- [ ] **Step 1: Write failing readiness tests**

Assert readiness aggregates DB, Redis, migrations, required plugins, and task components when enabled.

- [ ] **Step 2: Run test**

Run: `go test ./core/observability`

Expected: FAIL.

- [ ] **Step 3: Implement probes and core endpoints**

Register only `/healthz`, `/readyz`, and `/metrics` from core.

- [ ] **Step 4: Run test**

Run: `go test ./core/observability ./core/app`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/observability core/app
git commit -m "feat: add core health readiness metrics"
```

---

## Chunk 7: Swagger And Code Generation

### Task 16: Add Swagger fragment aggregation

**Files:**
- Create: `core/swagger/fragment.go`
- Create: `core/swagger/aggregate.go`
- Create: `core/swagger/handler.go`
- Create: `cmd/fbago/internal/swagger/scan.go`
- Test: `core/swagger/aggregate_test.go`
- Test: `cmd/fbago/internal/swagger/scan_test.go`

- [ ] **Step 1: Write failing aggregation tests**

Cover path merge, schema merge, plugin-prefixed schema names, and duplicate method/path conflict.

- [ ] **Step 2: Run test**

Run: `go test ./core/swagger ./cmd/fbago/internal/swagger`

Expected: FAIL.

- [ ] **Step 3: Implement aggregator and handlers**

Expose handlers for `/docs`, `/openapi`, and `/swagger/doc.json`. Add the first `fbago swagger scan` implementation that reads plugin fragments from the plugin lock file and writes the aggregated OpenAPI document.

- [ ] **Step 4: Run test**

Run: `go test ./core/swagger ./cmd/fbago/internal/swagger`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add core/swagger cmd/fbago/internal/swagger
git commit -m "feat: add swagger aggregation"
```

### Task 17: Implement `fbago plugin scan`

**Files:**
- Create: `cmd/fbago/main.go`
- Create: `cmd/fbago/internal/plugin/manifest.go`
- Create: `cmd/fbago/internal/plugin/scan.go`
- Create: `cmd/fbago/internal/plugin/generate.go`
- Test: `cmd/fbago/internal/plugin/scan_test.go`

- [ ] **Step 1: Write failing scanner tests**

Use `internal/testplugin` fixtures and assert manifest, import scan, and local scan merge into one lock file with deterministic order.

- [ ] **Step 2: Run test**

Run: `go test ./cmd/fbago/internal/plugin`

Expected: FAIL.

- [ ] **Step 3: Implement manifest parser and generator**

Generate:

```text
internal/generated/fba_plugins.gen.go
internal/generated/plugin_graph.dot
internal/generated/plugin_manifest.lock
```

Generated code must call `reg.Add(plugin.FBAPlugin(), plugin.ModeAuto)` or the configured mode.

- [ ] **Step 4: Run test**

Run: `go test ./cmd/fbago/internal/plugin`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add cmd/fbago internal/testplugin
git commit -m "feat: add plugin code generation"
```

### Task 18: Implement contract snapshot and contract test CLI

**Files:**
- Create: `cmd/fbago/internal/contract/contract.go`
- Create: `cmd/fbago/internal/contract/snapshot.go`
- Create: `cmd/fbago/internal/contract/test.go`
- Test: `cmd/fbago/internal/contract/contract_test.go`

- [ ] **Step 1: Write failing contract parser tests**

Assert route/method/path/response/Redis contracts parse from `contracts/*.yaml`.

- [ ] **Step 2: Run test**

Run: `go test ./cmd/fbago/internal/contract`

Expected: FAIL.

- [ ] **Step 3: Implement parser and HTTP contract runner**

The first implementation may validate route existence, envelope fields, cookie/header behavior, and priority endpoints. Add full DTO checks incrementally.

- [ ] **Step 4: Run tests**

Run: `go test ./cmd/fbago/internal/contract ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add cmd/fbago/internal/contract contracts
git commit -m "feat: add api contract tooling"
```

---

## Chunk 8: Example Host App

### Task 19: Build a plugin-driven host example

**Files:**
- Create: `examples/compat-host/go.mod`
- Create: `examples/compat-host/main.go`
- Create: `examples/compat-host/plugins.yaml`
- Create: `examples/compat-host/internal/generated/fba_plugins.gen.go`
- Test: `examples/compat-host/README.md`

- [ ] **Step 1: Create host app skeleton**

The example imports `github.com/yuWorm/fba-go` and generated plugin registration code. Use test fixture plugins until external official plugins exist.

- [ ] **Step 2: Generate plugin registration**

Run:

```bash
go run ./cmd/fbago plugin scan \
  --mode manifest \
  --manifest examples/compat-host/plugins.yaml \
  --out examples/compat-host/internal/generated/fba_plugins.gen.go
```

Expected: generated file lists configured plugins.

- [ ] **Step 3: Run example tests**

Run: `go test ./examples/compat-host/...`

Expected: PASS.

- [ ] **Step 4: Commit**

Run:

```bash
git add examples/compat-host
git commit -m "docs: add plugin driven host example"
```

---

## Chunk 9: Official Business Plugin Track

These tasks are not implemented inside this repository unless the plugin repositories are added as separate worktrees or repositories. They define the first plugin delivery sequence.

### Task 20: Create `fba-plugin-admin`

**Files in external module:**
- Create: `go.mod`
- Create: `plugin.yaml`
- Create: `plugin.go`
- Create: `api/auth_handler.go`
- Create: `api/user_handler.go`
- Create: `api/role_handler.go`
- Create: `api/menu_handler.go`
- Create: `api/dept_handler.go`
- Create: `api/data_scope_handler.go`
- Create: `api/file_handler.go`
- Create: `dto/*.go`
- Create: `service/*.go`
- Create: `repo/*.go`
- Create: `model/*.go`
- Create: `migration/*.sql`

- [ ] **Step 1: Write contract tests for priority auth/sys endpoints**

Run from host app: `fbago contract test --base-url http://127.0.0.1:8001 --contract contracts/api.contract.yaml`

Expected initially: FAIL.

- [ ] **Step 2: Implement login, token refresh, logout, current user, codes, sidebar**

Use core `auth`, `rbac`, `redisx`, `response`, `db`, and `migration`; do not duplicate framework code.

- [ ] **Step 3: Run contract test**

Expected: priority auth/sys endpoints PASS.

- [ ] **Step 4: Continue CRUD endpoints by contract priority**

Implement users, roles, menus, departments, data rules, data scopes, files, plugin status UI endpoints.

### Task 21: Create `fba-plugin-dict`

**Files in external module:**
- Create: `go.mod`
- Create: `plugin.yaml`
- Create: `plugin.go`
- Create: `api/dict_type_handler.go`
- Create: `api/dict_data_handler.go`
- Create: `dto/*.go`
- Create: `service/*.go`
- Create: `repo/*.go`
- Create: `model/*.go`
- Create: `migration/*.sql`

- [ ] **Step 1: Write failing tests for `/api/v1/dict-datas/type-codes/{code}`**

Run: `fbago contract test --base-url http://127.0.0.1:8001 --contract contracts/api.contract.yaml`

Expected: FAIL until dict plugin is registered and implemented.

- [ ] **Step 2: Implement dict type/data CRUD and cache invalidation**

Use Redis pub/sub channel `fba:cache:invalidate` for cache invalidation.

- [ ] **Step 3: Run contract test**

Expected: dict endpoints PASS.

### Task 22: Create `fba-plugin-task`

**Files in external module:**
- Create: `go.mod`
- Create: `plugin.yaml`
- Create: `plugin.go`
- Create: `api/task_handler.go`
- Create: `api/task_result_handler.go`
- Create: `api/scheduler_handler.go`
- Create: `service/scheduler_service.go`
- Create: `repo/*.go`
- Create: `model/*.go`
- Create: `migration/*.sql`

- [ ] **Step 1: Write failing tests for registered tasks and scheduler CRUD**

Run: `fbago contract test --base-url http://127.0.0.1:8001 --contract contracts/api.contract.yaml`

Expected: FAIL.

- [ ] **Step 2: Implement task_result compatibility adapter**

Map Asynq states to `PENDING`, `STARTED`, `SUCCESS`, `RETRY`, and `FAILURE`.

- [ ] **Step 3: Implement scheduler leader behavior**

Use Redis lease key `fba:task:scheduler:leader`.

- [ ] **Step 4: Run contract test**

Expected: task endpoints PASS.

---

## Chunk 10: Integration And Release Gates

### Task 23: Add all-package verification

**Files:**
- Create: `Makefile`
- Create: `.github/workflows/ci.yml` if GitHub Actions is used

- [ ] **Step 1: Add local verification targets**

```makefile
.PHONY: test
test:
	go test ./...

.PHONY: generate
generate:
	go run ./cmd/fbago plugin scan --mode manifest --manifest examples/compat-host/plugins.yaml --out examples/compat-host/internal/generated/fba_plugins.gen.go

.PHONY: contract
contract:
	go run ./cmd/fbago contract test --base-url http://127.0.0.1:8001 --contract contracts/api.contract.yaml
```

- [ ] **Step 2: Run verification**

Run: `make test`

Expected: PASS.

- [ ] **Step 3: Commit**

Run:

```bash
git add Makefile .github
git commit -m "chore: add verification targets"
```

### Task 24: Release readiness checklist

**Files:**
- Create: `docs/release/core-module-readiness.md`

- [ ] **Step 1: Document readiness gates**

Include:

```text
go test ./...
fbago plugin graph
fbago contract snapshot
fbago contract test
Redis single/sentinel smoke test
DB migration dry-run
Scheduler leader lock smoke test
Asynq worker smoke test
Example host boot smoke test
```

- [ ] **Step 2: Document non-goals**

State that runtime `.so` plugins, Redis Cluster support, plugin hot reload, and non-compatible table refactors are not first-release goals.

- [ ] **Step 3: Commit**

Run:

```bash
git add docs/release/core-module-readiness.md
git commit -m "docs: add core module release gates"
```

---

## Execution Order

Implement in this order:

```text
1. Core skeleton and contracts
2. Response/error/pagination/datetime
3. App, Fiber, DI, logger
4. Redis, DB, migration
5. Plugin SDK and registry
6. Auth/RBAC/task/observability primitives
7. Swagger and fbago
8. Example host app
9. External official plugins: admin, dict, task
10. Contract test and HA release gates
```

The first shippable milestone is reached when:

```text
go test ./...
go run ./cmd/fbago plugin scan ...
examples/compat-host boots
core has no business route handlers
```

The first frontend-compatible milestone is reached when:

```text
admin + dict plugins pass priority contract tests
frontend can login
frontend can load current user
frontend can load sidebar menus
frontend can load permission codes
frontend can load dict data by type code
```
