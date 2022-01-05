#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

go version

make install
VERSION=$(pachctl version --client-only)
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
make docker-build

# REGISTRY="default-route-openshift-image-registry.apps-crc.testing"
REGISTRY=$(oc get route -n openshift-image-registry | awk 'FNR == 2 {print $2}') &&
docker login -u unused -p $(oc whoami -t) ${REGISTRY} &&

docker tag pachyderm/pachd:local ${REGISTRY}/pach-test/pachd:local &&
docker tag pachyderm/worker:local ${REGISTRY}/pach-test/worker:local &&

docker push ${REGISTRY}/pach-test/pachd:local &&
docker push ${REGISTRY}/pach-test/worker:local 
