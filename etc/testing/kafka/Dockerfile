FROM golang:1.16.4
WORKDIR /app
ADD . /app/
RUN go build -o /app/main . 
CMD ["/app/main"]
