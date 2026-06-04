# Compatibility Host Example

This example shows a host application that imports `github.com/yuWorm/fba-go`, loads generated plugin registration code, and registers a fixture plugin.

Generate plugin registration:

```bash
go run ./cmd/fbago plugin scan \
  --mode manifest \
  --manifest examples/compat-host/plugins.yaml \
  --out examples/compat-host/internal/generated/fba_plugins.gen.go
```

Run tests:

```bash
go test ./examples/compat-host/...
```
