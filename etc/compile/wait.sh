#!/bin/sh

CONTAINER="${1}"

docker wait "$CONTAINER" 2>/dev/null 1>/dev/null
RET=$?

if [ "$RET" -ne 0 ]
then
	docker logs "$CONTAINER"
	exit "$RET"
fi
