#!/usr/bin/env bash
set -xeu -o pipefail -o errexit

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5001'
if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:5000"]
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
          - containerPort: 30443
            hostPort: 443
            protocol: TCP
EOF

# connect the registry to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
  docker network connect "kind" "${reg_name}"
fi

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
    name: pachd-config
data:
    MODE: full
EOF

#helm install pachyderm etc/helm/pachyderm -f etc/kind/hostname-values.yaml -f etc/kind/enterprise-key-values.yaml -f etc/kind/values.yaml

helm install pachyderm ./etc/helm/pachyderm -f etc/kind/hostname-values.yaml -f etc/kind/enterprise-key-values.yaml -f etc/kind/values.yaml

kubectl rollout status deployment pachd --timeout=10s

pachctl config set context kind --overwrite <<EOF
{"pachd_address": "grpc://localhost:80", "session_token": "V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6"}
EOF

pachctl config set active-context kind

pachctl version

mc alias set kind http://localhost V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6 V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6

mc ls kind
