#!/bin/bash

set -Eex

# Parse flags
minikube_args=(
  "--vm-driver=none"
  "--kubernetes-version=v1.13.0"
)

if [[ -n "${TRAVIS}" ]]; then
  minikube_args+=("--bootstrapper=kubeadm")
fi


# Repeatedly restart minikube until it comes up. This corrects for an issue in
# Travis, where minikube will get stuck on startup and never recover
while true; do
  sudo env "PATH=$PATH" "CHANGE_MINIKUBE_NONE_USER=true" minikube start "${minikube_args[@]}"
  HEALTHY=false
  # Try to connect for one minute
  for _ in $(seq 12); do
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
