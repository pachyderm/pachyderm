#!/bin/bash

set -euxo pipefail

# shellcheck disable=SC1090
source "$(dirname "$0")/env.sh"

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

helm install pachyderm etc/helm/pachyderm -f etc/testing/circle/helm-values.yaml

kubectl wait --for=condition=ready pod -l app=pachd --timeout=5m

# Print client and server versions, for debugging.  (Also waits for proxy to discover pachd, etc.)
for i in $(seq 1 20); do
    if pachctl version; then
        echo "pachd ready after $i attempts"
        break
    else
        sleep 5
        continue
    fi
done

# Run load tests.
set +e
PFS_RESPONSE_SPEC=$(pachctl run pfs-load-test "${@}")

if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi

if [[ "$CIRCLE_JOB" == *"nightly_load"* ]]; then

DURATION=$(echo "$PFS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": $DURATION,"type": "PFS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"}
EOF

fi

PPS_RESPONSE_SPEC=$(pachctl run pps-load-test "${@}")

if [[ "$CIRCLE_JOB" == *"nightly_load"* ]]; then

DURATION=$(echo "$PPS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": $DURATION,"type": "PPS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"}
EOF

fi

if [ "${?}" -ne 0 ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi
set -e
pachctl debug dump /tmp/debug-dump
