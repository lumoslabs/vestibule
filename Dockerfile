FROM golang:1.11-alpine3.8 AS build

RUN apk add --no-cache file git

ENV CGO_ENABLED=0 \
  BUILD_FLAGS="-v -ldflags '-d -s -w'" \
  GO111MODULE=on
WORKDIR /build
COPY go.mod ./
RUN go mod download
COPY . ./
RUN eval "GOARCH=amd64 go build $BUILD_FLAGS -o /go/bin/vest ./cmd/vest"

FROM alpine:3.8
COPY --from=build /go/bin/vest /bin/vest
ENV USER=root
RUN { \
  echo '#!/usr/bin/dumb-init /bin/sh'; \
  echo 'exec /bin/vest $@'; \
  } >/entrypoint.sh \
  && chmod 755 /entrypoint.sh \ 
  && apk add --update dumb-init
ENTRYPOINT [ "/entrypoint.sh" ]