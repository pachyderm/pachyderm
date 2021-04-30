#!/bin/bash

set -ex

GOPATH=/home/circleci/.go_workspace
export GOPATH

PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PATH

go get github.com/go-bindata/go-bindata/...

go-bindata -o src/server/cmd/worker/assets/assets.go -pkg assets /etc/ssl/certs/...

CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o init "./etc/worker/init.go"
CGO_ENABLED=0 go build -o pachctl_nocgo -ldflags "${LD_FLAGS}" -gcflags "${GC_FLAGS}" ./src/server/cmd/pachctl
CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o pachd "src/server/cmd/pachd/main.go"
CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o worker "src/server/cmd/worker/main.go"

go build -o pachctl -ldflags "${LD_FLAGS}" -gcflags "${GC_FLAGS}" ./src/server/cmd/pachctl

cp -R /etc/ssl/certs .
cp -R /usr/share/ca-certificates .

eval $(minikube docker-env)
docker build -t pachyderm/pachd:local . -f etc/testing/circle/Dockerfile.pachd
docker build -t pachyderm/pachctl:local . -f etc/testing/circle/Dockerfile.pachctl
docker build -t pachyderm/worker:local . -f etc/testing/circle/Dockerfile.worker
make install
make docker-build-pipeline-build
