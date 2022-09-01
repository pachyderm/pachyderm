#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH
export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"
export PACHD_VERSION="$(jq -r .pachReleaseCommit version.json)"

helm repo add pachyderm https://pachyderm.github.io/helmchart
helm repo update
helm install pachd pachyderm/pachyderm --set deployTarget=LOCAL --set console.enabled=false --version=${PACHYDERM_VERSION} --set pachd.image.tag=${PACHD_VERSION}

kubectl wait --for=condition=available deployment -l app=pachd --timeout=5m
pachctl version

