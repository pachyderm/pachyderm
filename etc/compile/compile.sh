#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

BINARY="${1}"
LD_FLAGS="${2}"
PROFILE="${3}"

mkdir -p _tmp
go build \
  -a \
  -installsuffix netgo \
  -tags netgo \
  -o _tmp/${BINARY} \
  -ldflags "${LD_FLAGS}" \
  src/server/cmd/${BINARY}/main.go

# When creating profile binaries, we dont want to detach or do docker ops
if [ -z ${PROFILE} ]
then
    docker-compose build ${BINARY}
    docker tag -f pachyderm_${BINARY}:latest pachyderm/${BINARY}:latest
fi

tar cf - _tmp/${BINARY}

