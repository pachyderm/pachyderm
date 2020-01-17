#!/bin/bash

set -e

cd ${GOPATH}/src/github.com/pachyderm/pachyderm

function die {
  echo "error: $1" >2
  exit 1
}
export -f die

# get_images builds or download pachd and worker images
function get_images {
  if [[ "${PACH_VERSION}" = local ]]; then
    make install || die "could not build pachctl"
    make docker-build || die "could not build pachd/worker"
  else
    for i in pachd worker; do
      echo docker pull pachyderm/${i}:${PACH_VERSION}
      docker pull pachyderm/${i}:${PACH_VERSION}
    done
  fi
}
export -f get_images

## If the caller provided a tag, build and use that
export PACH_VERSION=local
KUBE_VERSION=v1.13.0
MINIKUBE_FLAGS=(--kubernetes-version="${KUBE_VERSION}")
eval "set -- $( getopt -l "tag:,cpus:,memory:" "--" "${0}" "${@:-}" )"
while true; do
  case "${1}" in
    --cpus)
      MINIKUBE_FLAGS+=(--cpus=${2})
      shift 2
      ;;
    --memory)
      MINIKUBE_FLAGS+=(--memory=${2})
      shift 2
      ;;
    --tag)
      export PACH_VERSION=${2##v}  # remove "v" prefix, e.g. v1.7.0
      shift 2
      ;;
    --)
      shift
      break
      ;;
  esac
done


# In parallel, start minikube, build pachctl, and get/build the pachd and worker images
cat <<EOF
Running:
  get_images
  minikube start ${MINIKUBE_FLAGS[@]}
EOF
cat <<EOF | tail -c+1 | xargs -I{} -n1 -P3 -- /bin/bash -c {}
get_images
minikube start ${MINIKUBE_FLAGS[@]}
EOF

# Print a spinning wheel while waiting for minikube to come up
set +x
WHEEL='-\|/'
until minikube ip 2>/dev/null; do
    # advance wheel 1/4 turn
    WHEEL=${WHEEL:1}${WHEEL::1}
    # jump to beginning of line & print message
    echo -en "\e[G\e[K${WHEEL::1} Waiting for minikube to start..."
    sleep 1
done
set -x

# Push pachyderm images to minikube VM
# (extract correct dash image from pachctl deploy)
dash_image="$(pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == "dash" and has("image")).image')"
grpc_proxy_image="$(pachctl deploy local -d --dry-run | jq -r '.. | select(.name? == "grpc-proxy").image')"
etcd_image="quay.io/coreos/etcd:v3.3.5"
docker pull ${etcd_image}
docker pull ${grpc_proxy_image}
docker pull ${dash_image}
etc/kube/push-to-minikube.sh pachyderm/pachd:${PACH_VERSION}
etc/kube/push-to-minikube.sh pachyderm/worker:${PACH_VERSION}
etc/kube/push-to-minikube.sh ${etcd_image}

# Deploy Pachyderm
if [[ "${PACH_VERSION}" = "local" ]]; then
  pachctl deploy local -d
else
  # deploy with -d (disable auth, small footprint), but use official version
  pachctl deploy local -d --dry-run | sed "s/:local/:${PACH_VERSION}/g" | kubectl create -f -
fi

active_kube_context=`kubectl config current-context`
pachctl config set context $active_kube_context -k $active_kube_context --overwrite
pachctl config update context $active_kube_context --pachd-address=$(minikube ip):30650
pachctl config set active-context $active_kube_context

# Wait for pachyderm to come up
set +x
until pachctl version; do
    # advance wheel 1/4 turn
    WHEEL=${WHEEL:1}${WHEEL:0:1}
    # jump to beginning of line & print message
    echo -en "\e[G\e[K${WHEEL:0:1} waiting for pachyderm to start..."
  sleep 1
done
set -x

# Kill pachctl port-forward and kubectl proxy
killall kubectl || true

# Port forward to etcd (for pfs/server/server_test.go)
export ETCD_POD=$(kubectl get pod -l suite=pachyderm,app=etcd -o jsonpath={.items[].metadata.name})
kubectl port-forward $ETCD_POD 32379:2379 &
