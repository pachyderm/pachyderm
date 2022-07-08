#!/bin/bash

set -euxo pipefail

pachctl config set context pachops --overwrite <<EOF
{"pachd_address":"grpcs://pachyderm.pachops.com:443", "session_token":"$PACHOPS_PACHYDERM_ROBOT_TOKEN"}
EOF

pachctl version

pachctl put file -r build-results@master:/pachyderm/$CIRCLE_BRANCH/$CIRCLE_BUILD_NUM/$CIRCLE_JOB/$CIRCLE_WORKFLOW_JOB_ID/ -f /tmp/test-results
