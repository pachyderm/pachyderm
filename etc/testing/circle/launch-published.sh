#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

# deploy object storage
kubectl apply -f etc/testing/minio.yaml

helm repo add pach https://helm.pachyderm.com
helm repo update

VERSION="2.0.4"
helm install pachyderm -f etc/testing/circle/helm-values.yaml pach/pachyderm --version "$VERSION" --set pachd.image.tag="$VERSION";

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

# Wait for loki to be deployed
kubectl wait --for=condition=ready pod -l release=loki --timeout=5m

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"