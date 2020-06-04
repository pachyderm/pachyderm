ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine3.11 AS worker_build
WORKDIR /app
COPY init.go .
ARG LD_FLAGS
RUN go build -ldflags "${LD_FLAGS}" -o init "init.go"

FROM alpine:3.11
WORKDIR /app
COPY --from=pachyderm_build /app/worker .
COPY --from=worker_build /app/init .
