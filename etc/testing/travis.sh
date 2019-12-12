#!/bin/bash

set -ex

# Note that we update the `PATH` to include
# `~/cached-deps` in `.travis.yml`, but this doesn't update the PATH for
# calls using `sudo`. If you need to make a `sudo` call to a binary in
# `~/cached-deps`, you'll need to explicitly set the path like so:
#
#     sudo env "PATH=$PATH" minikube foo

kubectl version --client
etcdctl --version

kind create cluster
export KUBECONFIG="$(kind get kubeconfig-path)"

echo "Running test suite based on BUCKET=$BUCKET"

make docker-build

kind load docker-image pachyderm/pachd:local
kind load docker-image pachyderm/worker:local

# fix for docker build process messing with permissions
sudo chown -R ${USER}:${USER} ${GOPATH}

kubectl apply -f pachconfig.yaml
until timeout 1s ./etc/kube/check_ready.sh app=pachd; do sleep 1; done

pachctl config update context `pachctl config get active-context` --pachd-address=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' kind-control-plane):30650

function test_bucket {
    set +x
    package="${1}"
    target="${2}"
    bucket="${3}"
    num_buckets="${4}"
    if (( bucket == 0 )); then
        echo "Error: bucket should be > 0, but was 0" >/dev/stderr
        exit 1
    fi

    echo "Running bucket $bucket of $num_buckets"
    tests=( $(go test -v  "${package}" -list ".*" | grep -v ok | grep -v Benchmark) )
    total_tests="${#tests[@]}"
    # Determine the offset and length of the sub-array of tests we want to run
    # The last bucket may have a few extra tests, to accommodate rounding errors from bucketing:
    let \
        "bucket_size=total_tests/num_buckets" \
        "start=bucket_size * (bucket-1)" \
        "bucket_size+=bucket < num_buckets ? 0 : total_tests%num_buckets"
    test_regex="$(IFS=\|; echo "${tests[*]:start:bucket_size}")"
    echo "Running ${bucket_size} tests of ${total_tests} total tests"
    make RUN="-run=\"${test_regex}\"" "${target}"
    set -x
}

# Clean cached test results
go clean -testcache

case "${BUCKET}" in
 MISC)
    if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
        echo "Running the full misc test suite because secret env vars exist"
        make lint
        make enterprise-code-checkin-test
        make test-cmds
        make test-libs
        make test-tls
        make test-vault
        make test-enterprise
        make test-worker
        make test-s3gateway-integration
        make test-proto-static
        make test-transaction
        make test-deploy-manifests
    else
        echo "Running the misc test suite with some tests disabled because secret env vars have not been set"
        make lint
        make enterprise-code-checkin-test
        make test-cmds
        make test-libs
        make test-tls
        make test-proto-static
        make test-transaction
        make test-deploy-manifests
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
    make docker-build-kafka
    bucket_num="${BUCKET#PPS}"
    test_bucket "./src/server" test-pps "${bucket_num}" "${PPS_BUCKETS}"
    if [[ "${bucket_num}" -eq "${PPS_BUCKETS}" ]]; then
      go test -v -count=1 ./src/server/pps/server -timeout 3600s
    fi
    ;;
 AUTH?)
    bucket_num="${BUCKET#AUTH}"
    test_bucket "./src/server/auth/server" test-auth "${bucket_num}" "${AUTH_BUCKETS}"
    set +x
    ;;
 *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
