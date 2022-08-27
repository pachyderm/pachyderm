# syntax=docker/dockerfile:1

# Build snowpach
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY src ./src
RUN ls -la ./src/*
RUN go build -o snowpach ./src/integrations/connectors/snowflake/snowpach/

# Deploy
FROM alpine:3.16.2

COPY --from=build /app/snowpach /usr/local/bin/

RUN addgroup -g 1000 snowflake && adduser -u 1000 -G snowflake -h /home/snowflake -D snowflake
USER snowflake:snowflake

WORKDIR /home/snowflake

ENTRYPOINT ["snowpach"]
