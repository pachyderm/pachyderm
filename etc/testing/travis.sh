#!/bin/bash

set -ex

export GOPATH=/home/circleci/.go_workspace
export PATH=$(pwd):$GOPATH/bin:$PATH

export PACH_PORT="30650"
export ENTERPRISE_PORT="31650"

# Repeatedly restart minikube until it comes up. This corrects for an issue in
# Travis, where minikube will get stuck on startup and never recover
while true; do
  # In case minikube delete doesn't work (see minikube#2519)
  for C in $(docker ps -aq); do docker rm -f "$C"; done || true

  # Belt and braces
  sudo rm -rf \
    /etc/kubernetes \
    /data/minikube \
    /var/lib/minikube \
    "${HOME}"/.pachyderm/config.json  # In case we're retrying on a new cluster

  if make launch-kube; then
    break
  fi
done

# make launch-kube connects with kubernetes, so it should just be available
minikube status
kubectl version

export VM_IP="$(minikube ip)"

echo "Running test suite based on BUCKET=$BUCKET"

for image in $(ls /tmp/docker-cache); do 
    docker load -i /tmp/docker-cache/$image
    ( eval $(minikube docker-env) && docker load -i /tmp/cache/$image)
done

make install
make docker-build
minikube cache add pachyderm/pachd:local
minikube cache add pachyderm/worker:local

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
