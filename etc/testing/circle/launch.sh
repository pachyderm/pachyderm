#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

make launch-loki

for i in $(seq 3); do
    make clean-launch-dev || true # may be nothing to delete
    make launch-dev && break
    (( i < 3 )) # false if this is the last loop (causes exit)
    sleep 10
done

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
