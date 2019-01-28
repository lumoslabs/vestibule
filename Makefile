PROJECT = vest
GOARCH ?= amd64
BUILD_FLAGS ?= -v
LINK_FLAGS ?= '-d -s -w'

build: darwin linux

darwin:
	GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags '-s -w' -o ./bin/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)
linux:
	GOOS=$@ GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -ldflags $(LINK_FLAGS) -o ./bin/$(PROJECT)-$@-$(GOARCH) ./cmd/$(PROJECT)


