#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

helm repo remove loki || true
helm repo add loki https://grafana.github.io/loki/charts
helm repo update
helm upgrade --install loki loki/loki-stack

pachctl deploy local --no-guaranteed -d --dry-run | kubectl apply -f -

# Block until all images are pulled - this allows us to pull loki and pachd images in parallel
until timeout 1s ./etc/kube/check_ready.sh app=pachd; do sleep 1; done
until timeout 1s ./etc/kube/check_ready.sh release=loki; do sleep 1; done

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"
