#!/bin/bash

set -ex

# Make sure cache dirs exist and are writable
mkdir -p ~/.cache/go-build
sudo chown -R `whoami` ~/.cache/go-build
sudo chown -R `whoami` ~/cached-deps

# Note that this script executes as the user `travis`, vs the pre-install
# script which executes as `root`. Without `chown`ing `~/cached-deps` (where
# we store cacheable binaries), any calls to those binaries would fail because
# they're otherwised owned by `root`.
#
# To further complicate things, we update the `PATH` to include
# `~/cached-deps` in `.travis.yml`, but this doesn't update the PATH for
# calls using `sudo`. If you need to make a `sudo` call to a binary in
# `~/cached-deps`, you'll need to explicitly set the path like so:
#
#     sudo env "PATH=$PATH" minikube foo
#
kubectl version --client
etcdctl --version

minikube delete || true  # In case we get a recycled machine
make launch-kube
sleep 5

# Wait until a connection with kubernetes has been established
echo "Waiting for connection to kubernetes..."
max_t=90
WHEEL='\|/-';
until {
  minikube status 2>&1 >/dev/null
  kubectl version 2>&1 >/dev/null
}; do
    if ((max_t-- <= 0)); then
        echo "Could not connect to minikube"
        echo "minikube status --alsologtostderr --loglevel=0 -v9:"
        echo "==================================================="
        minikube status --alsologtostderr --loglevel=0 -v9
        exit 1
    fi
    echo -en "\e[G$${WHEEL:0:1}";
    WHEEL="$${WHEEL:1}$${WHEEL:0:1}";
    sleep 1;
done
minikube status
kubectl version

echo "Running test suite based on BUCKET=$BUCKET"

PPS_SUITE=`echo $BUCKET | grep PPS > /dev/null; echo $?`

make install
make docker-build
make docker-build-kafka
for i in $(seq 3); do
    make clean-launch-dev || true # may be nothing to delete
    make launch-dev && break
    (( i < 3 )) # false if this is the last loop (causes exit)
    sleep 10
done

go install ./src/testing/match

if [[ "$BUCKET" == "MISC" ]]; then
    if [[ "$TRAVIS_SECURE_ENV_VARS" == "true" ]]; then
        echo "Running the full misc test suite because secret env vars exist"

        make lint enterprise-code-checkin-test docker-build test-pfs-server \
            test-pfs-cmds test-pfs-storage test-deploy-cmds test-libs test-vault test-auth \
            test-enterprise test-worker test-admin test-s3gateway-integration \
            test-proto-static test-transaction
    else
        echo "Running the misc test suite with some tests disabled because secret env vars have not been set"

        # Do not run some tests when we don't have access to secret
        # credentials
        make lint enterprise-code-checkin-test docker-build test-pfs-server \
            test-pfs-cmds test-pfs-storage test-deploy-cmds test-libs test-admin \
            test-s3gateway-integration
    fi
elif [[ "$BUCKET" == "EXAMPLES" ]]; then
    echo "Running the example test suite"
    ./etc/testing/examples.sh
elif [[ $PPS_SUITE -eq 0 ]]; then
    set +x
    PART=`echo $BUCKET | grep -Po '\d+'`
    NUM_BUCKETS=`cat etc/build/PPS_BUILD_BUCKET_COUNT`
    echo "Running pps test suite, part $PART of $NUM_BUCKETS"
    LIST=`go test -v  ./src/server/ -list ".*" | grep -v ok | grep -v Benchmark`
    COUNT=`echo $LIST | tr " " "\n" | wc -l`
    BUCKET_SIZE=$(( $COUNT / $NUM_BUCKETS ))
    MIN=$(( $BUCKET_SIZE * $(( $PART - 1 )) ))
    #The last bucket may have a few extra tests, to accommodate rounding errors from bucketing:
    MAX=$COUNT
    if [[ $PART -ne $NUM_BUCKETS ]]; then
        MAX=$(( $MIN + $BUCKET_SIZE ))
    fi

    RUN=""
    INDEX=0

    for test in $LIST; do
        if [[ $INDEX -ge $MIN ]] && [[ $INDEX -lt $MAX ]] ; then
            if [[ "$RUN" == "" ]]; then
                RUN=$test
            else
                RUN="$RUN|$test"
            fi
        fi
        INDEX=$(( $INDEX + 1 ))
    done
    echo "Running $( echo $RUN | tr '|' '\n' | wc -l ) tests of $COUNT total tests"
    set -x
    make RUN=-run=\"$RUN\" test-pps-helper
else
    echo "Unknown bucket"
    exit 1
fi
