FROM golang:latest

WORKDIR /go/src/app
COPY src/map.go .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]
