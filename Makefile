GO ?= go
GOCACHE_DIR ?= /private/tmp/fba-go-build-cache
GOENV := GOCACHE=$(GOCACHE_DIR)

.PHONY: test
test:
	$(GOENV) $(GO) test ./...
	$(GOENV) $(GO) test ./plugins/fba-plugin-admin/...
	$(GOENV) $(GO) test ./plugins/fba-plugin-dict/...
	$(GOENV) $(GO) test ./plugins/fba-plugin-notice/...
	$(GOENV) $(GO) test ./plugins/fba-plugin-task/...
	$(GOENV) $(GO) test ./examples/compat-host/...

.PHONY: generate
generate:
	$(GOENV) $(GO) run ./cmd/fbagen plugin scan --mode manifest --manifest examples/compat-host/plugins.yaml --out examples/compat-host/internal/generated/fba_plugins.gen.go

.PHONY: contract
contract:
	@set -e; \
	log=$$(mktemp); \
	$(GOENV) $(GO) run ./examples/compat-host > "$$log" 2>&1 & \
	pid=$$!; \
	trap 'kill "$$pid" 2>/dev/null || true; rm -f "$$log"' EXIT; \
	for _ in $$(seq 1 50); do \
		if curl -sf http://127.0.0.1:8000/readyz >/dev/null; then \
			break; \
		fi; \
		sleep 0.1; \
	done; \
	if ! curl -sf http://127.0.0.1:8000/readyz >/dev/null; then \
		cat "$$log"; \
		exit 1; \
	fi; \
	$(GOENV) $(GO) run ./cmd/fbagen contract test --base-url http://127.0.0.1:8000 --contract contracts

.PHONY: contract-db
contract-db:
	@set -e; \
	log=$$(mktemp); \
	FBA_COMPAT_DB=sqlite FBA_COMPAT_SQLITE_DSN="file:fba_contract?mode=memory&cache=shared" $(GOENV) $(GO) run ./examples/compat-host > "$$log" 2>&1 & \
	pid=$$!; \
	trap 'kill "$$pid" 2>/dev/null || true; rm -f "$$log"' EXIT; \
	for _ in $$(seq 1 50); do \
		if curl -sf http://127.0.0.1:8000/readyz >/dev/null; then \
			break; \
		fi; \
		sleep 0.1; \
	done; \
	if ! curl -sf http://127.0.0.1:8000/readyz >/dev/null; then \
		cat "$$log"; \
		exit 1; \
	fi; \
	$(GOENV) $(GO) run ./cmd/fbagen contract test --base-url http://127.0.0.1:8000 --contract contracts
