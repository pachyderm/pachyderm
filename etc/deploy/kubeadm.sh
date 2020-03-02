#! /bin/bash

# set kubectl config locations
echo "export KUBECONFIG=/$USER/admin.conf" >> "/$USER/.bashrc"

# start kubernetes
kubeadm init

# connect kubectl
sudo cp /etc/kubernetes/admin.conf "/$USER/"
sudo chown "$USER" "/$USER/admin.conf"
export KUBECONFIG=/$USER/admin.conf

# master isolation
kubectl taint nodes --all node-role.kubernetes.io/master-

# install networking
kubever=$(kubectl version | base64 | tr -d '\n')
export kubever
kubectl apply -f "https://cloud.weave.works/k8s/net?k8s-version=$kubever"

echo "Waiting for networking to come up"
start_time=$(date +%s)
while true; do
  kube_dns_running="$(kubectl get pods --all-namespaces | grep kube-dns | grep Running)"
  if [[ -n "$kube_dns_running" ]]; then
    break;
  fi
  printf "."
  sleep 1
  # shellcheck disable=SC2004
  runtime=$(($(date +%s)-$start_time))
  if [ $runtime -ge 120 ]; then
    (>&2 echo "Timed out waiting for kube-dns (120s)")
    exit 1;
  fi
done

# allow services to act as admin (not great in general, but an easy way
# to make sure pachyderm has access to what it needs in the k8s api)
kubectl create clusterrolebinding serviceaccounts-cluster-admin \
  --clusterrole=cluster-admin \
  --group=system:serviceaccounts

# deploy pachyderm
pachctl deploy local
