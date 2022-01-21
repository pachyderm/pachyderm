#!/bin/bash

set -ex

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

go version

#make install
VERSION=$(pachctl version --client-only)
git config user.email "donotreply@pachyderm.com"
git config user.name "anonymous"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
#make docker-build

# oc login -u system:admin

# REGISTRY="$(oc get route -n openshift-image-registry | awk 'FNR == 2 {print $2}')" &&
# docker login -u unused -p "$(oc whoami -t)" "${REGISTRY}";

# docker tag pachyderm/pachd:local "${REGISTRY}"/pach-test/pachd:local &&
# docker tag pachyderm/worker:local "${REGISTRY}"/pach-test/worker:local &&

# docker push "${REGISTRY}"/pach-test/pachd:local &&
# docker push "${REGISTRY}"/pach-test/worker:local 


kubectl apply -f etc/testing/minio-openshift.yaml

# REGISTRY=$(oc get route -n openshift-image-registry | awk 'FNR == 2 {print $2}') &&
# PROJECT="pach-test"
# PACHD_REPO="${PROJECT}/pachd"
# WORKER_REPO="${PROJECT}/worker"

helm install pachyderm etc/helm/pachyderm -f etc/testing/circle/helm-values.yaml \
-f etc/testing/circle/helm-values-openshift.yaml \
--set pachd.image.tag="2.0.5" --set pachd.worker.image.tag="2.0.5";

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m