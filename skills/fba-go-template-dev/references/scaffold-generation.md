# Scaffold Generation

## Source Map

- CLI entry: `cmd/fbago/main.go`
- Scaffold init: `cmd/fbago/internal/scaffold/init.go`
- Embedded templates: `cmd/fbago/internal/scaffold/templates/basic` and `admin`
- Scaffold tests: `cmd/fbago/internal/scaffold/*_test.go`
- CLI tests: `cmd/fbago/main_test.go`
- Versioned Admin module checkout: `../fba-go-admin`

## Init Flow

`fbago init` parses:

- module argument
- `--dir`
- `--template`
- `--template-replace`
- `--force`
- `--core-replace`
- `--core-version`

`scaffold.Init` then:

1. validates module name
2. resolves template name or path
3. loads embedded, local, or remote Git template files
4. rejects overwriting existing files unless forced
5. computes template data
6. reads embedded dependency metadata or applies custom-template
   `.fbago-template.yaml` rewrite boundaries
7. rewrites custom template-module imports unless a path opts out
8. renders `.tmpl` files and formats generated Go
9. writes files
10. when `plugins.yaml` exists, runs plugin sync to generate registration,
    tidy module dependencies, and write the version-aware lock

## Template Data

Template files can use:

- `[[ .Module ]]`: target module name.
- `[[ .TemplateModule ]]`: versioned module that supplies template capabilities.
- `[[ .TemplateName ]]`: resolved template name.
- `[[ .TemplateSource ]]`: `embedded`, `local`, or `remote`.
- `[[ .TemplateRepo ]]`: best-effort Git origin URL for local/remote templates.
- `[[ .TemplateRef ]]`: remote template ref when provided.
- `[[ .TemplateCommit ]]`: best-effort Git commit for local/remote templates.
- `[[ .TemplatePath ]]`: template root path relative to the template repository.
- `[[ .CoreReplace ]]`: optional replace path for local core module.
- `[[ .CoreVersion ]]`: resolved FBA Go core module version.
- `[[ .TemplateVersion ]]`: version from embedded metadata or a remote release ref.
- `[[ .TemplateReplace ]]`: local module replacement for custom templates or development checkouts.

Keep delimiters as `[[` and `]]` because Go module files and other template content may contain normal braces.

## Embedded Template Rules

`admin` is the default embedded template; `basic` remains available explicitly
through `--template basic`. Embedded `admin/fbago-template.yaml` is scaffold
metadata, not generated-project content. It declares:

- the versioned Admin module path
- the semantic version emitted by released `fbago` binaries
- no repository-relative development checkout path

Embedded Admin scaffolds use the declared semantic version without a
replacement. Local integration must opt in with `--template-replace` or
`FBAGO_TEMPLATE_REPLACE`; that combination renders `v0.0.0 + replace`.
Only the thin project skeleton is embedded; Admin implementation packages stay
in the versioned module.

## Project Manifest Rules

Generated projects must include a root `.fbago.yaml` when a template expects
future updates. The file is rendered from `fbago.yaml.tmpl`; the template file is
stored without a leading dot because embedded template directories do not carry
dotfiles reliably.

Manifest v1 records template origin and the small bootstrap surface that remains
template-managed. Official Admin implementation source is a versioned module and
must not appear under generated-project managed paths:

```yaml
version: 1

template:
  name: admin
  module: github.com/your-org/my-admin
  source_module: github.com/yuWorm/fba-go-admin
  source: embedded
  template_path: admin
  core_version: v0.1.0

managed:
  - name: bootstrap
    kind: generated
    mode: source
    path: cmd/api/main.go
    source_path: cmd/api/main.go
```

Template origin fields:

- `template.name`: stable template name, such as `basic` or `admin`.
- `template.module`: generated project module.
- `template.source_module`: runnable template module from `.fbago-template.yaml`.
- `template.source`: `embedded`, `local`, or `remote`.
- `template.repo`: best-effort Git origin URL for local templates, or clone URL for remote templates.
- `template.ref`: remote template ref when the caller specified one.
- `template.commit`: best-effort source commit. This is metadata, not the update selector.
- `template.template_path`: template root path inside the source repository, such as `admin`.
- `template.core_version`: generated core module version.

Managed entry fields:

- `name`: stable logical file group.
- `kind`: normally `generated` for thin bootstrap files.
- `mode`: `source` for template-managed source.
- `path`: target path in the generated project.
- `source_path`: rendered source path in the template.

Official modules, project `internal/app`, `plugins.yaml`, and generated plugin
registration are not template-update surfaces. Use Go module versions,
project-owned source, and `fbago plugin sync` respectively.

## Template Diff / Update Rules

`fbago template diff`:

1. reads project `.fbago.yaml`
2. resolves the source template from the manifest unless `--template` overrides it
3. renders the template with recorded module versions
4. compares only thin managed bootstrap paths
5. prints `A`, `M`, or `D` file-level changes

`fbago template update` follows the same plan and then writes changes:

- `--dry-run` never writes files.
- New managed files (`A`) may be written without `--force`.
- Modified managed files (`M`) must require `--force`.
- Removed managed files (`D`) must require `--force`.
- The manifest must be written after managed files, so a partial failure does not
  record a new template state before source files are updated.
- Use `--template <local-template-path>` when testing local template checkout changes.

Deletion semantics:

- If a managed entry disappears from the new template manifest, report `D` for
  files under the old managed path.
- Do not delete files by default; require `--force` because the old path may
  contain project business changes.
- If the removed path overlaps another still-managed path, do not plan deletion
  for the overlapping path.

Safety constraint:

- There is currently no per-file baseline hash. Treat every existing file that
  differs from the freshly rendered template as potentially project-modified.
  This is why update requires explicit `--force` for `M` and `D`.

## Core Version Rules

Generated `go.mod` must use a concrete module version. Never render `@latest` into `go.mod`.

Resolution rules:

- Explicit `--core-version vX.Y.Z` wins.
- `FBAGO_CORE_VERSION` is used when the flag is absent.
- Explicit `--core-version latest` resolves through `go list`.
- When a local replace exists, use `v0.0.0` because replace makes the selected version irrelevant.
- Release builds use the binary build version when no replace is present.
- Development fallback is `v0.0.0`.

Local development should produce `v0.0.0 + replace`. Published use should produce a semver version.

## Core Replace Rules

`CoreReplace` comes from:

1. `--core-replace`
2. `FBAGO_CORE_REPLACE`
3. auto-discovered local module root for development builds

Released binaries should not force a local replace.


## Template Replace Rules

`TemplateReplace` comes from:

1. `--template-replace`
2. `FBAGO_TEMPLATE_REPLACE`
3. the loaded template root when `--template` selects a local template

An explicit flag wins over the environment and local-template default. The
embedded Admin starter never guesses a sibling checkout. Any replacement uses
`v0.0.0` because Go resolves source from the replacement path.

## Local Template Rules

Local templates are expected to be runnable repositories. They may include:

- real `go.mod` for testing the template itself
- `.fbago-template.yaml` with source module
- `go.mod.tmpl` for generated projects
- `.tmpl` files for renderable content
- `exclude` paths that belong only to the runnable/versioned template module
- `preserve_module_paths` whose imports must keep pointing at that module

Skipped directories:

- `.cache`
- `.codegraph`
- `.git`
- `.hg`
- `.svn`
- `bin`
- `node_modules`
- `tmp`

Skipped files:

- `.DS_Store`
- `.fbago-template.yaml`
- `Thumbs.db`

## Remote Git Templates

Supported forms:

```bash
github.com/acme/custom-template/admin@admin/v0.1.0
https://github.com/acme/custom-template.git//admin@admin/v0.1.0
git+https://github.com/acme/custom-template.git//admin@admin/v0.1.0
git@github.com:acme/custom-template.git//admin@admin/v0.1.0
```

Subdirectory Go modules use tags such as `admin/v0.1.0`; the generated
`go.mod` receives the semantic suffix `v0.1.0`.

## Verification

Scaffold package:

```bash
GOWORK=off GOCACHE=/private/tmp/fba-go-gocache go test ./cmd/fbago ./cmd/fbago/internal/scaffold
```

Generated admin smoke test:

```bash
make verify-template
```
