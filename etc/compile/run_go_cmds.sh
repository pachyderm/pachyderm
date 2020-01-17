#!/bin/bash
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
which go

# Navigate to root of repo
cd "$(dirname "${0}")/../.."

BINARY="${1}"
LD_FLAGS="${2}"
PROFILE="${3}"

# Ubuntu mounts an in-memory filesystem at /dev/shm (rather than /tmp) so use
# that parent for performance
TMP="$(mktemp -t docker_build_${BINARY}.XXXXXXXX)"
rm -rf "${TMP}" # in case a directory was left behind by a prior failed build
mkdir -p "${TMP}"
CGO_ENABLED=0 GOOS=linux go build \
  -installsuffix netgo \
  -tags netgo \
  -o "${TMP}/${BINARY}" \
  -ldflags "${LD_FLAGS}" \
  -gcflags "all=-trimpath=$GOPATH" \
  "src/server/cmd/${BINARY}/main.go"

# When creating profile binaries, we dont want to detach or do docker ops
if [[ -z "${PROFILE}" ]]; then
    cp "Dockerfile.${BINARY}" "${TMP}/Dockerfile"
    if [[ "${BINARY}" = "worker" ]]; then
        cp ./etc/worker/* "${TMP}/"
    fi
    cp /etc/ssl/certs/ca-certificates.crt "${TMP}/ca-certificates.crt"
    docker build ${DOCKER_BUILD_FLAGS} -t "pachyderm_${BINARY}" "${TMP}"
    docker tag "pachyderm_${BINARY}:latest" "pachyderm/${BINARY}:latest"
    docker tag "pachyderm_${BINARY}:latest" "pachyderm/${BINARY}:local"
else
    cd "${TMP}"
    tar cf - "${BINARY}"
fi
rm -rf "${TMP}"
