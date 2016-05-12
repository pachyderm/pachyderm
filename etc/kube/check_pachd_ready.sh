#!/bin/bash


results=`kubectl get pods \
  -l app=pachd \
  -o jsonpath='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'`

if [ -z $results ]; then \
  echo "Empty result"; \
  echo $results; \
  exit 1; \
fi

echo $results \
  | tr ';' "\n" \
  | grep "Ready=" \
  | while read line; do \
    echo $line | grep True; \
    ready=$?; \
    echo "line: $line"; \
    echo "ready? ($ready)"; \
    if [[ $ready -ne "0" ]]; then \
        echo "Not ready, exiting"; \
        exit 1; \
    fi; \
  done


echo "Success"
exit 0
