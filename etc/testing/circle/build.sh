#!/bin/bash

set -ex

# shellcheck source=./env.sh
source "$(dirname "$0")/env.sh"

eval "$(minikube docker-env)"
make install
make docker-build
make docker-build-pipeline-build
