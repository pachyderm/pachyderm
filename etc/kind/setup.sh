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
          - containerPort: 30400
            hostPort: 30400
            protocol: TCP
          - containerPort: 30600
            hostPort: 30600
            protocol: TCP
          - containerPort: 30650
            hostPort: 30650
            protocol: TCP
          - containerPort: 30651
            hostPort: 30651
            protocol: TCP
          - containerPort: 30652
            hostPort: 30652
            protocol: TCP
          - containerPort: 30653
            hostPort: 30653
            protocol: TCP
          - containerPort: 30654
            hostPort: 30654
            protocol: TCP
          - containerPort: 30655
            hostPort: 30655
            protocol: TCP
          - containerPort: 30656
            hostPort: 30656
            protocol: TCP
          - containerPort: 30657
            hostPort: 30657
            protocol: TCP
          - containerPort: 30658
            hostPort: 30658
            protocol: TCP
          - containerPort: 30659
            hostPort: 30659
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
          - containerPort: 30666
            hostPort: 30666
            protocol: TCP
          - containerPort: 30667
            hostPort: 30667
            protocol: TCP
          - containerPort: 30668
            hostPort: 30668
            protocol: TCP
          - containerPort: 30669
            hostPort: 30669
            protocol: TCP
          - containerPort: 30670
            hostPort: 30670
            protocol: TCP
          - containerPort: 30671
            hostPort: 30671
            protocol: TCP
          - containerPort: 30672
            hostPort: 30672
            protocol: TCP
          - containerPort: 30673
            hostPort: 30673
            protocol: TCP
          - containerPort: 30674
            hostPort: 30674
            protocol: TCP
          - containerPort: 30675
            hostPort: 30675
            protocol: TCP
          - containerPort: 30676
            hostPort: 30676
            protocol: TCP
          - containerPort: 30677
            hostPort: 30677
            protocol: TCP
          - containerPort: 30678
            hostPort: 30678
            protocol: TCP
          - containerPort: 30679
            hostPort: 30679
            protocol: TCP
          - containerPort: 30680
            hostPort: 30680
            protocol: TCP
          - containerPort: 30681
            hostPort: 30681
            protocol: TCP
          - containerPort: 30682
            hostPort: 30682
            protocol: TCP
          - containerPort: 30683
            hostPort: 30683
            protocol: TCP
          - containerPort: 30684
            hostPort: 30684
            protocol: TCP
          - containerPort: 30685
            hostPort: 30685
            protocol: TCP
          - containerPort: 30686
            hostPort: 30686
            protocol: TCP
          - containerPort: 30687
            hostPort: 30687
            protocol: TCP
          - containerPort: 30688
            hostPort: 30688
            protocol: TCP
          - containerPort: 30689
            hostPort: 30689
            protocol: TCP
          - containerPort: 30690
            hostPort: 30690
            protocol: TCP
          - containerPort: 30691
            hostPort: 30691
            protocol: TCP
          - containerPort: 30692
            hostPort: 30692
            protocol: TCP
          - containerPort: 30693
            hostPort: 30693
            protocol: TCP
          - containerPort: 30694
            hostPort: 30694
            protocol: TCP
          - containerPort: 30695
            hostPort: 30695
            protocol: TCP
          - containerPort: 30696
            hostPort: 30696
            protocol: TCP
          - containerPort: 30697
            hostPort: 30697
            protocol: TCP
          - containerPort: 30698
            hostPort: 30698
            protocol: TCP
          - containerPort: 30699
            hostPort: 30699
            protocol: TCP
          - containerPort: 30700
            hostPort: 30700
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

kubectl apply -f etc/testing/minio.yaml

helm install pachyderm etc/helm/pachyderm -f etc/kind/hostname-values.yaml -f etc/kind/enterprise-key-values.yaml -f etc/kind/values.yaml

#helm install pachyderm ../pachyderm-envoy/etc/helm/pachyderm -f etc/kind/hostname-values.yaml -f etc/kind/enterprise-key-values.yaml -f etc/kind/values.yaml

kubectl rollout status deployment pachd --timeout=10s

pachctl config set context kind --overwrite <<EOF
{"pachd_address": "grpc://localhost:80", "session_token": "V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6"}
EOF

pachctl config set context kind-direct --overwrite <<EOF
{"pachd_address": "grpc://localhost:36050", "session_token": "V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6"}
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

mc alias set kind-direct http://localhost:30600 V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6 V4Ptmxchcj04TY5vmngJAD0RiuJ3JYc6

mc ls kind-direct
