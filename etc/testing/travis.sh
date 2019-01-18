#!/bin/bash

set -e

# Make sure cache dir exists and is writable
mkdir -p ~/.cache/go-build
sudo chown -R `whoami` ~/.cache/go-build

minikube delete || true  # In case we get a recycled machine
make launch-kube
sleep 5

# Wait until a connection with kubernetes has been established
echo "Waiting for connection to kubernetes..."
max_t=90
WHEEL="\|/-";
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
            test-pfs-cmds test-deploy-cmds test-libs test-vault test-auth \
            test-enterprise test-worker test-admin
    else
        echo "Running the misc test suite with some tests disabled because secret env vars have not been set"

        # Do not run some tests when we don't have access to secret
        # credentials
        make lint enterprise-code-checkin-test docker-build test-pfs-server \
            test-pfs-cmds test-deploy-cmds test-libs test-admin
    fi
elif [[ $PPS_SUITE -eq 0 ]]; then
    PART="${BUCKET#PPS}"
    # >>>
    if [[ "${PART}" -ne 3 ]]; then exit 0; fi
    # >>>
    NUM_BUCKETS=`cat etc/build/PPS_BUILD_BUCKET_COUNT`
    echo "Running pps test suite, part $PART of $NUM_BUCKETS"
    TESTS=( $(go test -v  ./src/server/ -list ".*" | grep -v ok | grep -v Benchmark) )
    NUM_TESTS="${#TESTS[@]}"
    SIZE=$(( NUM_TESTS/NUM_BUCKETS ))
    MIN=$(( SIZE * (PART-1) ))
    if [[ $PART -eq $NUM_BUCKETS ]]; then
        # last bucket picks up extra tests
        SIZE=$((SIZE + (NUM_TESTS%NUM_BUCKETS) ))
    fi
    RUN="$(IFS=\|; echo "${TESTS[*]:MIN:SIZE}")"
    echo "Running ${SIZE} tests of ${NUM_TESTS} total tests"

    # apply our own timeout on top of travis, so that if we hit it we emit
    # 'kubectl get all' and such
    timeout 5m make RUN=-run=\"$RUN\" test-pps-helper
else
    echo "Unknown bucket"
    exit 1
fi

# Disable aws CI for now, see:
# https://github.com/pachyderm/pachyderm/issues/2109
exit 0

echo "Running local tests"
make local-test

echo "Running aws tests"

sudo pip install awscli
sudo apt-get install realpath uuid jq
wget --quiet https://github.com/kubernetes/kops/releases/download/1.7.10/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops

# Use the secrets in the travis environment to setup the aws creds for the aws command:
echo -e "${AWS_ACCESS_KEY_ID}\n${AWS_SECRET_ACCESS_KEY}\n\n\n" \
    | aws configure

make install
echo "pachctl is installed here:"
which pachctl

# Travis doesn't come w an ssh key
# kops needs one in place (because it enables ssh access to nodes w it)
# for now we'll just generate one on the fly
# travis supports adding a persistent one if we pay: https://docs.travis-ci.com/user/private-dependencies/#Generating-a-new-key
if [[ ! -e ${HOME}/.ssh/id_rsa ]]; then
    ssh-keygen -t rsa -b 4096 -C "buildbot@pachyderm.io" -f ${HOME}/.ssh/id_rsa -N ''
    echo "generated ssh keys:"
    ls ~/.ssh
    cat ~/.ssh/id_rsa.pub
fi

# Need to login so that travis can push the bench image
docker login -u pachydermbuildbot -p ${DOCKER_PWD}

# Run tests in the cloud
make aws-test
