ARG GOVERSION=golang:1.19.0
FROM $GOVERSION
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]
