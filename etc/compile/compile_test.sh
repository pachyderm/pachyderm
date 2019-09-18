#!/bin/sh

set -Ee

# Builds a docker container that is just golang + docker client
if [ "${1}" = "--make-env" ]; then
	docker build ${DOCKER_BUILD_FLAGS} -t pachyderm_test_buildenv - <<EOF
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
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/pachyderm_test \
  ./src/server
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/pfs_server_test \
  ./src/server/pfs/server
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/pfs_cmds_test \
  ./src/server/pfs/cmds
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/pps_cmds_test \
  ./src/server/pps/cmds
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/auth_server_test \
  ./src/server/auth/server
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/auth_cmds_test \
  ./src/server/auth/cmds
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/enterprise_server_test \
  ./src/server/enterprise/server
go test \
  -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/worker_test \
  ./src/server/worker
go test -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/collection_test ./src/server/pkg/collection
go test -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/hashtree_test ./src/server/pkg/hashtree
go test -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/cert_test ./src/server/pkg/cert
go test -c -o /go/src/github.com/pachyderm/pachyderm/_tmp/vault_test ./src/plugin/vault
echo "Test built..."
pwd
ls /go/src/github.com/pachyderm/pachyderm/_tmp

cp Dockerfile.test _tmp/Dockerfile
cp etc/testing/artifacts/giphy.gif _tmp/
docker build -t pachyderm_test _tmp
docker tag pachyderm_test:latest pachyderm/test:latest
docker tag pachyderm_test:latest pachyderm/test:local
