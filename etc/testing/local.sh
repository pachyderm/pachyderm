#!/bin/bash

set -euo pipefail

# make sure we're testing against minikube
kubectl config use-context minikube

export SHARED_DATA_DIR=/tmp/pach/
mkdir -p $SHARED_DATA_DIR/logs
rm -f $SHARED_DATA_DIR/logs/*

# connect to the k8s API in minikube
export KUBERNETES_PORT_443_TCP_ADDR=$(minikube ip)
export KUBERNETES_PORT=8443

# mount the bearer token for the default k8s service account
export KUBERNETES_BEARER_TOKEN_FILE=/tmp/pach/kubernetes-default-token
kubectl get -n kube-system -o json secret $(kubectl -n kube-system get secrets  | grep 'default-token' | awk '{print $1}') | jq -r '.data.token' | base64 -d > $KUBERNETES_BEARER_TOKEN_FILE

# point the serviceenv to the services
export ETCD_SERVICE_HOST="$(minikube ip)"
export ETCD_SERVICE_PORT=32379
export POSTGRES_SERVICE_HOST="$(minikube ip)"
export POSTGRES_SERVICE_PORT=32228
export POSTGRES_SERVICE_SSL=disable
export WORKER_IMAGE=pachyderm/worker:local
export WORKER_SIDECAR_IMAGE=pachyderm/pachd:local

export PACH_INMEMORY=true
export STORAGE_BACKEND=LOCAL
export PACHD_POD_NAME=local

$@
