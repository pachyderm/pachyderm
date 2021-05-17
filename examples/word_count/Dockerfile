FROM golang:1.16.4

WORKDIR /go/src/app
COPY src/map.go .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]
