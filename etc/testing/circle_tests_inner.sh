#!/bin/bash
set -euo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd "$DIR"/../..

VM_IP="localhost"
export VM_IP
PACH_PORT="30650"
export PACH_PORT
GOPATH="${HOME}/go"
export GOPATH
PATH="${GOPATH}/bin:${PATH}"
export PATH

# Some tests (e.g. TestMigrateFrom1_7) expect
# $GOPATH/src/github.com/pachyderm/pachyderm to point to .
# a hangover from a pre-go mod world.
mkdir -p "$GOPATH"/src/github.com/pachyderm
ln -s "$(pwd)" "$GOPATH"/src/github.com/pachyderm/pachyderm

kubectl version

make install
VERSION=$(pachctl version --client-only)
git commit -am "Commit temp file"
git tag -f -am "Circle CI test v$VERSION" v"$VERSION"
make docker-build
make launch-dev

echo "Running test suite based on BUCKET=$BUCKET"

pachctl config update context "$(pachctl config get active-context)" --pachd-address="${VM_IP}:${PACH_PORT}"

# should be able to connect to pachyderm via localhost
pachctl version

function test_bucket {
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
    if [[ "${TRAVIS_SECURE_ENV_VARS:-""}" == "true" ]]; then
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
      go test -v -count=1 ./src/server/pps/server -timeout 300s
    fi
    ;;
 AUTH?)
    bucket_num="${BUCKET#AUTH}"
    test_bucket "./src/server/auth/server/testing" test-auth "${bucket_num}" "${AUTH_BUCKETS}"
    ;;
 *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
