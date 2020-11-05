#!/bin/bash

set -euo pipefail

# In case we're retrying on a new cluster
rm -f "${HOME}"/.pachyderm/config.json

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

# [x]: get docker image over there (well, we did already)
# we assume 'make docker-build' has been done by a previous build step. see .circleci/config.yml

# [ ]: send files across TODO: make this use rsync --delete
echo "Copying context to runner."
time ./etc/testing/testctl-rsync.sh . /root/project
echo "Finished copying context."

# [x]: pass environment variables through, at least ENT_ACT_CODE, BUCKET
# [x]: pass arguments over

# workaround https://serverfault.com/questions/482907/setting-a-variable-for-a-given-ssh-host

echo "Starting test $BUCKET."
time ./etc/testing/testctl-ssh.sh \
    -o SendEnv=PPS_BUCKETS \
    -o SendEnv=AUTH_BUCKETS \
    -o SendEnv=GOPROXY \
    -o SendEnv=ENT_ACT_CODE \
    -o SendEnv=BUCKET \
    -- ./project/etc/testing/circle_tests_inner.sh "$@"
echo "Finished test $BUCKET."
