#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

export CGO_LDFLAGS="/usr/local/lib/libgit2.a $(pkg-config --libs --static /usr/local/lib/pkgconfig/libgit2.pc) \
/usr/lib/x86_64-linux-gnu/libssl.a $(pkg-config --libs --static /usr/lib/x86_64-linux-gnu/pkgconfig/libssl.pc)"

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
  src/cmd/${BINARY}/main.go
docker-compose build ${IMAGE}
docker tag -f pachyderm_${IMAGE}:latest pachyderm/${IMAGE}:latest
