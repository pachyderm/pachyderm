#!/bin/sh
docker run --rm -it \
-e POSTGRES_DB=pgc \
-e POSTGRES_HOST_AUTH_METHOD=trust \
-p 32228:5432 \
postgres:13.0-alpine
