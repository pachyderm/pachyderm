# syntax=docker/dockerfile:1

# Build
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY src ./src
RUN ls -la ./src/*
RUN go build -o snowpach ./src/integrations/connectors/snowflake/cmd/

# Deploy
FROM alpine:3.16.2

WORKDIR /usr/local/bin

COPY --from=build /app/snowpach /usr/local/bin/snowpach
