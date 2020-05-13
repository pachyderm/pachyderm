#!/bin/bash

set -Eex

# Parse flags
VERSION=v1.17.0
minikube_args=(
  --vm-driver=none
  --kubernetes-version="${VERSION}"
)
while getopts ":v:r" opt; do
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


# Repeatedly restart minikube until it comes up. This corrects for an issue in
# Travis, where minikube will get stuck on startup and never recover
while true; do
  sudo env "PATH=$PATH" "CHANGE_MINIKUBE_NONE_USER=true" minikube start "${minikube_args[@]}"
  HEALTHY=false
  # Try to connect for one minute
  for i in $(seq 12); do
    if {
      kubectl version 2>/dev/null >/dev/null
    }; then
      HEALTHY=true
      break
    fi
    sleep 5
  done
  if [ "${HEALTHY}" = "true" ]; then break; fi

  # Give up--kubernetes isn't coming up
  minikube delete
  sleep 10 # Wait for minikube to go completely down
done
