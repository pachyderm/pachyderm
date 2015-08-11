#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

export CGO_LDFLAGS="/usr/local/lib/libgit2.a $(pkg-config --libs --static /usr/local/lib/pkgconfig/libgit2.pc) \
/usr/lib/x86_64-linux-gnu/libssl.a $(pkg-config --libs --static /usr/lib/x86_64-linux-gnu/pkgconfig/libssl.pc)"

mkdir -p _tmp
go build \
  -a \
  -installsuffix netgo \
  -tags netgo \
  -o _tmp/${1} \
  src/cmd/${1}/main.go
cat docker-compose.yml
docker-compose build ${1}
docker tag -f pachyderm_${1}:latest pachyderm/${1}:latest
