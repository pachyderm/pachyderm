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
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/v1.8.2/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    sudo mv ./kubectl /usr/local/bin/kubectl

# Install minikube
curl -Lo minikube https://storage.googleapis.com/minikube/releases/v0.23.0/minikube-linux-amd64 && \
    chmod +x ./minikube && \
    sudo mv ./minikube /usr/local/bin/minikube

