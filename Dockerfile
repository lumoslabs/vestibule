FROM golang:1.11-alpine3.8

RUN apk add --no-cache file git

ENV CGO_ENABLED=0 \
  BUILD_FLAGS="-v -ldflags '-d -s -w'" \
  GO111MODULE=on
WORKDIR /build
COPY go.mod ./
RUN go get -v
COPY . ./


RUN set -eux; \
  eval "GOARCH=amd64 go build $BUILD_FLAGS -o /go/bin/vest ./cmd/vest"; \
  file /go/bin/vest; \
  /go/bin/vest --version; \
  /go/bin/vest nobody id; \
  /go/bin/vest nobody ls -l /proc/self/fd/1; \
  /go/bin/vest --help