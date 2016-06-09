#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

BINARY="${1}"
IMAGE="${2}"
if [ -z "${IMAGE}" ]; then
  IMAGE="${1}"
fi

mkdir -p _tmp
go build \
  -a \
  -installsuffix netgo \
  -tags netgo \
  -o _tmp/${BINARY} \
  src/server/cmd/${BINARY}/main.go
docker-compose build ${IMAGE}
docker tag -f pachyderm_${IMAGE}:latest derekchiang/${IMAGE}:latest
