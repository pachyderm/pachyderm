#!/bin/bash


cmds=(
  'kubectl version'
  'kubectl get all'
  'kubectl get all --namespace kafka'
  'kubectl describe pod -l app=pachd'
  'kubectl describe pod -l suite=pachyderm,app=etcd'
  'kubectl logs --tail=100 -l app=pachd'
  'kubectl logs --tail=100 -l app=pachd --previous'
)
for c in "${cmds[@]}"; do
  echo "======================================================================"
  echo "${c}"
  echo "----------------------------------------------------------------------"
  eval "${c}"
done
