#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

eval $(minikube docker-env)
make install
make docker-build
