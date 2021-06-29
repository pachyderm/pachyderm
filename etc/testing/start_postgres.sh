#!/bin/sh

if [ "$CI" = "true" ]
then 
    echo "Skipping launching postgres, using PG launched in kube"
else 
    if ! docker ps | grep -q postgres
    then
        echo "starting postgres..."
        docker run -d \
        -e POSTGRES_DB=pgc \
        -e POSTGRES_HOST_AUTH_METHOD=trust \
        -p 32228:5432 \
        postgres:13.0-alpine
    else
        echo "postgres already started"
    fi
fi