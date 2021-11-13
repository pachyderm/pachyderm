FROM golang:1.17.3

WORKDIR /go/src/app
COPY src/map.go .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]
