#!/bin/bash

set -exo pipefail

VM_IP="$(minikube ip)"
export VM_IP

ENTERPRISE_PORT="31650"
export ENTEPRRISE_PORT

TESTFLAGS="-v | stdbuf -i0 tee -a /tmp/results"
export TESTFLAGS

# make launch-kube connects with kubernetes, so it should just be available
minikube status
kubectl version

# any tests that build images will do it directly in minikube's docker registry
eval "$(minikube docker-env)"

echo "Running test suite based on BUCKET=$BUCKET"

# Clean cached test results
go clean -testcache

case "${BUCKET}" in
 MISC)
    make lint
    make enterprise-code-checkin-test
    go install -v ./src/testing/match
    make test-cmds
    make test-proto-static
    make test-transaction
    make test-s3gateway-unit
    make test-worker
    bash -ceo pipefail "go test -p 1 -count 1 ./src/server/debug/... ${TESTFLAGS}"
    # these tests require secure env vars to run, which aren't available
    # when the PR is coming from an outside contributor - so we just
    # disable them
    # make test-tls
    ;;
  EXAMPLES)
    echo "Running the example test suite"
    ./etc/testing/examples.sh
    ;;
  PFS)
    make test-pfs-server
    make test-fuse
    ;;
  S3_AUTH)
    export PACH_TEST_WITH_AUTH=1
    go test -count=1 -tags=k8s ./src/server/pps/server/s3g_sidecar_test.go -timeout 420s -v | stdbuf -i0 tee -a /tmp/results
    ;;
  AUTH)
    make test-auth
    make test-admin
    ;;
  ENTERPRISE)
    make test-license
    make test-enterprise
    # Launch a stand-alone enterprise server in a separate namespace
    make launch-enterprise
    echo "{\"pachd_address\": \"grpc://${VM_IP}:${ENTERPRISE_PORT}\", \"source\": 2}" | pachctl config set context "enterprise" --overwrite
    pachctl config set active-enterprise-context enterprise
    make test-enterprise-integration
    ;;
  CONNECTORS)
    make test-connectors
    ;;
  *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
