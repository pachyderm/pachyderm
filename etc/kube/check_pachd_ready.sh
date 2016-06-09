#!/bin/bash


results=`kubectl get pods \
  -l app=pachd \
  -o jsonpath='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}' \
  2>/dev/null | tr ';' "\n"`

if [ -z "$results" ]; then
  echo "Empty result"
  echo $results
  exit 1
fi


readyPods=`echo $results | tr ' ' "\n" | grep "Ready=True" | wc -l`
allPods=`echo $results | tr ' ' "\n" | grep "Ready=" | wc -l`

if [ "$allPods" -eq 0 ]; then
    echo "No pods found yet"
    exit 1
fi

if [ "$readyPods" -ne "$allPods" ]; then
    exit 1
fi

echo "All pods are ready."
exit 0
