#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"

helm repo add pachyderm https://pachyderm.github.io/helmchart
helm repo update
helm install pachd pachyderm/pachyderm --set deployTarget=LOCAL --version ${PACHYDERM_VERSION}

kubectl wait --for=condition=available deployment -l app=pachd --timeout=5m
pachctl version
echo $PACHYDERM_ENTERPRISE_KEY | pachctl license activate
touch .env.development.local
CI=true make setup-auth

