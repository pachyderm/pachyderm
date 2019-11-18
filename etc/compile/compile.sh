#!/bin/sh

set -Eex

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"
pwd

BINARY="${1}"
LD_FLAGS="${2}"
PROFILE="${3}"

mkdir -p _tmp
# Note that github.com/pachyderm/pachyderm is mounted into the
# {pachd,worker}_compile docker container that this script is running in, so
# 'mkdir' below actually creates a dir on the host machine. The dir name
# includes ${BINARY} so that it doesn't collide with any concurrently-running
# pachyderm builds (e.g. when we build pachd and worker concurrently in 'make
# docker-build'). See https://github.com/pachyderm/pachyderm/issues/3845
TMP=docker_build_${BINARY}.tmpdir
mkdir -p "${TMP}"
CGO_ENABLED=0 GOOS=linux go build \
  -mod=vendor \
  -installsuffix netgo \
  -tags netgo \
  -o ${TMP}/${BINARY} \
  -ldflags "${LD_FLAGS}" \
  -gcflags "all=-trimpath=$GOPATH" \
  src/server/cmd/${BINARY}/main.go

echo "LD_FLAGS=$LD_FLAGS"

# When creating profile binaries, we dont want to detach or do docker ops
if [ -z ${PROFILE} ]
then
    cp Dockerfile.${BINARY} ${TMP}/Dockerfile
    if [ ${BINARY} = "worker" ]; then
        cp ./etc/worker/* ${TMP}/
    fi
    cp /etc/ssl/certs/ca-certificates.crt ${TMP}/ca-certificates.crt
    docker build ${DOCKER_BUILD_FLAGS} -t pachyderm_${BINARY} ${TMP}
    docker tag pachyderm_${BINARY}:latest pachyderm/${BINARY}:latest
    docker tag pachyderm_${BINARY}:latest pachyderm/${BINARY}:local
else
    cd ${TMP}
    tar cf - ${BINARY}
fi
rm -rf "${TMP}"
