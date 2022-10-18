#!/bin/bash
set -xeuo pipefail
export IMAGE=quay.io/lukemarsden/spark_s3_demo:v0.0.13
docker buildx build -t $IMAGE .
docker push $IMAGE