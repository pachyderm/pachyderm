#!/bin/bash

set -ex

echo 'DOCKER_OPTS="-H unix:///var/run/docker.sock -s devicemapper"' | tee /etc/default/docker > /dev/null

# Install fuse
apt-get install -qq pkg-config fuse
modprobe fuse
chmod 666 /dev/fuse
cp etc/build/fuse.conf /etc/fuse.conf
chown root:$USER /etc/fuse.conf

# Install kubectl
# Latest as of 5/30/2018
# To get the latest kubectl version:
# curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt
KUBECTL_VERSION=v1.10.3
curl -L -o kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    sudo mv ./kubectl /usr/local/bin/kubectl
kubectl version --client

# Install minikube
# Latest as of 8/7/2018
# To get the latest minikube version:
# curl https://api.github.com/repos/kubernetes/minikube/releases | jq -r .[].tag_name | sort | tail -n1
MINIKUBE_VERSION=v0.28.2
curl -L -o minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64 && \
    chmod +x ./minikube && \
    sudo mv ./minikube /usr/local/bin/minikube

# Install vault
curl -Lo vault.zip https://releases.hashicorp.com/vault/0.9.5/vault_0.9.5_linux_amd64.zip && \
    unzip vault.zip && \
    sudo mv ./vault /usr/local/bin/vault
which vault

# Install etcdctl
# To get the latest etcd version:
# curl -s https://api.github.com/repos/coreos/etcd/releases | jq -r .[].tag_name | sort | tail -n1
ETCD_VERSION=v3.3.6
curl -L https://storage.googleapis.com/etcd/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz | \
  tar xzf - --strip-components=1 && \
  sudo mv ./etcdctl /usr/local/bin
etcdctl --version
