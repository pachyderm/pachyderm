#!/bin/sh

set -Ee

# Builds a docker container that is just golang + docker client
if [ "${1}" = "--make-env" ]; then
	docker build -t pachyderm_test_buildenv - <<EOF
FROM golang:1.11.1
RUN curl -fsSL https://get.docker.com/builds/Linux/x86_64/docker-1.12.1.tgz \
  | tar -C /bin -xz docker/docker --strip-components=1 \
 && chmod +x /bin/docker
EOF
  exit 0
fi

# Runs inside docker container built above -- compiles pachyderm test
rm -rf ./_tmp/*
echo "Building test"
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/test \
  ./src/server
echo "Test built..."
pwd
ls /go/src/github.com/pachyderm/pachyderm/_tmp

cp Dockerfile.test _tmp/Dockerfile
cp etc/testing/artifacts/giphy.gif _tmp/
docker build -t pachyderm_test _tmp
docker tag pachyderm_test pachyderm/test:latest
docker tag pachyderm_test pachyderm/test:local
