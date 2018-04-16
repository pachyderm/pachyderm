#!/bin/bash

set -x

cd ${GOPATH}/src/github.com/pachyderm/pachyderm

# push_to_minikube pushes dockers images to the minikube vm
function push_to_minikube {
  if [[ $# -ne 1 ]]; then
    echo "error: need the name of the docker image to push"
  fi
  docker save "${1}" | pv | (
    eval $(minikube docker-env)
    docker load
  )
}

# get_images builds or download pachd and worker images
function get_images {
  if [[ "${PACH_VERSION}" = local ]]; then
    make docker-build
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
eval "set -- $( getopt -l "tag:" "--" "${0}" "${@:-}" )"
while true; do
  case "${1}" in
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
cat <<EOF | tail -c+1 | xargs -d\\n -n1 -P3 -- /bin/bash -c
make install
get_images
minikube start
EOF

# Push pachyderm images (and etcd) to minikube VM
push_to_minikube pachyderm/pachd:${PACH_VERSION}
push_to_minikube pachyderm/worker:${PACH_VERSION}
push_to_minikube pachyderm/etcd:v3.2.7

# Deploy Pachyderm
export ADDRESS=$(minikube ip):30650
echo "ADDRESS is ${ADDRESS}"
if [[ "${PACH_VERSION}" = "local" ]]; then
  pachctl deploy local -d
else
  # deploy with -d (disable auth, small footprint), but use official version
  pachctl deploy local -d --dry-run | sed "s/:local/:${PACH_VERSION}/g" | kubectl create -f -
fi

# Wait for pachyderm to come up
until pachctl version; do
  export ADDRESS=$(minikube ip):30650
  echo "ADDRESS is ${ADDRESS}"
  sleep 1
done

# Kill pachctl port-forward and kubectl proxy
killall kubectl || true
