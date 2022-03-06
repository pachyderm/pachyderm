#!/bin/bash
set -xeuo pipefail
minikube delete
(cd ..
 minikube start
 eval $(minikube docker-env)
 make docker-build
 make install
 make launch-dev
)
sudo mkdir -p /pfs
sudo chown $USER /pfs
pachctl mount-server