#!/bin/bash

set -ex

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
    KUBECTL_VERSION=v1.23.5
    curl -L -o kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
        chmod +x ./kubectl
        mv ./kubectl cached-deps/kubectl
fi

# Install minikube
# To get the latest minikube version:
# curl https://api.github.com/repos/kubernetes/minikube/releases | jq -r .[].tag_name | sort -V | tail -n1
if [ ! -f cached-deps/minikube ] ; then
    MINIKUBE_VERSION=v1.25.2 # If changed, also do etc/kube/start-minikube.sh
    curl -L -o minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
        chmod +x ./minikube
        mv ./minikube cached-deps/minikube
fi

# Install kubeval
if [ ! -f cached-deps/kubeval ]; then
  KUBEVAL_VERSION=0.15.0
  curl -L https://github.com/instrumenta/kubeval/releases/download/${KUBEVAL_VERSION}/kubeval-linux-amd64.tar.gz \
      | tar xzf - kubeval
      mv ./kubeval cached-deps/kubeval
fi

# Install helm
if [ ! -f cached-deps/helm ]; then
  HELM_VERSION=3.5.4
  curl -L https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
      | tar xzf - linux-amd64/helm
      mv ./linux-amd64/helm cached-deps/helm
fi

# Install goreleaser
if [ ! -f cached-deps/goreleaser ]; then
  GORELEASER_VERSION=0.169.0
  curl -L https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz \
      | tar xzf - -C cached-deps goreleaser
fi
