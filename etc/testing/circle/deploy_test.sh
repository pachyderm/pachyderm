#!/bin/bash

set -exo pipefail

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

helm repo add pachyderm https://helm.pachyderm.com

helm repo update

go test -v -failfast -timeout 3600s ./src/testing/deploy -tags=k8s | stdbuf -i0 tee -a /tmp/results
