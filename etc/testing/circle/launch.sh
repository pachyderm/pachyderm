#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
