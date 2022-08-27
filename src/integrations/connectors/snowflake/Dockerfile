# syntax=docker/dockerfile:1

FROM golang:1.19 AS build

WORKDIR /app
COPY go.mod go.sum ./
COPY src ./src

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/snowpach ./src/integrations/connectors/snowflake/snowpach

# We need debug for shell, since we have an issue with directly writing to /pfs/out
FROM gcr.io/distroless/static-debian11:debug

COPY --from=build /go/bin/snowpach /usr/local/bin/

CMD ["snowpach"]
