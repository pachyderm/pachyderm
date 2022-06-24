#!/bin/bash

echo "=== TEST FAILED OR TIMED OUT, DUMPING DEBUG INFO ==="

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

# TODO: Extend this to show kubectl describe output for failed pods, this will
# probably show why things are hanging.

# SC2016 blocks variables in single-quoted strings, but these are 'eval'ed below
# shellcheck disable=SC2016
cmds=(
  'pachctl version'
  'pachctl list repo'
  'pachctl list repo --raw | jq -r ".repo.name" | while read r; do
     echo "---";
     pachctl inspect repo "${r}";
     echo "Commits:";
     pachctl list commit "${r}";
     echo;
   done'
  'pachctl list pipeline'
  'pachctl list job --no-pager'
  'kubectl version'
  'kubectl get all --all-namespaces'
  'kubectl describe pod -l suite=pachyderm,app=pachd'
  'kubectl describe pod -l suite=pachyderm,app=etcd'
  'curl -s http://$(minikube ip):30656/metrics'
  # Set --tail b/c by default 'kubectl logs' only outputs 10 lines if -l is set
  'kubectl logs --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-1 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-2 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-3 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-4 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-5 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  'kubectl logs --namespace=test-cluster-6 --tail=1500 -l suite=pachyderm --all-containers=true --prefix=true'
  # if pachd restarted
  'kubectl logs --tail=1500 -l suite=pachyderm,app=pachd --previous --prefix=true'
  'sudo dmesg | tail -n 40'
  'minikube logs | tail -n 100'
  'top -b -n 1 | head -n 40'
  'df -h'
)
for c in "${cmds[@]}"; do
  echo "======================================================================"
  echo "${c}"
  echo "----------------------------------------------------------------------"
  eval "${c}"
done
echo "======================================================================"
