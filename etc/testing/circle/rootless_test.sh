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
    --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy \
    --addons=pod-security-policy

# add a PodSecurityPolicy which disables root
kubectl delete psp restricted privileged || true
kubectl apply -f etc/testing/pod-security-policy.yaml

./etc/testing/circle/build.sh

./etc/testing/circle/launch.sh

# Run TestSimplePipeline
go test -v ./src/server -run TestSimplePipeline
