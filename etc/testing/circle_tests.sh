#!/bin/bash

set -euo pipefail

# Get a kubernetes cluster
# Specify the slot so that future builds on this branch+suite id automatically
# clean up previous VMs
BRANCH="${CIRCLE_BRANCH:-$GITHUB_REF}"
echo "Getting VM."
time testctl get --config .testfaster.yml --slot "${BRANCH},${BUCKET}"
echo "Finished getting VM."

echo "==== KUBECONFIG ===="
cat kubeconfig
echo "===================="

KUBECONFIG="$(pwd)/kubeconfig"
export KUBECONFIG

echo "Copying context to runner."
time ./etc/testing/testctl-rsync.sh . /root/project
echo "Finished copying context."

# NB: https://serverfault.com/questions/482907/setting-a-variable-for-a-given-ssh-host

echo "Starting test $BUCKET."
#time ./etc/testing/testctl-ssh.sh \
#    -o SendEnv=PPS_BUCKETS \
#    -o SendEnv=AUTH_BUCKETS \
#    -o SendEnv=GOPROXY \
#    -o SendEnv=ENT_ACT_CODE \
#    -o SendEnv=BUCKET \
#    -o SendEnv=CIRCLE_BRANCH \
#    -o SendEnv=RUN_BAD_TESTS \
#    -- ./project/etc/testing/circle_tests_inner.sh "$@"
time ./etc/testing/testctl-ssh.sh \
    -o SendEnv=PPS_BUCKETS \
    -- ./project/etc/testing/circle_tests_inner.sh "$@"
echo "Finished test $BUCKET."
