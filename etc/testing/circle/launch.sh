#!/bin/bash

set -ex

# deploy object storage
kubectl apply -f etc/testing/minio.yaml
