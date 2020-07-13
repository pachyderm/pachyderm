#!/bin/bash

set -e

if [ ! -z "$1" ]; then
    pachd_address="$1"
elif [ ! -z "$PACHD_PEER_SERVICE_HOST" ]; then
    pachd_address="$PACHD_PEER_SERVICE_HOST:$PACHD_PEER_SERVICE_PORT"
else
    pachd_address="localhost:30650"
fi

echo "{\"pachd_address\": \"$pachd_address\"}" | pachctl config set context examples
pachctl config set active-context examples
cat ~/.pachyderm/config.json
echo

./etc/testing/examples.sh
