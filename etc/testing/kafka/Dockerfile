FROM golang:1.15.4
WORKDIR /app
ADD . /app/
RUN go build -o /app/main . 
CMD ["/app/main"]
