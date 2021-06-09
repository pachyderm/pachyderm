#!/bin/bash

set -ex

mkdir -p cached-deps

# Install deps
sudo snap install --classic goreleaser
sudo snap install --channel=edge shellcheck
sudo snap install jq
sudo apt update -y
sudo apt-get install -y -qq \
  silversearcher-ag \
  python3 \
  python3-pip \
  python3-setuptools \
  pkg-config \
  fuse \
  conntrack \
  pv

# Install fuse
sudo modprobe fuse
sudo chmod 666 /dev/fuse
sudo cp etc/build/fuse.conf /etc/fuse.conf
sudo chown root:root /etc/fuse.conf

# Install aws CLI (for TLS test)
pip3 install --upgrade --user wheel
pip3 install --upgrade --user awscli s3transfer==0.3.4

# Install kubectl
# To get the latest kubectl version:
# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
if [ ! -f cached-deps/kubectl ] ; then
    KUBECTL_VERSION=v1.19.2
    curl -L -o kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
        chmod +x ./kubectl
        mv ./kubectl cached-deps/kubectl
fi

# Install minikube
# To get the latest minikube version:
# curl https://api.github.com/repos/kubernetes/minikube/releases | jq -r .[].tag_name | sort -V | tail -n1
if [ ! -f cached-deps/minikube ] ; then
    MINIKUBE_VERSION=v1.13.1 # If changed, also do etc/kube/start-minikube.sh
    curl -L -o minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
        chmod +x ./minikube
        mv ./minikube cached-deps/minikube
fi

# Install etcdctl
# To get the latest etcd version:
# curl -Ls https://api.github.com/repos/etcd-io/etcd/releases | jq -r .[].tag_name
if [ ! -f cached-deps/etcdctl ] ; then
    ETCD_VERSION=v3.3.12
    curl -L https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz \
        | tar xzf - --strip-components=1
        mv ./etcdctl cached-deps/etcdctl
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
