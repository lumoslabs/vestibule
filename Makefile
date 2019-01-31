PROJECT = vest
GOARCH ?= amd64
REF ?= $(shell git rev-parse --abbrev-ref HEAD)
SHA ?= $(shell git rev-parse --short=8 HEAD)
BUILD_FLAGS ?= -v
LINK_FLAGS ?= '-s -w -X main.Ref=$(REF) -X main.Sha=$(SHA)'
TEST_FLAGS ?= -v
OUT_DIR ?= ./bin

PKG_LIST := $(shell go list ./...)

build: linux darwin

test:
	@go test $(TEST_FLAGS) $(PKG_LIST)

test-race:
	@go test -race -short $(PKG_LIST)

test-memory:
	@go test -msan -short $(PKG_LIST)

linux darwin:
	@echo "==> Building $(PROJECT)-$@-$(GOARCH)		ref=$(REF) sha=$(SHA) out=$(OUT_DIR)/$(PROJECT)-$@-$(GOARCH)"
	@GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags $(LINK_FLAGS) -o $(OUT_DIR)/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)


