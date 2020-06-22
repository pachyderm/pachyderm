# syntax=docker/dockerfile:1.0-experimental
ARG GO_VERSION
FROM golang:${GO_VERSION}
RUN apt update && apt install ca-certificates
RUN go get github.com/go-bindata/go-bindata/...
WORKDIR /app
COPY . .
ARG LD_FLAGS
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go-bindata -o src/server/cmd/worker/assets/assets.go -pkg assets /etc/ssl/certs/... && \
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o pachd "src/server/cmd/pachd/main.go" && \
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o worker "src/server/cmd/worker/main.go"
# symlink for systems (e.g. hub) that expect pachd to be in the old location
RUN ln -s /app/pachd /pachd