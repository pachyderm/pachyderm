#!/bin/sh

CONTAINER="${1}"

RET=`docker wait "$CONTAINER"`

if [ "$RET" -ne 0 ]
then
	docker logs "$CONTAINER"
	exit "$RET"
fi
