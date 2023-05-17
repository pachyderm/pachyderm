ARG GOVERSION=golang:1.20.4
FROM $GOVERSION
WORKDIR /app
ADD . /app/
RUN go build -o /app/main .
CMD ["/app/main"]
