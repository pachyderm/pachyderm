#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

eval $(minikube docker-env)
make install
make docker-build
make docker-build-pipeline-build
