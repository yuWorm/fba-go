GO ?= go
APP ?= fba-api
BIN_DIR ?= bin
GOCACHE_DIR ?= .cache/go-build
GOENV := GOWORK=off GOCACHE=$(CURDIR)/$(GOCACHE_DIR)

.PHONY: tidy
tidy:
	$(GOENV) $(GO) mod tidy

.PHONY: test
test:
	$(GOENV) $(GO) test ./...

.PHONY: run
run:
	$(GOENV) $(GO) run ./cmd/api

.PHONY: dev
dev: run

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	$(GOENV) $(GO) build -o $(BIN_DIR)/$(APP) ./cmd/api

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) .cache
