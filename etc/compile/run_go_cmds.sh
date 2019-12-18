#!/bin/sh
#
# This file actually builds Pachyderm binaries ('pachd' and 'worker') and is
# called by etc/compile/compile.sh inside the Pachyderm build container, but it
# run as an unprivileged user with the host caller's user ID, instead of as
# root. This prevents 'go build' from littering the pachyderm git directory
# with root-owned files

set -Eex

if [[ -n "${ROOT_PATH}" ]]; then
  # Called from linux via 'su' -- must reset PATH
  export PATH="${ROOT_PATH}"
fi

# Navigate to root of repo
cd "$(dirname "${0}")/../.."

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
