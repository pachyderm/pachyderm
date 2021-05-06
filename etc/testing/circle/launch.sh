#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

make launch-loki

for i in $(seq 3); do
    make clean-launch-dev || true # may be nothing to delete
    make launch-dev && break
    (( i < 3 )) # false if this is the last loop (causes exit)
    sleep 10
done

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
