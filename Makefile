PROJECT = vest
GOARCH ?= amd64
BUILD_FLAGS ?= -v
LINK_FLAGS ?= '-d -s -w'
TEST_FLAGS ?= -v

build: darwin linux

test:
	go test $(TEST_FLAGS) ./pkg/...

darwin:
	GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags '-s -w' -o ./bin/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)
linux:
	GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags $(LINK_FLAGS) -o ./bin/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)


