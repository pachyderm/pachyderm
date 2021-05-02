#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

make launch-loki

pachctl deploy local --no-guaranteed -d --dry-run | kubectl apply -f -
until timeout 1s ./etc/kube/check_ready.sh app=pachd; do sleep 1; done
pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
