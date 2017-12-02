#!/bin/sh

set -Eex

echo 'pre resolv conf:'
cat /etc/resolve.conf
sudo echo 'nameserver 8.8.8.8' > /etc/resolv.conf
echo 'post resolv conf:'
cat /etc/resolve.conf

sudo systemctl stop systemd-resolved
sudo systemctl disable systemd-resolved

# Parse flags
VERSION=v1.8.0
while getopts ":v:" opt; do
  case "${opt}" in
    v)
      VERSION="v${OPTARG}"
      ;;
    \?)
      echo "Invalid argument: ${opt}"
      exit 1
      ;;
  esac
done

#sudo CHANGE_MINIKUBE_NONE_USER=true minikube start --vm-driver=none --kubernetes-version="${VERSION}" --extra-config=kubelet.CertDir=/var/lib/kubelet/pki --extra-config=apiserver.ServiceClusterIpRange=10.0.0.1/12
sudo CHANGE_MINIKUBE_NONE_USER=true minikube start --vm-driver=none --kubernetes-version="${VERSION}" --extra-config=kubelet.CertDir=/var/lib/kubelet/pki
until kubectl version 2>/dev/null >/dev/null; do sleep 5; done
