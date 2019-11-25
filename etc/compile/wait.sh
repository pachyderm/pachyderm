#!/bin/sh

CONTAINER="${1}"

RET=`docker wait "$CONTAINER" 2>/dev/null`

if [ "$RET" -ne 0 ]; then
	docker logs "$CONTAINER"
	exit "$RET"
fi
