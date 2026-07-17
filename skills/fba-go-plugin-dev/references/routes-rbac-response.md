# Routes, RBAC, and Responses

## Source Map

- Route declarations: `core/plugin/route.go`
- Route mounting and auth wrapper: `core/plugin/mount.go`
- RBAC contracts: `core/rbac/*`
- Response helpers: `core/response/response.go`
- Error mapping: `core/middleware/error_handler.go`
- Admin routes: `../fba-go-admin/internal/app/admin/api/routes.go`
- Task routes: `../fba-go-admin/plugins/task/api/routes.go`

## Route Declaration

Prefer route groups that return `[]plugin.Route`:

```go
func Routes(h Handler) []plugin.Route {
	return []plugin.Route{
		plugin.GET("/tasks/registered", "Registered tasks", h.RegisteredTasks, plugin.Auth()),
		plugin.POST("/schedulers/:pk/execute", "Execute scheduler", h.ExecuteScheduler, plugin.Auth(), plugin.Perm("sys:task:exec")),
	}
}
```

Then register them in the module:

```go
handler := taskapi.NewHandler(svc)
return plugin.RegisterRoutes(ctx, taskapi.Routes(handler))
```

## Auth Options

- `plugin.Auth()`: require a valid authenticated user.
- `plugin.Perm("code")`: require RBAC permission code.
- `plugin.Superuser()`: require superuser.

`plugin.Perm` does not automatically set `AuthRequired`, so call it with `plugin.Auth()` unless the helper is changed. Existing routes usually use both.

## Authenticator

The admin module provides a `plugin.Authenticator` from its handler:

```go
ctx.Provide(func() plugin.Authenticator {
	return handler
})
```

`MountRoutes` resolves this authenticator from DI. Protected routes fail with a response envelope and request trace ID when auth fails.

## Handler Pattern

Handlers should:

1. Bind request params with Fiber binding.
2. Call service methods with `c.RequestCtx()`.
3. Return `c.JSON(response.Success(data))` for success.
4. Return errors directly for middleware mapping.

Do not hand-roll success JSON shapes.

## Response Contract

Success:

```go
return c.JSON(response.Success(result))
```

Error:

```go
return errors.New(fiber.StatusBadRequest, 400, "message", err)
```

The middleware maps errors to:

```json
{"code":400,"msg":"message","data":null,"trace_id":"..."}
```

## Python Alignment

When migrating admin API behavior, compare:

- route path
- HTTP method
- permission code
- auth requirement
- request and response DTO fields
- pagination shape
- status and error messages

Use `sources/fastapi-best-architecture/` as the behavior source.
