#!/bin/bash

set -xeuo pipefail

# Get a kubernetes cluster
echo "==== KUBECONFIG ===="
cat kubeconfig
echo "===================="

KUBECONFIG="$(pwd)/kubeconfig"
export KUBECONFIG

echo "Fetching new code in VM"
time ./etc/testing/testctl-ssh.sh -- bash -c "set -x; cd project/pachyderm; pwd; git fetch; git reset --hard HEAD; git checkout ${CIRCLE_SHA1}"
echo "Finished fetching new code in VM"

#echo "Copying context to runner."
## trailing slash means _contents_ of this directory are copied _into_ target
## directory.
#time ./etc/testing/testctl-rsync.sh "$(pwd)"/ /root/project/pachyderm
#echo "Finished copying context."

# NB: https://serverfault.com/questions/482907/setting-a-variable-for-a-given-ssh-host

ENV_VARS=(PPS_BUCKETS AUTH_BUCKETS GOPROXY ENT_ACT_CODE BUCKET RUN_BAD_TESTS DOCKER_PWD)

# For object tests, provide the parameters and credentials for running against object storage providers
if [[ "$BUCKET" == "OBJECT" ]]; then
    ENV_VARS+=(AMAZON_CLIENT_ID AMAZON_CLIENT_SECRET AMAZON_CLIENT_BUCKET AMAZON_CLIENT_REGION)
    ENV_VARS+=(ECS_CLIENT_ID ECS_CLIENT_SECRET ECS_CLIENT_BUCKET ECS_CLIENT_CUSTOM_ENDPOINT)
    ENV_VARS+=(GOOGLE_CLIENT_BUCKET GOOGLE_CLIENT_CREDS)
    ENV_VARS+=(GOOGLE_CLIENT_HMAC_ID GOOGLE_CLIENT_HMAC_SECRET GOOGLE_CLIENT_REGION)
    ENV_VARS+=(MICROSOFT_CLIENT_ID MICROSOFT_CLIENT_SECRET MICROSOFT_CLIENT_CONTAINER)
fi

TESTCTL_OPTIONS=()
for VAR in "${ENV_VARS[@]}"; do
    TESTCTL_OPTIONS+=("-o" "SendEnv=$VAR")
done

echo "Starting test $BUCKET."
time ./etc/testing/testctl-ssh.sh "${TESTCTL_OPTIONS[@]}" \
    -- ./project/pachyderm/etc/testing/circle_tests_inner.sh "$@"
echo "Finished test $BUCKET."
