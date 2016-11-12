#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

BINARY="${1}"
LD_FLAGS="${2}"
PROFILE="${3}"

mkdir -p _tmp
CGO_ENABLED=0 GOOS=linux go build \
  -a \
  -installsuffix netgo \
  -tags netgo \
  -o _tmp/${BINARY} \
  -ldflags "${LD_FLAGS}" \
  src/server/cmd/${BINARY}/main.go

echo "LD_FLAGS=$LD_FLAGS"

# When creating profile binaries, we dont want to detach or do docker ops
if [ -z ${PROFILE} ]
then
    cp Dockerfile.${BINARY} _tmp/Dockerfile
    if [ ${BINARY} = "job-shim" ]; then
        cp ./etc/job-shim/* _tmp/
    fi
    cp /etc/ssl/certs/ca-certificates.crt _tmp/ca-certificates.crt
    docker build -t pachyderm_${BINARY} _tmp
    docker tag pachyderm_${BINARY}:latest pachyderm/${BINARY}:latest
    docker tag pachyderm_${BINARY}:latest pachyderm/${BINARY}:local
else
    cd _tmp
    tar cf - ${BINARY}
fi


