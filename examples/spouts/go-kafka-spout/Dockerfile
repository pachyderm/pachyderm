FROM golang:1.15.4
WORKDIR /app/source
ADD source .
RUN go build -o /app/main . 
CMD ["/app/main"]
