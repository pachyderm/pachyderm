#!/bin/sh

set -ve

export PATH="${PWD}:${PWD}/cached-deps:${GOPATH}/bin:${PATH}"

VERSION=v1.19.0

# wait for docker or timeout
timeout=120
while ! docker version >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for docker daemon"
    exit 1
  fi
  sleep 1
done

# start minikube with pod security admission plugin
minikube start \
    --vm-driver=docker \
    --kubernetes-version=${VERSION} \
    --cpus=4 \
    --memory=12Gi \
    --wait=all \

# install gatekeeper
kubectl apply -f etc/testing/gatekeeper.yaml

# install gatekeeper OPA Templates
kubectl apply -f etc/testing/opa-policies/
sleep 5
#Install gatekeeper OPA constraints 
kubectl apply -f etc/testing/opa-constraints.yaml

./etc/testing/circle/build.sh

./etc/testing/circle/launch.sh

# Run TestSimplePipeline
go test -v ./src/server -run TestSimplePipeline
