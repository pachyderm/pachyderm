#!/bin/bash

set -euo pipefail

# This is only required once to configure the shared volume
# minikube stop
# VBoxManage sharedfolder add minikube --name pach --hostpath /tmp/pach || true
# minikube start

# Mount the shared folder in the minikube VM, required on every restart
minikube ssh 'mkdir -p /tmp/pach && sudo mount -t vboxsf pach /tmp/pach'

# Build pachctl
make

# Build the worker docker images and tag them with `local`
make docker-build
VERSION=local make docker-tag

# Deploy all the pods except pachd
pachctl deploy local -d --no-dashboard --dry-run | jq 'select(.metadata.name != "pachd")' | kubectl create -f -
