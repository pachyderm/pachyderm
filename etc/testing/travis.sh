#!/bin/bash

set -ex

(
    # Stop Travis from timing us out in 10m when we want to get retried after 20m
    set +x
    while true; do
        echo "liveness ping $(date)"
        sleep 60
    done
) &

# Note that we update the `PATH` to include
# `~/cached-deps` in `.travis.yml`, but this doesn't update the PATH for
# calls using `sudo`. If you need to make a `sudo` call to a binary in
# `~/cached-deps`, you'll need to explicitly set the path like so:
#
#     sudo env "PATH=$PATH" minikube foo

# In case minikube delete doesn't work (see minikube#2519)
for C in $(docker ps -aq); do docker rm -f "$C"; done || true

# Belt and braces
sudo rm -rf /etc/kubernetes /data/minikube /var/lib/minikube

# In case we're retrying on a new cluster
rm -f "${HOME}"/.pachyderm/config.json

make launch-kube
sleep 5

# Wait until a connection with kubernetes has been established
echo "Waiting for connection to kubernetes..."
max_t=90
wheel='\|/-';
until {
  minikube status &>/dev/null
  kubectl version &>/dev/null
}; do
    if ((max_t-- <= 0)); then
        echo "Could not connect to minikube"
        echo "minikube status --alsologtostderr --loglevel=0 -v9:"
        echo "==================================================="
        minikube status --alsologtostderr --loglevel=0 -v9
        exit 1
    fi
    echo -en "\e[G$${wheel:0:1}";
    # shellcheck disable=SC2034
    wheel="$${wheel:1}$${wheel:0:1}";
    sleep 1;
done
minikube status
kubectl version

echo "Running test suite based on BUCKET=$BUCKET"

if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
    # Pull the pre-built images. This is only done if we have access to the
    # secret env vars, because otherwise the build step would've had to be
    # skipped.
    make install
    version=$(pachctl version --client-only)
    docker pull "pachyderm/pachd:${version}"
    docker tag "pachyderm/pachd:${version}" "pachyderm/pachd:local"
    docker pull "pachyderm/worker:${version}"
    docker tag "pachyderm/worker:${version}" "pachyderm/worker:local"
else
    make docker-build
    # push pipeline build images
    pushd etc/pipeline-build
        make push-to-minikube
    popd
fi

make launch-loki

for i in $(seq 3); do
    make clean-launch-dev || true # may be nothing to delete
    make launch-dev && break
    (( i < 3 )) # false if this is the last loop (causes exit)
    sleep 10
done

pachctl config update context "$(pachctl config get active-context)" --pachd-address="$(minikube ip):30650"

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
    bucket_num="${BUCKET#AUTH}"
    test_bucket "./src/server/auth/server/testing" test-auth "${bucket_num}" "${AUTH_BUCKETS}"
    set +x
    ;;
 *)
    echo "Unknown bucket"
    exit 1
    ;;
esac
