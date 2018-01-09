#!/bin/sh

set -Eex

# Parse flags
VERSION=v1.8.0
RBAC=""
while getopts ":v:r" opt; do
  case "${opt}" in
    v)
      VERSION="v${OPTARG}"
      ;;
    r)
      RBAC="--extra-config=apiserver.Authorization.Mode=RBAC"
      ;;
    \?)
      echo "Invalid argument: ${opt}"
      exit 1
      ;;
  esac
done

sudo CHANGE_MINIKUBE_NONE_USER=true minikube start --vm-driver=none --kubernetes-version="${VERSION}" "${RBAC}"
until kubectl version 2>/dev/null >/dev/null; do sleep 5; done

# Apply some manual changes to fix DNS.
kubectl -n kube-system create sa kube-dns
until kubectl -n kube-system patch deploy/kube-dns -p '{"spec": {"template": {"spec": {"serviceAccountName": "kube-dns"}}}}' 2>/dev/null >/dev/null; do sleep 5; done
