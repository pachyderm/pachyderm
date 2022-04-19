#!/bin/bash

set -Eex

case $(uname -m) in

arm64 | aarch64)
  CPUS=4
  ;;

x86_64 | amd64)
  CPUS=7
  ;;
*)
  CPUS=4
  ;;
esac

export PATH="${PWD}:${PWD}/cached-deps:${GOPATH}/bin:${PATH}"

# Parse flags
VERSION=v1.19.0
minikube_args=(
  "--vm-driver=docker"
  "--kubernetes-version=${VERSION}"
  "--cpus=${CPUS}"
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

# wait for docker or timeout
timeout=120
while ! docker version >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for docker daemon"
    exit 1
  fi
  sleep 1
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
