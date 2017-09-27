#!/bin/bash

set -ex

wget https://storage.googleapis.com/kubernetes-release/release/v1.6.2/bin/linux/amd64/kubectl
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
make launch-kube
kubectl get all
