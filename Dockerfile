# TODO: pull in from etc/compile/GO_VERSION
FROM golang:1.13.8 AS builder

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build "src/server/cmd/pachd/main.go" && \
    mv main pachd && \
    CGO_ENABLED=0 GOOS=linux go build "src/server/cmd/worker/main.go" && \
    mv main worker

FROM debian:buster
WORKDIR /app
COPY --from=builder /app/pachd .
COPY --from=builder /app/worker .
