#!/bin/bash
set -Ee

jq '.spec.containers[1].command |= . + [ "--storage-backend=etcd2" ]' </etc/kubernetes/manifests/master.json >/etc/kubernetes/manifests/master2.json
mv /etc/kubernetes/manifests/master2.json /etc/kubernetes/manifests/master.json
/hyperkube kubelet \
    --containerized \
    --hostname-override="127.0.0.1" \
    --address="0.0.0.0" \
    --api-servers=http://localhost:8080 \
    --cluster_dns=10.0.0.10 \
    --cluster_domain=cluster.local \
    --pod-manifest-path=/etc/kubernetes/manifests \
    --allow-privileged=true \
    --feature-gates="Accelerators=true"
