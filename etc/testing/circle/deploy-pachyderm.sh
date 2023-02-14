#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"
export PACHD_VERSION="$(jq -r .pachReleaseCommit version.json)"

helm repo add pachyderm https://pachyderm.github.io/helmchart
helm repo update
helm install \
    --wait --timeout 10m pachd pachyderm/pachyderm \
    --set deployTarget=LOCAL \
    --set console.enabled=false \
    --version=${PACHYDERM_VERSION} \
    --set pachd.image.tag=${PACHD_VERSION}

# There is a bug with wait for json path https://github.com/kubernetes/kubectl/issues/1236
# This bug has been fixed but the fix is only available in version 1.26
# However, using --wait in helm above seems to get us to the same solution.
# Therefore, the following code may not be needed.
kubectl wait --for=condition=available deployment -l app=pachd --timeout=5m
kubectl wait statefulset.apps/pachd-loki '--for=jsonpath={.status.readyReplicas}=1' --timeout=5m
pachctl version
