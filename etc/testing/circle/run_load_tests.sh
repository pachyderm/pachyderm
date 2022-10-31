#!/bin/bash

set -euxo pipefail

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

go version

# Install goreleaser.
GORELEASER_VERSION=0.169.0
curl -L "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz" \
    | tar xzf - -C "${HOME}/go/bin" goreleaser

# Build pachctl.
make install
pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION


JOB=$(echo "$CIRCLE_JOB" | tr '[:upper:]' '[:lower:]')

# # provision a pulumi load test env
curl --fail -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
 -F name=commit-"${CIRCLE_SHA1:0:7}-${JOB}" -F pachdVersion="${CIRCLE_SHA1}" -F backend=gcp_namespace_only -F clusterStack="pachyderm/helium/nightly-cluster" \
  -F disableNotebooks="True" -F valuesYaml=@etc/testing/circle/helm-load-env-values.yaml \
  https://helium.pachyderm.io/v1/api/workspace

# wait for helium to kick off to pulumi before pinging it.
sleep 5

for _ in $(seq 108); do
  STATUS=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/commit-${CIRCLE_SHA1:0:7}-${JOB}" | jq .Workspace.Status | tr -d '"')
  if [[ ${STATUS} == "ready" ]]
  then
    echo "success"
    break
  fi
  echo 'sleeping'
  sleep 10
done

pachdIp=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/commit-${CIRCLE_SHA1:0:7}-${JOB}"  | jq .Workspace.PachdIp)

echo "{\"pachd_address\": ${pachdIp}, \"source\": 2}" | tr -d \\ | pachctl config set context "commit-${CIRCLE_SHA1:0:7}-${JOB}" --overwrite && pachctl config set active-context "commit-${CIRCLE_SHA1:0:7}-${JOB}"

echo "${HELIUM_PACHCTL_AUTH_TOKEN}" | pachctl auth use-auth-token

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

if [[ "$CIRCLE_JOB" == *"nightly-load"* ]]; then

DURATION=$(echo "$PFS_RESPONSE_SPEC" | jq '.duration')
DURATION=${DURATION: 1:-2}
bq insert --ignore_unknown_values insights.load-tests << EOF
  {"gitBranch": "$CIRCLE_BRANCH","specName": "$BUCKET","duration": $DURATION,"type": "PFS", "commit": "$CIRCLE_SHA1", "timeStamp": "$(date +%Y-%m-%dT%H:%M:%S)"}
EOF

fi

set -e
pachctl debug dump /tmp/debug-dump
