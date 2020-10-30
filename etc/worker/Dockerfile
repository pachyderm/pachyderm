ARG GO_VERSION
FROM golang:${GO_VERSION} AS worker_build
WORKDIR /app
COPY etc/worker/init.go .
ARG LD_FLAGS
RUN CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o init "init.go"

FROM scratch
WORKDIR /app
COPY --from=pachyderm_build /app/worker .
COPY --from=worker_build /app/init .
