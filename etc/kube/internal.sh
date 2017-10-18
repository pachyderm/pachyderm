set -Ee

#cat /etc/kubernetes/manifests/master.json | jq '.spec.containers[1].command |= . + [ "--storage-backend=etcd2" ]' >/etc/kubernetes/manifests/master2.json
#mv /etc/kubernetes/manifests/master2.json /etc/kubernetes/manifests/master.json
echo 'kubeconfig:'
cat $1

#    --pod-manifest-path=/etc/kubernetes/manifests \

/hyperkube kubelet \
    --containerized \
    --kubeconfig=$1 \
    --hostname-override="127.0.0.1" \
    --address="0.0.0.0" \
    --cluster_dns=10.0.0.10 \
    --cluster_domain=cluster.local \
    --allow-privileged=true \
    --feature-gates="Accelerators=true" \
    --fail-swap-on=false
