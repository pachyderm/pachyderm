#!/bin/bash

set -Eex

# Parse flags
VERSION=v1.19.0
minikube_args=(
  "--vm-driver=docker"
  "--kubernetes-version=${VERSION}"
  "--cpus=7"
  "--memory=12g"
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
