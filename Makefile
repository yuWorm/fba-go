GO ?= go
GOCACHE_DIR ?= /private/tmp/fba-go-build-cache
GOENV := GOCACHE=$(GOCACHE_DIR)

.PHONY: test
test:
	$(GOENV) $(GO) test ./...

.PHONY: verify-template
verify-template:
	scripts/verify-template.sh
