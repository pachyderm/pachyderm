FROM golang:1.15.2-alpine3.12 as builder

RUN apk add --no-cache git openssh

WORKDIR /build

COPY source/main.go .
COPY source/go.mod .

RUN go mod download && go build -o /build/app

# Build final image
FROM alpine:3.12 as runner

COPY --from=builder /build/app /bin/app

CMD ["/bin/app"]