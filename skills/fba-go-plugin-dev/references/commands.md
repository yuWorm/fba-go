# Plugin Commands

## Source Map

- Command contract: `core/command/command.go`
- Admin runtime command aggregation: `../fba-go-admin/internal/runtime/runtime.go`
- Task plugin command example: `../fba-go-admin/plugins/task/module.go`

## Command Shape

Plugins register commands with:

```go
ctx.Command(command.Command{
	Use:   "task reload",
	Short: "Reload task runtime",
	Run: func(ctx context.Context, runtime command.Runtime, args []string) error {
		return executor.Reload(ctx)
	},
})
```

`Use` is space-separated. `core/command` builds nested Cobra commands from it.

## Runtime Access

Command handlers receive `command.Runtime`, which exposes:

- `Container()`
- `Config()`
- `Output()`
- `ErrorOutput()`

Use these instead of global variables or direct `os.Stdout` writes.

## Default Command

The admin runtime sets `DefaultCommand: "server"`, so running the binary without arguments starts the server.

Plugins must not change the default command. Register explicit subcommands instead.

## Design Rules

- Keep commands narrow and operational.
- Return errors instead of calling `os.Exit`.
- Write command output to `runtime.Output()`.
- Use `DisableFlagParsing` only for passthrough commands that own their parser.
- Prefer command paths that start with the plugin ID, such as `task reload`.

## Tests

Test command behavior through `core/command.Execute` or the admin runtime with injected output buffers. Avoid tests that depend on process exit.
