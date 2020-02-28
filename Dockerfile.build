# syntax=docker/dockerfile:1.0-experimental
# TODO(ys): ignore files that don't need to be synced
ARG GO_VERSION
FROM golang:${GO_VERSION}
WORKDIR /app
COPY . .
ARG LD_FLAGS
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o pachd "src/server/cmd/pachd/main.go" && \
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o worker "src/server/cmd/worker/main.go"
