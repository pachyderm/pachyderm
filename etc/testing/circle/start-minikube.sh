#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

# Parse flags
KUBE_VERSION="$( jq -r .kubernetes <"$(git rev-parse --show-toplevel)/version.json" )"
minikube_args=(
  "--vm-driver=docker"
  "--kubernetes-version=${KUBE_VERSION}"
)
while getopts ":v" opt; do
  case "${opt}" in
    v)
      KUBE_VERSION="v${OPTARG}"
      ;;
    \?)
      echo "Invalid argument: ${opt}"
      exit 1
      ;;
  esac
done

minikube start "${minikube_args[@]}"

# Try to connect for three minutes
for _ in $(seq 36); do
  if kubectl version &>/dev/null; then
    exit 0
  fi
  sleep 5
done

# Give up--kubernetes isn't coming up
minikube delete
sleep 30 # Wait for minikube to go completely down
exit 1
