FROM golang:1.14.2

RUN go get -u github.com/n0madic/twitter-scraper
ADD tweets.go /
RUN go build /tweets.go
