#!/bin/bash

set -Eex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

# Parse flags
minikube_args=(
  "--image=kindest/node:v1.19.11@sha256:07db187ae84b4b7de440a73886f008cf903fcf5764ba8106a9fd5243d6f32729"
  "--wait=3m"
)

kind create cluster "${minikube_args[@]}"

for i in $(ls images); do
  kind load images/$i
done
