#!/bin/bash

set -euxo pipefail

pachctl config set context pachops --overwrite <<EOF
{"pachd_address":"grpcs://pachyderm.pachops.com:443", "session_token":"$PACHOPS_PACHYDERM_ROBOT_TOKEN"}
EOF

pachctl config set active-context pachops

pachctl version

pachctl put file -r pach-core-ci-results@master:/pachyderm/$(echo $CIRCLE_BRANCH | sed 's/\//_/g')/$CIRCLE_SHA1/$CIRCLE_JOB/$CIRCLE_WORKFLOW_JOB_ID/ -f /tmp/test-results
