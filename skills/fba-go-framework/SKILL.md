---
name: fba-go-framework
description: Understand, modify, or review the FBA Go core framework. Use when working on core application startup, configuration, Fiber runtime, middleware, response contracts, CLI command execution, task runtime contracts, database providers, realtime support, observability, or cross-cutting behavior under core/, fba.go, or template runtime wiring.
---

# FBA Go Framework

## Workflow

Use this skill to keep framework changes aligned with the current FBA Go architecture instead of inventing parallel contracts.

1. Inspect the relevant core package before editing. Prefer codegraph context for symbols, then read the specific files listed in the references.
2. Identify whether the change belongs in core, admin template runtime, or a plugin. Core should expose stable contracts; template runtime should compose them; plugins should declare capabilities through `plugin.Context`.
3. Preserve Python-compatible API and environment behavior when the touched surface was migrated from `sources/fastapi-best-architecture/`.
4. Add or update tests near the package that owns the contract. For cross-cutting behavior, run root `go test ./...` and admin template `make L=1 test`.

## Load References

- Read `references/architecture.md` for application, DI, plugin, routing, migration, and runtime boundaries.
- Read `references/runtime-config.md` for `.env`, defaults, CORS, database, Redis, realtime, and startup behavior.
- Read `references/cli-response-task.md` for CLI command composition, response envelopes, errors, and task contracts.

## Guardrails

- Do not expose driver-specific details from core contracts unless the package already owns that driver integration.
- Keep `core/response` envelopes compatible with admin frontend expectations: `code`, `msg`, `data`, and optional `trace_id` for errors.
- Keep `core/command` plugin-safe: plugins register `command.Command`; runtime decides default command and outputs.
- Keep local development behavior distinct from release behavior. Generated projects may use `v0.0.0 + replace`; released CLI binaries should use semver module versions.
- Keep comments for non-obvious constraints, especially defaults that mirror the Python project or production behavior.

## Verification

Run the smallest relevant tests first, then broaden:

```bash
GOWORK=off GOCACHE=/private/tmp/fba-go-gocache go test ./...
make -C ../fba-go-admin test
```
