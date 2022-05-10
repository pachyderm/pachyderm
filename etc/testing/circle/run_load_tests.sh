#!/bin/bash

set -euxo pipefail

mkdir -p "${HOME}/go/bin"
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GOPATH="${HOME}/go"

# Install go.
sudo rm -rf /usr/local/go
curl -L https://golang.org/dl/go1.17.3.linux-amd64.tar.gz | sudo tar xzf - -C /usr/local/
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

# Build and push docker images.
make docker-build
make docker-push

# pwd
# ls -lah
# cat helm-values.yaml

# # provision a pulumi load test env
# curl -X POST -H "Authorization: Bearer exvTH4eXVGh3FDTtHZ3wzTnF" \
#  -F name=load-test-CI1 -F pachdVersion=${VERSION} -F valuesYaml=@etc/testing/circle/helm-values.yaml \
#   https://0f52-172-98-132-18.ngrok.io/v1/api/workspace

for _ in $(seq 36); do
  STATUS=$(curl -s -H "Authorization: Bearer exvTH4eXVGh3FDTtHZ3wzTnF" https://0f52-172-98-132-18.ngrok.io/v1/api/workspace/sean-named-this-110 | jq .Workspace.Status | tr -d '"')
  if [[ ${STATUS} == "ready" ]]
  then
    echo "success"
    break
  fi
  echo 'sleeping'
  sleep 10
done

pachdIp=$(curl -s -H "Authorization: Bearer exvTH4eXVGh3FDTtHZ3wzTnF" https://0f52-172-98-132-18.ngrok.io/v1/api/workspace/sean-named-this-110 | jq .Workspace.PachdIp)

echo '{"pachd_address": ${pachdIp}, "source": 2}' | pachctl config set context sean-named-this-110 --overwrite && pachctl config set active-context sean-named-this-110

# Print client and server versions, for debugging.
pachctl version

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
