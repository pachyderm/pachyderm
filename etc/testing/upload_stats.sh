#!/bin/bash

set -xeuo pipefail

KUBECONFIG="$(pwd)/kubeconfig"
export KUBECONFIG

ENV_VARS=(CIRCLE_REPOSITORY_URL CIRCLE_STAGE CIRCLE_BUILD_NUM CIRCLE_SHA1 BIGQUERY_PROJECT BIGQUERY_DATASET BIGQUERY_TABLE TEST_RESULTS_BUCKET GOOGLE_TEST_UPLOAD_CREDS)

TESTCTL_OPTIONS=()
for VAR in "${ENV_VARS[@]}"; do
    TESTCTL_OPTIONS+=("-o" "SendEnv=$VAR")
done

etc/testing/testctl-ssh.sh "${TESTCTL_OPTIONS[@]}" -- ./project/pachyderm/etc/testing/upload_stats_inner.sh

mkdir -p /tmp/test-results

etc/testing/testctl-ssh.sh "${TESTCTL_OPTIONS[@]}" -- ./project/pachyderm/etc/testing/report_junit.sh > /tmp/test-results/results.xml
