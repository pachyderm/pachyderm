#!/bin/bash

set -x

echo
kubectl version
echo
kubectl get all
echo
kubectl get all --namespace kafka
echo
kubectl describe pod -l app=pachd
echo
kubectl describe pod -l suite=pachyderm,app=etcd
echo
kubectl logs -l app=pachd | tail -n 100
echo
kubectl logs -l app=pachd --previous | tail -n 100
