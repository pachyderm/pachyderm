#!/bin/bash

set -Eex

export PATH="${PWD}:${PWD}/cached-deps:${GOPATH}/bin:${PATH}"

# Parse flags
VERSION=v1.19.0
minikube_args=(
  "--vm-driver=docker"
  "--kubernetes-version=${VERSION}"
  "--cpus=7"
  "--memory=12Gi"
  "--wait=all"
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
