#!/bin/sh

set -ve

# Note: Minikube needs to be running before this script is called

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
