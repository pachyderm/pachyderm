#!/bin/bash

set -ex;

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh";

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

helm repo add pach https://helm.pachyderm.com

helm repo update

go test -v ./src/testing/deploy --timeout=3600s -v | stdbuf -i0 tee -a /tmp/results
