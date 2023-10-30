#!/bin/bash

echo "=== TEST FAILED OR TIMED OUT, DUMPING DEBUG INFO ==="

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
  'docker ps'
  'pachctl list pipeline'
  'pachctl list job --no-pager'
  'kubectl version'
  'kubectl get all --all-namespaces'
  'kubectl get pvc --all-namespaces'
  'kubectl get persistentvolume --all-namespaces'
  'kubectl get secret --all-namespaces'
  'kubectl get pods -o wide --all-namespaces'
  'kubectl describe pod -l suite=pachyderm --namespace=test-cluster-1'
  'kubectl describe pod -l suite=pachyderm --namespace=test-cluster-2'
  'kubectl describe pod -l suite=pachyderm --namespace=test-cluster-3'
  'kubectl describe pod -l suite=pachyderm --namespace=test-cluster-4'
  'kubectl describe pod -l suite=pachyderm --namespace=test-cluster-5'
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
  'kubectl logs --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-1 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-2 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-3 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-4 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-5 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs --namespace=test-cluster-6 --tail=1500 -l suite=pachyderm --previous --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-1 --tail=1500 --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-2 --tail=1500 --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-3 --tail=1500 --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-4 --tail=1500 --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-5 --tail=1500 --prefix=true'
  'kubectl logs postgres-0 --namespace=test-cluster-6 --tail=1500 --prefix=true'
  'kubectl logs -l k8s-app=kube-dns -n kube-system'
  'kubectl logs -l k8s-app=kube-dns -n kube-system --previous'
  'kubectl logs storage-provisioner -n kube-system'
  'kubectl logs storage-provisioner -n kube-system --previous'
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
