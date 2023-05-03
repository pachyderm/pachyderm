#!/bin/sh

set -ve

VERSION=v1.19.0

# wait for docker or timeout
timeout=120
while ! docker version >/dev/null 2>&1; do
  timeout=$((timeout - 1))
  if [ $timeout -eq 0 ]; then
    echo "Timed out waiting for docker daemon"
    exit 1
  fi
  sleep 1
done

# start minikube with pod security admission plugin
minikube start \
    --vm-driver=docker \
    --kubernetes-version=${VERSION} \
    --cpus=4 \
    --memory=12Gi \
    --wait=all \

# install gatekeeper
kubectl apply -f etc/testing/gatekeeper.yaml
kubectl -n gatekeeper-system wait --for=condition=ready pod -l control-plane=audit-controller --timeout=5m
# install gatekeeper OPA Templates
kubectl apply -f etc/testing/opa-policies/
sleep 20
#Install gatekeeper OPA constraints 
kubectl apply -f etc/testing/opa-constraints.yaml

./etc/testing/circle/build.sh

kubectl apply -f etc/testing/minio.yaml

# Run TestSimplePipelineNonRoot TestSimplePipelinePodPatchNonRoot
go test -v ./src/server -run NonRoot -v=test2json tags=k8s -cover -test.gocoverdir="$TEST_RESULTS" -covermode=atomic -coverpkg=./...| stdbuf -i0 tee -a /tmp/go-test-results.txt
