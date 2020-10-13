#!/bin/bash
set -euo pipefail
export MNIST_PIPELINE_VERSION=$(git rev-parse HEAD)-$(date +%s)
cat pipeline.yaml.template | envsubst > pipeline.yaml
docker build -t quay.io/lukemarsden/mnist_pipeline:$MNIST_PIPELINE_VERSION .
docker push quay.io/lukemarsden/mnist_pipeline:$MNIST_PIPELINE_VERSION
