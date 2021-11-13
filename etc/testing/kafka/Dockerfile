FROM golang:1.17.3
WORKDIR /app
ADD . /app/
RUN go build -o /app/main . 
CMD ["/app/main"]
