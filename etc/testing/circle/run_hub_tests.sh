#!/bin/bash

set -euxo pipefail

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

# Print client and server versions, for debugging.
pachctl version

# Run load tests.
set +e
PFS_RESPONSE_SPEC=$(pachctl run pfs-load-test "${@}")

if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi

set +u
DURATION=$(echo "$PFS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": ${DURATION:='default_load'},"type": "PFS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"}
EOF
set -u

PPS_RESPONSE_SPEC=$(pachctl run pps-load-test "${@}")

set +u
DURATION=$(echo "$PPS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": ${DURATION:='default_load'},"type": "PPS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"} 
EOF
set -u

if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi
set -e
pachctl debug dump /tmp/debug-dump
