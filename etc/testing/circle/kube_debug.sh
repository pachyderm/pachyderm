#!/bin/bash

echo "=== TEST FAILED OR TIMED OUT, DUMPING DEBUG INFO ==="

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
  'pachctl list job'
  'kubectl version'
  'kubectl get all --all-namespaces'
  'kubectl describe pod -l suite=pachyderm,app=pachd'
  'kubectl describe pod -l suite=pachyderm,app=etcd'
  # Set --tail b/c by default 'kubectl logs' only outputs 10 lines if -l is set
  'kubectl logs --tail=500 -l suite=pachyderm,app=pachd'
  'kubectl logs --tail=1500 -l suite=pachyderm,app=pachd --previous # if pachd restarted'
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
