#!/bin/bash

set -ex

export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

cat <<EOF | kind create cluster --name=kind --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
    - role: control-plane
      kubeadmConfigPatches:
          - |
              kind: InitConfiguration
              nodeRegistration:
                  kubeletExtraArgs:
                      node-labels: "ingress-ready=true"
      extraPortMappings:
          - containerPort: 30080
            hostPort: 80
            protocol: TCP
EOF

kubectl config set current-context kind-kind