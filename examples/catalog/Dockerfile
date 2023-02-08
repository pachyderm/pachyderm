# syntax=docker/dockerfile:1

FROM golang:1.20-alpine

RUN go install github.com/blevesearch/bleve/v2/cmd/bleve@latest

RUN apk update
RUN apk add tree
RUN apk add jq

ADD /scripts /scripts
