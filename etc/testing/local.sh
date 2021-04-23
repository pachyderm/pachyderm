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

export SHARED_DATA_DIR=/tmp/pach/

# connect to the k8s API in minikube
export KUBERNETES_PORT_443_TCP_ADDR=$(minikube ip)
export KUBERNETES_PORT=8443

# mount the bearer token for the default k8s service account
export KUBERNETES_BEARER_TOKEN_FILE=/tmp/pach/kubernetes-default-token
kubectl get -n kube-system -o json secret $(kubectl -n kube-system get secrets  | grep 'default-token' | awk '{print $1}') | jq -r '.data.token' | base64 -D > $KUBERNETES_BEARER_TOKEN_FILE

export PACH_INMEMORY=true
export STORAGE_BACKEND=LOCAL
export STORAGE_UPLOAD_CONCURRENCY_LIMIT=100

$@
