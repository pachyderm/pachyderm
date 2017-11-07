#!/bin/sh

set -Ee

DIR="$(cd "$(dirname "${0}")/../.." && pwd)"
cd "${DIR}"

mkdir -p _tmp
go test \
  -c -o _tmp/test \
  ./src/server

cp Dockerfile.test _tmp/Dockerfile
cp etc/testing/artifacts/giphy.gif _tmp/
docker build -t pachyderm_test _tmp
docker tag pachyderm_test:latest pachyderm/test:latest
docker tag pachyderm_test:latest pachyderm/test:local
