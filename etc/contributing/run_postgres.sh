#!/bin/sh
docker run --rm -it \
-e POSTGRES_DB=pachyderm \
-e POSTGRES_USER=pachyderm \
-e POSTGRES_HOST_AUTH_METHOD=trust \
-p 32228:5432 \
postgres:13.0-alpine
