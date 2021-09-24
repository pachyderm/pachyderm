#!/bin/sh

set -ve
make pachd-image
make worker-image

# kill pachd
kubectl delete --force $(kubectl get --no-headers=true pods -o name -l=app=pachd)
# kill workers
kubectl delete --force $(kubectl get --no-headers=true pods -o name -l=component=worker)
