#!/bin/bash

set -ex

echo "=== TEST FAILED OR TIMED OUT, DUMPING DEBUG INFO ==="
kubectl get all --all-namespaces

# TODO: Extend this to show kubectl describe output for failed pods, this will
# probably show why things are hanging.

kubectl version
kubectl get all --namespace kafka
kubectl describe pod -l app=pachd
kubectl describe pod -l suite=pachyderm,app=etcd
kubectl logs -l app=pachd | tail -n 100
sudo dmesg | tail -n 20
{ minikube logs | tail -n 20; } || true
top -b -n 1 | head -n 20
df -h
