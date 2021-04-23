#!/bin/bash

set -euo pipefail

# Mount the shared folder in the minikube VM
minikube stop
VBoxManage sharedfolder add minikube --name pach --hostpath /tmp/pach || true
minikube start
minikube ssh 'mkdir -p /tmp/pach && sudo mount -t vboxsf pach /tmp/pach' || true

# Disable the pachd running in minikube
kubectl scale --replicas=0 deployment/pachd
