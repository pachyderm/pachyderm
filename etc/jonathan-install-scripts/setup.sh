#!/usr/bin/env bash
set -xeu -o pipefail -o errexit

if [ '!' -d etc/helm/pachyderm ]; then
    echo "Run from a pachyderm checkout"
fi

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5001'
if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --name=kind --config=-
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
          - containerPort: 30650
            hostPort: 30650
            protocol: TCP
          - containerPort: 30655
            hostPort: 30655
            protocol: TCP

          - containerPort: 30660
            hostPort: 30660
            protocol: TCP
          - containerPort: 30661
            hostPort: 30661
            protocol: TCP
          - containerPort: 30662
            hostPort: 30662
            protocol: TCP
          - containerPort: 30663
            hostPort: 30663
            protocol: TCP
          - containerPort: 30664
            hostPort: 30664
            protocol: TCP
          - containerPort: 30665
            hostPort: 30665
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

# minio
kubectl apply -f $SCRIPT_DIR/minio.yaml

# dns speed improvements
kubectl apply -f $SCRIPT_DIR/coredns_configmap.yaml

# metrics-server ("kubectl top pods")
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/

helm repo update

helm upgrade --install --set args={--kubelet-insecure-tls} metrics-server metrics-server/metrics-server --namespace kube-system

# TLS key for proxy.
# kubectl create secret tls envoy-tls --cert=/tls/tls.crt --key=/tls/tls.key

# Try to pre-load some images
for img in $(helm template pachyderm etc/helm/pachyderm -f $SCRIPT_DIR/hostname-values.yaml -f $SCRIPT_DIR/enterprise-key-values.yaml -f $SCRIPT_DIR/values.yaml | yq '.. | select(type|test("map")) | .image | select(. != null)' | grep -v -- --- | grep -v pachd); do
    (docker pull $img && kind load docker-image $img) || echo "Could not pre-load $img"
done

# Install pachyderm
helm install pachyderm etc/helm/pachyderm -f $SCRIPT_DIR/hostname-values.yaml -f $SCRIPT_DIR/enterprise-key-values.yaml -f $SCRIPT_DIR/values.yaml

kubectl rollout status deployment pachd

pachctl config set context kind --overwrite <<EOF
{"pachd_address": "grpc://localhost:80", "session_token": "V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6"}
EOF

pachctl config set active-context kind

for i in `seq 1 10`; do
    if pachctl version; then
        echo "pachd ok"
        break
    else
        sleep 5
        continue
    fi
done

make -C examples/opencv opencv

mc alias set kind http://localhost V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6 V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6

mc ls kind
