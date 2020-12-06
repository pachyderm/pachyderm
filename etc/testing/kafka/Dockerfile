FROM golang:1.15.4
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
RUN go get github.com/segmentio/kafka-go@0.24
RUN go build -o main . 
CMD ["/app/main"]
