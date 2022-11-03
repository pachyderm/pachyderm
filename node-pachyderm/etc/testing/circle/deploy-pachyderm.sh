#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"

helm repo add pachyderm https://pachyderm.github.io/helmchart
helm repo update
helm install pachd pachyderm/pachyderm --set deployTarget=LOCAL --version ${PACHYDERM_VERSION}

# kubectl wait --for=condition=available deployment -l app=pachd --timeout=5m
# kubectl wait statefulset.apps/pachd-loki --for=jsonpath='{.status.readyReplicas}'=1 --timeout=5m
# There is a bug with wait for json path https://github.com/kubernetes/kubectl/issues/1236

sleep 120s
kubectl wait --for=condition=available deployment -l app=pachd --timeout=5m
kubectl wait statefulset.apps/pachd-loki --for=jsonpath='{.status.readyReplicas}'=1 --timeout=5m

pachctl version
