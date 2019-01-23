FROM golang:1.11-alpine3.8

RUN apk add --no-cache file git

ENV CGO_ENABLED=0 \
  BUILD_FLAGS="-v -ldflags '-d -s -w'" \
  GO111MODULE=on
WORKDIR /go/src/github.com/lumoslabs/vestibule
COPY ["go.mod", "*.go", "./"]

RUN set -eux; \
  eval "GOARCH=amd64 go build $BUILD_FLAGS -o /go/bin/vest"; \
  file /go/bin/vest; \
  /go/bin/vest --version; \
  /go/bin/vest nobody id; \
  /go/bin/vest nobody ls -l /proc/self/fd