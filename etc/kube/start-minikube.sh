#!/bin/sh

set -Ee

sudo CHANGE_MINIKUBE_NONE_USER=true minikube start --vm-driver=none --kubernetes-version=v1.7.5
until kubectl version 2>/dev/null >/dev/null; do sleep 5; done
