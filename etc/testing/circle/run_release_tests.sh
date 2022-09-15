#!/bin/bash

set -euxo pipefail

tar -xvzf ./dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp

sudo mv /tmp/pachctl_*/pachctl /usr/local/bin && chmod +x /usr/local/bin/pachctl

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

# wait for 5mins artifact hub to sync helm chart.
# should be replaced with a ping 
sleep 300


# provision a pulumi test env
WORKSPACE=${CIRCLE_TAG//./-}

if [ "${1-default}" = "aws" ]; then
  curl -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
  -F name="${WORKSPACE}-${CIRCLE_SHA1:0:7}" -F pachdVersion="${CIRCLE_SHA1}" -F helmVersion="${CIRCLE_TAG:1}-${CIRCLE_SHA1}" \
  -F backend="aws_cluster" -F infraJson=@etc/testing/circle/workloads/aws-examples/infra.json -F valuesYaml=@etc/testing/circle/workloads/aws-examples/values.yaml \
    https://helium.pachyderm.io/v1/api/workspace
else
  # assume default gcp
  curl -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
  -F name="${WORKSPACE}-${CIRCLE_SHA1:0:7}" -F pachdVersion="${CIRCLE_SHA1}" -F helmVersion="${CIRCLE_TAG:1}-${CIRCLE_SHA1}" \
    https://helium.pachyderm.io/v1/api/workspace
fi


# wait for helium to kick off to pulumi before pinging it.
sleep 5

for _ in $(seq 180); do
  STATUS=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/${WORKSPACE}-${CIRCLE_SHA1:0:7}" | jq .Workspace.Status | tr -d '"')
  if [[ ${STATUS} == "ready" ]]
  then
    echo "success"
    break
  fi
  echo 'sleeping'
  sleep 10
done

if [[ ${STATUS} != "ready" ]]
then
  echo "failed to provision pulumi workspace"
  exit 1
fi

pachdIp=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/${WORKSPACE}-${CIRCLE_SHA1:0:7}"  | jq .Workspace.PachdIp)

if [ "${1-default}" = "aws" ]; then
  withEnvoy=$(echo "$pachdIp" | sed 's/grpc:\/\//grpc:\/\/pachd-/g' | sed 's/30651/30650/g')
  echo "${withEnvoy}"
  pachctlCtx=$(echo "{\"pachd_address\": ${withEnvoy}, \"source\": 2}" | tr -d \\)
  echo "${pachctlCtx}"
  echo "${pachctlCtx}" | pachctl config set context "${WORKSPACE}-${CIRCLE_SHA1:0:7}" --overwrite && pachctl config set active-context "${WORKSPACE}-${CIRCLE_SHA1:0:7}"
else
  # assume default gcp
  echo "{\"pachd_address\": ${pachdIp}, \"source\": 2}" | tr -d \\ | pachctl config set context "${WORKSPACE}-${CIRCLE_SHA1:0:7}" --overwrite && pachctl config set active-context "${WORKSPACE}-${CIRCLE_SHA1:0:7}"
fi

echo "${HELIUM_PACHCTL_AUTH_TOKEN}" | pachctl auth use-auth-token

# Print client and server versions, for debugging.  (Also waits for proxy to discover pachd, etc.)
READY=false
for i in $(seq 1 20); do
    if pachctl version; then
        echo "pachd ready after $i attempts"
        READY=true
        break
    else
        sleep 5
        continue
    fi
done

if [ "$READY" = false ] ; then
    echo 'pachd failed to start'
    exit 1
fi

pushd examples/opencv
    pachctl create repo images
    pachctl create pipeline -f edges.json
    pachctl create pipeline -f montage.json
    pachctl put file images@master -i images.txt
    pachctl put file images@master -i images2.txt

    # wait for everything to finish
    pachctl wait commit "montage@master"

    # ensure the montage image was generated
    pachctl inspect file montage@master:montage.png
popd

pachctl delete pipeline --all
pachctl delete repo --all