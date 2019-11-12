#!/bin/sh

CONTAINER="${1}"

docker wait "$CONTAINER" 2>/dev/null 1>/dev/null
RET=$?
docker logs "$CONTAINER"
exit "$RET"
