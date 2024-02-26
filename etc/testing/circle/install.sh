#!/bin/bash

set -euxo pipefail

ARCH=amd64
if [ "$(uname -m)" = "aarch64" ]; then ARCH=arm64; fi

mkdir -p cached-deps

# Install deps
sudo apt update -y
sudo apt-get install -y -qq \
  silversearcher-ag \
  pkg-config \
  fuse \
  conntrack \
  pv \
  shellcheck \
  moreutils

# Install fuse
sudo modprobe fuse
sudo chmod 666 /dev/fuse
sudo cp etc/build/fuse.conf /etc/fuse.conf
sudo chown root:root /etc/fuse.conf

# Install kubectl
# To get the latest kubectl version:
# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
if [ ! -f cached-deps/kubectl ] ; then
    KUBECTL_VERSION=v1.27.4
    curl -L -o kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl && \
        chmod +x ./kubectl
        mv ./kubectl cached-deps/kubectl
fi

# Install minikube
# To get the latest minikube version:
# curl https://api.github.com/repos/kubernetes/minikube/releases | jq -r .[].tag_name | sort -V | tail -n1
if [ ! -f cached-deps/minikube ] ; then
    MINIKUBE_VERSION=v1.31.2
    curl -L -o minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-${ARCH} && \
        chmod +x ./minikube
        mv ./minikube cached-deps/minikube
fi

# Install kubeval
if [ ! -f cached-deps/kubeval ] && [ "$ARCH" = "amd64" ]; then
  KUBEVAL_VERSION=0.15.0
  curl -L https://github.com/instrumenta/kubeval/releases/download/${KUBEVAL_VERSION}/kubeval-linux-${ARCH}.tar.gz \
      | tar xzf - kubeval
      mv ./kubeval cached-deps/kubeval
fi

# Install helm
if [ ! -f cached-deps/helm ]; then
  HELM_VERSION=3.5.4
  curl -L https://get.helm.sh/helm-v${HELM_VERSION}-linux-${ARCH}.tar.gz \
      | tar xzf - linux-${ARCH}/helm
      mv ./linux-${ARCH}/helm cached-deps/helm
fi

if [ ! -f cached-deps/datadog-ci ]; then
  curl -L --fail "https://github.com/DataDog/datadog-ci/releases/latest/download/datadog-ci_linux-x64" \
    --output "./datadog-ci" && chmod +x "./datadog-ci"
  mv ./datadog-ci cached-deps/datadog-ci
fi
            