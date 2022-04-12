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
# Install kubectl
# To get the latest kubectl version:
# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
if [ ! -f cached-deps/kubectl ] ; then
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
      chmod +x ./kubectl
      mv ./kubectl cached-deps/kubectl
fi


if [ ! -f cached-deps/kind ] ; then
    KIND_VERSION="v$(jq -r .kind version.json)"
    curl -fLo ./kind-linux-amd64 "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64" \
        && chmod +x ./kind-linux-amd64 \
        && sudo mv ./kind-linux-amd64 /usr/local/bin/kind
fi

export PACHYDERM_VERSION="$(jq -r .pachyderm version.json)"

# Install Pachyderm
curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v${PACHYDERM_VERSION}/pachctl_${PACHYDERM_VERSION}_amd64.deb
sudo dpkg -i /tmp/pachctl.deb
