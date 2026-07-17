# CLI, Response, Error, and Task Contracts

## CLI Source Map

- Core CLI: `core/command/command.go`
- Admin runtime commands: `../fba-go-admin/internal/runtime/runtime.go`
- Plugin command example: `../fba-go-admin/plugins/task/module.go`
- fbago developer CLI: `cmd/fbago/main.go`

## CLI Contract

`core/command.Command` contains:

- `Use`
- `Short`
- `Long`
- `Aliases`
- `DisableFlagParsing`
- `Run`

`command.Execute` builds a Cobra root and installs nested commands from space-separated `Use` paths. Example: `migrate up` creates a `migrate` parent and an `up` command.

The application runtime owns:

- command root name
- default command
- output and error output writers
- runtime object passed to handlers

Plugins should only register `command.Command` through `ctx.Command`.

## Default Admin Commands

The admin runtime registers:

- `server`: start HTTP server, optionally running migrations first when `DATABASE_AUTO_MIGRATE` is enabled.
- `migrate up`: run registered migrations.
- `migrate status`: print status for registered migrations.

Then it appends commands collected from plugin registration. The task plugin currently adds `task reload`.

## Response Envelope

Success responses use:

```go
response.Success(data)
```

JSON shape:

```json
{"code":200,"msg":"请求成功","data":{}}
```

Generic failures can use `response.Fail`, but handlers should usually return an error and let middleware map it.

Error responses use:

```go
response.Error(code, msg, traceID)
```

JSON shape:

```json
{"code":400,"msg":"message","data":null,"trace_id":"..."}
```

`trace_id` is optional and comes from request ID middleware.

## Error Mapping

`core/middleware.ErrorHandler` maps:

- `*errors.AppError` to its HTTP status, code, and public message.
- `*fiber.Error` to Fiber status and message.
- unknown errors to HTTP 500 with `内部服务器错误`.

Use `errors.New(httpStatus, code, message, cause)` when business logic needs a public API error.

## Task Contracts

Core task contracts are intentionally split:

- `core/task.Runtime`: stable business-facing runtime for reload, execute, and cancel.
- `core/task.Definition`: Asynq-backed task handler definition.
- `core/task.Registry`: stores task definitions by type.
- `core/plugin.TaskDefinition`: lightweight plugin declaration currently collected by plugin context.

Avoid leaking Asynq APIs into business services unless the package is specifically implementing Asynq integration.

## Task Runtime Behavior

`core/task.NoopRuntime` is the safe fallback. It validates required task names and IDs but does not enqueue work.

The task plugin resolves `coretask.Runtime` from DI if available, otherwise falls back to `NoopRuntime`. This lets generated projects start without Redis or a task worker.

## Asynq Helpers

`core/task.AsynqClient` wraps enqueue and cancel behavior:

- `Enqueue` JSON-encodes payload and enqueues an Asynq task.
- `Cancel` requires an inspector and cancels processing by task ID.

`BuildServeMux` registers handlers from a `core/task.Registry`.
