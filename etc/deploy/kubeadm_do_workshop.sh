#! /bin/bash

# make sure the rest of the node is ready
sleep 2m

# set kubectl config locations
echo "export KUBECONFIG=/root/admin.conf" >> /root/.bashrc
echo "export KUBECONFIG=/home/pachrat/admin.conf" >> /home/pachrat/.bashrc

# start kubernetes
kubeadm init

# connect kubectl
sudo cp /etc/kubernetes/admin.conf /root/
sudo cp /etc/kubernetes/admin.conf /home/pachrat/
sudo chown "$(id -u):$(id -g)" /root/admin.conf
sudo chown pachrat /home/pachrat/admin.conf
export KUBECONFIG=/root/admin.conf

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

# set password auth
usermod -aG sudo pachrat
sudo sed -i -- 's/PasswordAuthentication no/PasswordAuthentication yes/g' /etc/ssh/sshd_config
sudo service ssh restart
