#!/bin/bash
set -xeuo pipefail
sudo apt update && sudo apt install -y fio

# cleanup
fusermount -u /pfs || true
sudo pkill -f "pachctl mount-server" || true
minikube delete || true

(cd ..
 minikube start
 eval $(minikube docker-env)
 make docker-build
 make install
 make launch-dev
)

sudo mkdir -p /pfs
sudo chown $USER /pfs