FROM golang:1.17.7-alpine3.15 as builder

RUN apk add --no-cache git openssh

WORKDIR /build

COPY source/main.go .
COPY source/go.mod .
COPY source/go.sum .

RUN go mod download
RUN go build -o ./app .

# Build final image
FROM alpine:3.12 as runner

COPY --from=builder /build/app /bin/app

CMD ["/bin/app"]
