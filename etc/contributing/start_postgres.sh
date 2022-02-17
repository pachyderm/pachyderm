#!/bin/sh

set -ve

if ! docker ps | grep -q postgres
then
    echo "starting postgres..."
    postgres_id=$(docker run -d \
    -e POSTGRES_DB=pachyderm \
    -e POSTGRES_USER=pachyderm \
    -e POSTGRES_HOST_AUTH_METHOD=trust \
    -p 30228:5432 \
    postgres:13.0-alpine)

    postgres_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' "${postgres_id}")

    docker run -d \
    -e AUTH_TYPE=any \
    -e POSTGRESQL_USERNAME="pachyderm" \
    -e POSTGRESQL_PASSWORD="password" \
    -e POSTGRESQL_HOST="${postgres_ip}" \
    -e PGBOUNCER_PORT=5432 \
    -e PGBOUNCER_POOL_MODE=transaction \
    -p 30229:5432 \
    pachyderm/pgbouncer:1.16.1-1b
else
    echo "postgres already started"
fi

