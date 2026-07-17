# Python Alignment

## Source Map

- Python reference: `sources/fastapi-best-architecture/`
- Admin Go module: `../fba-go-admin`
- Contracts: `contracts/`
- Existing migration design: `docs/fba_go_module_migration_ha_design.md`

## Alignment Rule

When migrating or adjusting admin behavior, treat the Python project as the behavior source unless the user explicitly requests a Go-specific change.

Align:

- route path
- HTTP method
- auth requirement
- permission code
- request DTO names and JSON fields
- response fields
- pagination fields
- enum and dictionary values
- seed data IDs and codes used by frontend
- database table and column names
- soft-delete behavior
- operation log behavior
- error messages when frontend depends on them

Do not copy Python structure blindly. Preserve behavior while using Go package boundaries.

## SQL Baseline

For initialization data, prefer SQL migrations as the production baseline. This is especially true for:

- menu trees
- API records
- role/menu bindings
- users
- dictionaries
- notices
- plugin metadata
- scheduler presets

Generated seed data should be deterministic and upgradeable.

## Testing Against Python Contracts

When Python tests exist for the behavior, port the relevant expectations into Go tests. It is acceptable to keep coverage focused rather than exhaustive, but the important contracts should match.

Check:

- response code and message
- presence and shape of `data`
- auth and permission failures
- list ordering and pagination
- default seed rows
- status transitions

## Known Compatibility Details

- CORS defaults mirror Python-compatible enabled behavior.
- API base path defaults to `/api/v1`.
- Response success message is `请求成功`.
- Error envelope includes optional `trace_id`.
- Captcha image data is returned as raw base64 payload; frontend adds image data prefix.
- Local template testing uses `make L=1`.

## Migration Safety

For released SQL migrations, add a new migration instead of rewriting an existing one unless the release has not shipped.

When changing seed data for MySQL, verify charset assumptions for text fields that may contain 4-byte Unicode.
