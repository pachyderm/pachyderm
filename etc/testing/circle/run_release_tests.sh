#!/bin/bash

set -euxo pipefail

tar -xvf ./dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp && sudo cp /tmp/pachctl_*/pachctl /usr/local/bin

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

# provision a pulumi test env
curl -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
 -F name=release-"${CIRCLE_SHA1:0:7}" -F pachdVersion="${CIRCLE_SHA1}" -F valuesYaml=@etc/testing/circle/helm-values.yaml \
  https://helium.pachyderm.io/v1/api/workspace

# wait for helium to kick off to pulumi before pinging it.
sleep 5

READY=false
for _ in $(seq 108); do
  STATUS=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/release-${CIRCLE_SHA1:0:7}" | jq .Workspace.Status | tr -d '"')
  if [[ ${STATUS} == "ready" ]]
  then
    echo "success"
    READY=true
    break
  fi
  echo 'sleeping'
  sleep 10
done

if READY == false
then
  echo "failed to provision pulumi test env"
  exit 1
fi

pachdIp=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/release-${CIRCLE_SHA1:0:7}"  | jq .Workspace.PachdIp)

echo "{\"pachd_address\": ${pachdIp}, \"source\": 2}" | tr -d \\ | pachctl config set context "release-${CIRCLE_SHA1:0:7}" --overwrite && pachctl config set active-context "commit-${CIRCLE_SHA1:0:7}"

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