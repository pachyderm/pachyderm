#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"
export PACHD_VERSION="$(jq -r .pachReleaseCommit version.json)"

helm repo add pachyderm https://pachyderm.github.io/helmchart
helm repo update
helm install \
    --wait --timeout 10m pachd pachyderm/pachyderm \
    --version=${PACHYDERM_VERSION} \
    --set deployTarget=LOCAL \
    --set console.image.tag=${CIRCLE_SHA1} \
    --set console.config.disableTelemetry=true \
    --set pachd.enterpriseLicenseKey=${PACHYDERM_ENTERPRISE_KEY} \
    --set pachd.image.tag=${PACHD_VERSION} \
    --set pachd.metrics.enabled=false \
    --set pachd.rootToken='pizza' \
    -f kind.yaml

pachctl connect grpc://127.0.0.1:80

echo "Waiting for pachd to be ready after deployment."
for i in `seq 1 30`; do
    sleep 2
    date
    if pachctl version; then
        echo "Pachd is serving."
        break
    fi
done
