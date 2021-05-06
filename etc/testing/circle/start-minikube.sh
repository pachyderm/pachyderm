#!/bin/bash

set -Eex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

# Parse flags
VERSION=v1.13.0
minikube_args=(
  "--vm-driver=docker"
  "--kubernetes-version=${VERSION}"
)
while getopts ":v" opt; do
  case "${opt}" in
    v)
      VERSION="v${OPTARG}"
      ;;
    \?)
      echo "Invalid argument: ${opt}"
      exit 1
      ;;
  esac
done

if [[ -n "${TRAVIS}" ]]; then
  minikube_args+=("--bootstrapper=kubeadm")
fi

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
