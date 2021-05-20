#!/bin/bash

set -ex

source "$(dirname "$0")/env.sh"

VM_IP="$(minikube ip)"
export VM_IP

PACH_PORT="30650"
export PACH_PORT

POSTGRES_SERVICE_HOST="$(minikube ip)"
export POSTGRES_SERVICE_HOST

POSTGRES_SERVICE_PORT=32228
export POSTGRES_SERVICE_PORT

TESTFLAGS="-v | stdbuf -i0 tee -a /tmp/results"
export TESTFLAGS

# make launch-kube connects with kubernetes, so it should just be available
minikube status
kubectl version

echo "Running test suite based on BUCKET=$BUCKET"

function test_bucket {
    set +x
    package="${1}"
    target="${2}"
    bucket_num="${3}"
    num_buckets="${4}"
    if (( bucket_num == 0 )); then
        echo "Error: bucket_num should be > 0, but was 0" >/dev/stderr
        exit 1
    fi

    echo "Running bucket $bucket_num of $num_buckets"
    # shellcheck disable=SC2207
    tests=( $(go test -v  "${package}" -list ".*" | grep -v '^ok' | grep -v '^Benchmark') )
    # Add anchors for the regex so we don't run collateral tests
    tests=( "${tests[@]/#/^}" )
    tests=( "${tests[@]/%/\$\$}" )
    total_tests="${#tests[@]}"
    # Determine the offset and length of the sub-array of tests we want to run
    # The last bucket may have a few extra tests, to accommodate rounding
    # errors from bucketing:
    let "bucket_size=total_tests/num_buckets" \
        "start=bucket_size * (bucket_num-1)" \
        "bucket_size+=bucket_num < num_buckets ? 0 : total_tests%num_buckets"
    test_regex="$(IFS=\|; echo "${tests[*]:start:bucket_size}")"
    echo "Running ${bucket_size} tests of ${total_tests} total tests"
    make RUN="-run='${test_regex}'" "${target}"
    set -x
}


# Clean cached test results
go clean -testcache

case "${BUCKET}" in
 MISC)
    make lint
    make enterprise-code-checkin-test
    make test-cmds
    make test-libs
    make test-proto-static
    make test-transaction
    make test-deploy-manifests
    make test-s3gateway-unit
    make test-enterprise
    make test-worker
    if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
        # these tests require secure env vars to run, which aren't available
        # when the PR is coming from an outside contributor - so we just
        # disable them
        make test-tls
        make test-vault
    fi
    ;;
 ADMIN)
    make test-admin
    ;;
 EXAMPLES)
    echo "Running the example test suite"
    ./etc/testing/examples.sh
    ;;
 PFS)
    make test-pfs-server
    make test-pfs-storage
    ;;
 PPS?)
    pushd etc/testing/images/ubuntu_with_s3_clients
    make push-to-minikube
    popd
    make docker-build-kafka
    bucket_num="${BUCKET#PPS}"
    test_bucket "./src/server" test-pps "${bucket_num}" "${PPS_BUCKETS}"
    if [[ "${bucket_num}" -eq "${PPS_BUCKETS}" ]]; then
      go test -v -count=1 ./src/server/pps/server -timeout 3600s
    fi
    ;;
 AUTH?)
    make launch-dex
    bucket_num="${BUCKET#AUTH}"
    test_bucket "./src/server/auth/server/testing" test-auth "${bucket_num}" "${AUTH_BUCKETS}"
    set +x
    ;;
OBJECT)
    make test-object-clients
    ;;
 *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
