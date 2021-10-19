#!/bin/sh

set -ve

VERSION=v1.19.0
VMDRIVER="none"

# start minikube with pod security admission plugin
minikube start \
    --vmdriver=${VMDRIVER} \
    --kubernetes-version=${VERSION} \
    --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy \
    --addons=pod-security-policy

# add a PodSecurityPolicy which disables root
kubectl delete psp restricted privileged || true
kubectl apply -f etc/testing/pod-security-policy.yaml

make docker-build

./etc/testing/circle/launch-loki.sh
./etc/testing/circle/launch.sh

# Run TestSimplePipeline
go test -v ./src/server -run TestSimplePipeline
