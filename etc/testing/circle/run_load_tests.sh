#!/bin/bash

set -euxo pipefail

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

go version

# Build pachctl.
make install
pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

echo "{\"pachd_address\": \"grpc://${PACHD_IP}:80\", \"session_token\": \"test\"}" | tr -d \\ | pachctl config set context "test" --overwrite
pachctl config set active-context "test"

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

# shellcheck disable=SC2086
# shellcheck disable=SC2046
if [ "${?}" -ne 0 ] || [ $(jq 'has("error")' <<< $PFS_RESPONSE_SPEC) == true ]; then
	pachctl debug dump /tmp/debug-dump
	exit 1
fi

if [[ "$CIRCLE_JOB" == *"nightly-load"* ]]; then

DURATION=$(echo "$PFS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": $DURATION,"type": "PFS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"}
EOF

fi

pachctl delete pipeline --all
pachctl delete repo --all

# give 5 mins for gc to get ahead before next load test is run.
sleep 300

set -e
pachctl debug dump /tmp/debug-dump
