# syntax=docker/dockerfile:1.0-experimental
ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine3.11
RUN apk update && apk add ca-certificates git gcc build-base
RUN go get github.com/go-bindata/go-bindata/...
WORKDIR /app
COPY . .
ARG LD_FLAGS
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go-bindata -o src/server/cmd/worker/assets/assets.go -pkg assets /etc/ssl/certs/... && \
    go build -ldflags "${LD_FLAGS}" -o pachd "src/server/cmd/pachd/main.go" && \
    go build -ldflags "${LD_FLAGS}" -o worker "src/server/cmd/worker/main.go"
