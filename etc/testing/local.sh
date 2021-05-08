#!/bin/bash

set -euo pipefail

forward() {
  NAME=$1
  PORT=$2
  if [ -f /tmp/pach/port-forwards/$NAME.pid ]; then
    kill $(cat /tmp/pach/port-forwards/$NAME.pid) || true
  fi
  
  kubectl port-forward service/$NAME $PORT --address 127.0.0.1 >/dev/null 2>&1 &
  PID=$!
  
  mkdir -p /tmp/pach/port-forwards
  echo $PID > /tmp/pach/port-forwards/$NAME.pid
}

# make sure we're testing against minikube
kubectl config use-context minikube

# port-forward postgres and etcd
forward postgres 32228:5432
forward etcd 2379

echo '{"pachd_address": "grpc://localhost:650", "source": 2}' | pachctl config set context "local" --overwrite && pachctl config set active-context "local"

export ETCD_SERVICE_HOST="127.0.0.1"
export ETCD_SERVICE_PORT=2379
export POSTGRES_SERVICE_HOST="127.0.0.1"
export POSTGRES_SERVICE_PORT=32228
export POSTGRES_SERVICE_SSL=disable
export STORAGE_BACKEND=LOCAL
export WORKER_IMAGE=pachyderm/worker:local
export WORKER_SIDECAR_IMAGE=pachyderm/pachd:local
export STORAGE_UPLOAD_CONCURRENCY_LIMIT=100

export PACH_ROOT=/tmp/pach/storage
export STORAGE_HOST_PATH=/tmp/pach/storage
export PACH_CACHE_ROOT=/tmp/pach/cache

# connect to the k8s API in minikube
export KUBERNETES_PORT_443_TCP_ADDR=$(minikube ip)
export KUBERNETES_PORT=8443

# mount the bearer token for the default k8s service account
export KUBERNETES_BEARER_TOKEN_FILE=/tmp/pach/kubernetes-default-token
kubectl get -n kube-system -o json secret $(kubectl -n kube-system get secrets  | grep 'default-token' | awk '{print $1}') | jq -r '.data.token' | base64 -D > $KUBERNETES_BEARER_TOKEN_FILE

export LOCAL_TEST=true

$@
