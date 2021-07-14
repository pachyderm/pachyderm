#!/bin/bash

set -xeuo pipefail

# Run pachd locally like:
# CGO_ENABLED=0 go build ./src/server/cmd/pachd
# ./etc/testing/local.sh ./pachd
#
# If you use kind instead of minikube, you will have to override some configuration:
#
# LOCAL_PACH_CONTEXT=kind-kind LOCAL_PACH_APISERVER_HOST=127.0.0.1 LOCAL_PACH_APISERVER_PORT=46407 \
# ./etc/testing/local.sh ./pachd
#
# The truly motiviated engineer would teach pachd to read ~/.kube/config (testutil already does this
# if KUBERNETES_SERVICE_HOST is unset), or at least `jq` the host and port out of `kubectl config
# view -o json`.

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
kubectl config use-context ${LOCAL_PACH_CONTEXT:-'minikube'}

# TODO(jonathan): We should be able to make the helm chart generate this for us.
# add rbac rules
kubectl apply -f - <<EOF
---
apiVersion: v1
kind: ServiceAccount
metadata:
    labels:
        app: ""
        suite: pachyderm
    name: pachyderm
    namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
    labels:
        app: ""
        suite: pachyderm
    name: pachyderm
    namespace: default
rules:
    - apiGroups:
          - ""
      resources:
          - nodes
          - pods
          - pods/log
          - endpoints
      verbs:
          - get
          - list
          - watch
    - apiGroups:
          - ""
      resources:
          - replicationcontrollers
          - replicationcontrollers/scale
          - services
      verbs:
          - get
          - list
          - watch
          - create
          - update
          - delete
    - apiGroups:
          - ""
      resources:
          - secrets
      verbs:
          - get
          - list
          - watch
          - create
          - update
          - delete
          - deletecollection
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
    labels:
        app: ""
        suite: pachyderm
    name: pachyderm
    namespace: default
roleRef:
    apiGroup: ""
    kind: ClusterRole
    name: pachyderm
subjects:
    - kind: ServiceAccount
      name: pachyderm
      namespace: default
EOF

# port-forward postgres and etcd
forward postgres 32228:5432
forward etcd 2379

echo '{"pachd_address": "grpc://localhost:1650", "source": 2}' | pachctl config set context "local" --overwrite && pachctl config set active-context "local"

export LOCAL_TEST=true
export PACHD_POD_NAME=fake-pachd

export ETCD_SERVICE_HOST="127.0.0.1"
export ETCD_SERVICE_PORT=2379
export POSTGRES_HOST="127.0.0.1"
export POSTGRES_PORT=32228
export POSTGRES_SERVICE_SSL=disable
export STORAGE_BACKEND=LOCAL
export WORKER_IMAGE=pachyderm/worker:local
export WORKER_SIDECAR_IMAGE=pachyderm/pachd:local
export STORAGE_UPLOAD_CONCURRENCY_LIMIT=100

export PACH_ROOT=/tmp/pach/storage
export STORAGE_HOST_PATH=/tmp/pach/storage
export PACH_CACHE_ROOT=/tmp/pach/cache

# connect to the k8s API in minikube
export KUBERNETES_PORT_443_TCP_ADDR=${LOCAL_PACH_APISERVER_HOST:-$(minikube ip)}
export KUBERNETES_PORT=${LOCAL_PACH_APISERVER_PORT:-'8443'}

# mount the bearer token for the default k8s service account
export KUBERNETES_BEARER_TOKEN_FILE=/tmp/pach/kubernetes-default-token
kubectl get -n default -o json secret $(kubectl -n default get secrets  | grep 'pachyderm-token' | awk '{print $1}') | jq -r '.data.token | @base64d' > $KUBERNETES_BEARER_TOKEN_FILE

$@
