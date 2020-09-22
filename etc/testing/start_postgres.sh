#!/bin/sh

if [ -z "$(docker ps | grep postgres)" ]
then
    echo "starting postgres..."
    docker run -d \
    -e POSTGRES_USER=pachyderm \
    -e POSTGRES_DB=pgc \
    -e POSTGRES_HOST_AUTH_METHOD=trust \
    -p 32228:5432 \
    postgres:latest
else
    echo "postgres already started"
fi
