FROM golang:1.16.4
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
RUN go build -o main . 
CMD ["/app/main"]
