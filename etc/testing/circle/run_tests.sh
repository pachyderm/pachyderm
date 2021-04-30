#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$(pwd)/cached-deps:$GOPATH/bin:$PATH

export VM_IP="$(minikube ip)"
export PACH_PORT="30650"
export ENTERPRISE_PORT="31650"

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
    total_tests="${#tests[@]}"
    # Determine the offset and length of the sub-array of tests we want to run
    # The last bucket may have a few extra tests, to accommodate rounding
    # errors from bucketing:
    let "bucket_size=total_tests/num_buckets" \
        "start=bucket_size * (bucket_num-1)" \
        "bucket_size+=bucket_num < num_buckets ? 0 : total_tests%num_buckets"
    test_regex="$(IFS=\|; echo "${tests[*]:start:bucket_size}")"
    echo "Running ${bucket_size} tests of ${total_tests} total tests"
    make RUN="-run=\"${test_regex}\"" "${target}"
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
    make test-worker
    if [[ "${TRAVIS_SECURE_ENV_VARS:-""}" == "true" ]]; then
        # these tests require secure env vars to run, which aren't available
        # when the PR is coming from an outside contributor - so we just
        # disable them
        make test-tls
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
      go test -v -count=1 ./src/server/pps/server -timeout 300s
    fi
    ;;
  AUTH)
    make test-identity
    make test-auth
    ;;
  OBJECT)
    make test-object-clients
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
  *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
