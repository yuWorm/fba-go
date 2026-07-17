# Task Integration

## Source Map

- Stable task runtime: `core/task/runtime.go`
- Asynq task definitions: `core/task/definition.go`
- Asynq client helper: `core/task/client.go`
- Serve mux builder: `core/task/server.go`
- Scheduler compatibility: `core/task/scheduler.go`
- Task plugin module: `../fba-go-admin/plugins/task/module.go`
- Task plugin service: `../fba-go-admin/plugins/task/service/*`

## Stable Contract

Business code should depend on `core/task.Runtime`:

```go
type Runtime interface {
	Reload(context.Context) error
	Execute(ctx context.Context, task string, args any, kwargs any) error
	Cancel(ctx context.Context, taskID string) error
}
```

This keeps business modules independent from Asynq, Temporal, or any future queue engine.

## Fallback Behavior

The task plugin uses `coretask.NoopRuntime` if no runtime is provided through DI. This lets the admin template start without Redis or a worker.

When adding task features, preserve no-op behavior unless the feature explicitly requires a configured runtime.

## Task Definitions

Asynq handlers use `core/task.Definition`:

```go
coretask.Definition{
	Type:    "email.send",
	Name:    "Send email",
	Queue:   "default",
	Handler: handler,
}
```

Register definitions in a `coretask.Registry` when building a worker. Use `BuildServeMux` to produce an Asynq mux.

## Scheduler Plugin

The admin task plugin provides:

- registered task listing
- scheduler CRUD
- scheduler status toggles
- manual execution
- task result listing
- cancel by task ID

It resolves:

- task definition registry
- task runtime executor
- realtime hub
- Redis leader lease when Redis exists

## Leader Lease

Scheduler leadership defaults to `NoopLeaderLease`. When Redis is configured, the task plugin uses `NewRedisLeaderLease` with configured TTL and Redis key prefix.

Keep scheduler leadership optional in local tests.
