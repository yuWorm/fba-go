---
name: fba-go-template-dev
description: Maintain FBA Go scaffold templates and generated project behavior. Use when changing fbago init, embedded admin or basic templates, remote or local template loading, go.mod.tmpl version replacement, Admin module dependency metadata, template tests, release tags, or behavior under cmd/fbago/internal/scaffold.
---

# FBA Go Template Development

## Workflow

Use this skill to keep generated projects reproducible, release-safe, and
aligned with the independently versioned Admin module.

1. Identify the template surface: embedded `admin` or `basic`, custom local or
   remote templates, the `github.com/yuWorm/fba-go-admin` module, or release
   documentation.
2. Keep the independent Admin module and thin embedded starter explicit.
   `../fba-go-admin/go.mod` owns official source; the embedded
   `cmd/fbago/internal/scaffold/templates/admin/go.mod.tmpl` depends on it.
3. Preserve `[[ .Module ]]`, `[[ .TemplateModule ]]`,
   `[[ .TemplateVersion ]]`, `[[ .TemplateReplace ]]`, core-version fields,
   and template origin fields.
4. Keep `.fbago-template.yaml` exclude and module-preservation boundaries aligned
   when runnable-module files move.
5. For Python-aligned Admin behavior, compare `sources/fastapi-best-architecture/`
   before changing routes, models, migrations, or seed data.
6. Regenerate `internal/generated/fba_plugins.gen.go` and `plugins.lock` when
   official module facades change.
7. Verify both the scaffold package and an actual thin generated project.

## Load References

- Read `references/scaffold-generation.md` before changing `fbago init`, template parsing, remote Git templates, or version replacement.
- Read `references/admin-template.md` before changing the official admin template or generated project behavior.
- Read `references/python-alignment.md` before migrating or aligning Python admin behavior.

## Guardrails

- Do not write `@latest` into generated `go.mod`. Use a resolved semver or pseudo-version.
- Keep local development safe with `v0.0.0 + replace`.
- Do not copy repository metadata or local build artifacts into generated projects.
- Keep embedded `fbago-template.yaml` metadata internal to the scaffold; it
  declares the released Admin module path and version.
- Do not remove runnable-template `.fbago-template.yaml`; custom local and
  remote templates need it for source-module rewrite boundaries.
- Never copy official app/plugin implementation directories into generated
  projects. Version them through the Admin Go module.
- Keep project `internal/app` business source outside template management.
- `plugins.yaml` is project-owned; registration and lock files are refreshed by
  `fbago plugin sync`.
- Template update may manage only thin bootstrap files and must still refuse
  unsafe modified/removed files without explicit force.

## Verification

Run targeted scaffold and Admin module tests, then exercise the actual thin
generated project:

```bash
GOWORK=off GOCACHE=/private/tmp/fba-go-gocache go test ./cmd/fbago ./cmd/fbago/internal/scaffold
make -C ../fba-go-admin test
make verify-template
```

After manifest or sync changes, generate an Admin project, confirm official
implementation directories are absent, run `fbago plugin sync --check`,
`go test ./...`, `go run ./cmd/api --help`, and verify template diff is clean.
