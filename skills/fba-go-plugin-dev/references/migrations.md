# Plugin Migrations

## Source Map

- Core migration contracts: `core/migration/*`
- Admin runtime migration execution: `../fba-go-admin/internal/runtime/runtime.go`
- Admin migration examples: `../fba-go-admin/internal/app/admin/migration/*`
- Task migration examples: `../fba-go-admin/plugins/task/migration/*`
- SQL seed examples: `../fba-go-admin/internal/app/admin/migration/sql/*`

## Registration Pattern

Register migrations only when a write database is available:

```go
var provider db.Provider
if ctx.Container().Resolve(&provider) && provider != nil && provider.Write() != nil {
	repository = repo.NewGORMRepository(provider)
	if err := ctx.Migration(migration.AutoMigrate(provider)); err != nil {
		return err
	}
	if err := ctx.Migration(migration.InitialData(provider)); err != nil {
		return err
	}
}
```

This keeps memory-only tests and no-DB local runs working.

## SQL as Production Baseline

Use SQL migrations for production seed data and initialization that must be stable across releases. This is especially important for:

- menus
- APIs
- users
- roles
- permissions
- dictionaries
- plugins
- scheduler seed rows

When migration data comes from Python, prefer preserving table shape, IDs, permission codes, and timestamps unless there is a deliberate compatibility break.

## Database Compatibility

The Admin module carries SQL under driver-specific directories:

- `sql/mysql`
- `sql/postgresql`
- `sql/sqlite`

Keep SQL dialect details in the driver directory. Do not rely on one dialect's syntax in another directory.

For MySQL text containing emoji or 4-byte Unicode, ensure table and connection charset support `utf8mb4`. If a seed row is not essential, prefer avoiding decorative emoji in production seed content.

## Idempotency

Migrations should be safe to run once and tracked by the migration runner. Seed SQL should avoid accidental duplicate rows if the migration can be retried after partial failure.

Use stable migration versions and scopes. Do not rewrite an already-released migration unless the release process explicitly allows it; add a new migration instead.

## Verification

At minimum:

```bash
make -C ../fba-go-admin test
```

For driver-specific SQL, test with the real driver when possible. SQLite tests do not prove MySQL or PostgreSQL SQL syntax.
