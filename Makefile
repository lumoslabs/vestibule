PROJECT = vest
GOARCH ?= amd64
REF ?= $(shell git rev-parse --abbrev-ref HEAD)
SHA ?= $(shell git rev-parse --short=8 HEAD)
BUILD_FLAGS ?= -v
LINK_FLAGS ?= '-s -w -X main.version=$(REF) -X main.commit=$(SHA) -X main.date=$(shell date +%F)'
TEST_FLAGS ?= -v
OUT_DIR ?= ./bin

PKG_LIST := $(shell go list ./...)

.PHONY: snapshot release test test-race test-memory test-all linux darwin

snapshot:
	@goreleaser release --skip-publish --snapshot --rm-dist

release:
	@goreleaser release --rm-dist

test:
	@go test $(TEST_FLAGS) $(PKG_LIST)

test-race:
	@go test -race -short $(PKG_LIST)

test-memory:
	@go test -msan -short $(PKG_LIST)

test-all: test test-race test-memory

linux darwin:
	@echo "==> Building $(PROJECT)-$@-$(GOARCH)		ref=$(REF) sha=$(SHA) out=$(OUT_DIR)/$(PROJECT)-$@-$(GOARCH)"
	@GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags $(LINK_FLAGS) -o $(OUT_DIR)/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)


