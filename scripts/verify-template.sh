#!/usr/bin/env bash
set -euo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
core_candidate="${FBA_GO_ROOT:-${repository_root}}"
generated_parent="${FBAGO_VERIFY_OUT:-}"
keep_generated="${FBAGO_VERIFY_KEEP:-0}"

if ! core_root="$(cd "${core_candidate}" 2>/dev/null && pwd)"; then
	echo "FBA_GO_ROOT must point to a fba-go checkout; got ${core_candidate}" >&2
	exit 1
fi
if [[ ! -d "${core_root}/cmd/fbago" ]]; then
	echo "FBA_GO_ROOT must point to a fba-go checkout; got ${core_root}" >&2
	exit 1
fi

admin_root=""
if [[ -n "${FBA_ADMIN_ROOT:-}" ]]; then
	admin_candidate="${FBA_ADMIN_ROOT}"
elif [[ -f "${core_root}/../fba-go-admin/go.mod" ]]; then
	admin_candidate="${core_root}/../fba-go-admin"
else
	admin_candidate=""
fi
if [[ -n "${admin_candidate}" ]]; then
	if ! admin_root="$(cd "${admin_candidate}" 2>/dev/null && pwd)"; then
		echo "FBA_ADMIN_ROOT must point to a fba-go-admin checkout; got ${admin_candidate}" >&2
		exit 1
	fi
	if [[ ! -f "${admin_root}/go.mod" ]]; then
		echo "FBA_ADMIN_ROOT must point to a fba-go-admin checkout; got ${admin_root}" >&2
		exit 1
	fi
fi

if ! command -v go >/dev/null 2>&1; then
	echo "go is required" >&2
	exit 1
fi

export GOCACHE="${GOCACHE:-/private/tmp/fba-go-gocache}"

if [[ -z "${generated_parent}" ]]; then
	generated_parent="$(mktemp -d)"
else
	mkdir -p "${generated_parent}"
fi

cleanup() {
	if [[ "${keep_generated}" != "1" ]]; then
		rm -rf "${generated_parent}"
	fi
}
trap cleanup EXIT

generated_dir="${generated_parent}/admin-generated"
rm -rf "${generated_dir}"

init_args=(
	init
	github.com/acme/fbago-admin-generated
	--dir "${generated_dir}"
	--core-replace "${core_root}"
)
if [[ -n "${admin_root}" ]]; then
	init_args+=(--template-replace "${admin_root}")
fi

echo "==> generating embedded Admin starter"
(
	cd "${core_root}"
	GOWORK=off go run ./cmd/fbago "${init_args[@]}"
)

echo "==> verifying generated registration"
(
	cd "${core_root}"
	GOWORK=off go run ./cmd/fbago plugin sync --dir "${generated_dir}" --check
)

echo "==> testing generated project"
(
	cd "${generated_dir}"
	GOWORK=off go test ./...
	GOWORK=off go build ./cmd/api
	GOWORK=off go run ./cmd/api --help >/dev/null
)

echo "==> checking template drift"
(
	cd "${core_root}"
	GOWORK=off go run ./cmd/fbago template diff --dir "${generated_dir}"
)

echo "==> embedded Admin starter verified"
