#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

helm install pachyderm etc/helm/pachyderm -f etc/testing/circle/helm-values.yaml

kubectl logs -l app=pachd

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

kubectl logs -l app=pachd

# Wait for loki to be deployed
kubectl wait --for=condition=ready pod -l app=loki --timeout=5m
kubectl wait --for=condition=ready pod -l app=promtail --timeout=5m

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
