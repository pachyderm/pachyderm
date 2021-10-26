#!/bin/bash

set -ex

mkdir -p cached-deps

# Install deps
sudo apt update -y
sudo apt-get install -y -qq \
  pkg-config \
  conntrack

# Install Helm
if [ ! -f cached-deps/helm ]; then
  HELM_VERSION=3.5.4
  curl -L https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz \
      | tar xzf - linux-amd64/helm
      mv ./linux-amd64/helm cached-deps/helm
fi

# Install kubectl
# To get the latest kubectl version:
# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
if [ ! -f cached-deps/kubectl ] ; then
    KUBECTL_VERSION="v$(jq -r .kubernetes version.json)"
    curl -L -o kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
        chmod +x ./kubectl
        mv ./kubectl cached-deps/kubectl
fi

# Install minikube
# To get the latest minikube version:
# curl https://api.github.com/repos/kubernetes/minikube/releases | jq -r .[].tag_name | sort -V | tail -n1
if [ ! -f cached-deps/minikube ] ; then
    MINIKUBE_VERSION="v$(jq -r .minikube version.json)"
    curl -L -o minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
        chmod +x ./minikube
        mv ./minikube cached-deps/minikube
fi

export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"

# Install Pachyderm
curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v${PACHYDERM_VERSION}/pachctl_${PACHYDERM_VERSION}_amd64.deb
sudo dpkg -i /tmp/pachctl.deb
