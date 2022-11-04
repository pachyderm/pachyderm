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

# provision a aws pulumi test env
TIMESTAMP=$(date +%T | sed 's/://g')
WORKSPACE="${1}-${CIRCLE_SHA1:0:7}-${TIMESTAMP}"


if [ "${1}" = "wp" ]; then
  curl --fail -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
    -F name="${WORKSPACE}" -F pachdVersion="${CIRCLE_SHA1}" -F helmVersion="${CIRCLE_TAG:1}-${CIRCLE_SHA1}" -F backend="aws_cluster" \
    -F disableNotebooks="True" -F infraJson=@etc/testing/circle/workloads/aws-wp/infra.json -F valuesYaml=@etc/testing/circle/workloads/aws-wp/values.yaml \
https://helium.pachyderm.io/v1/api/workspace
elif [ "${1}" = "btl" ]; then
  curl --fail -X POST -H "Authorization: Bearer ${HELIUM_API_TOKEN}" \
    -F name="${WORKSPACE}" -F pachdVersion="${CIRCLE_SHA1}" -F helmVersion="${CIRCLE_TAG:1}-${CIRCLE_SHA1}" -F backend="aws_cluster" \
    -F disableNotebooks="True" -F infraJson=@etc/testing/circle/workloads/aws-btl/infra.json -F valuesYaml=@etc/testing/circle/workloads/aws-btl/values.yaml \
https://helium.pachyderm.io/v1/api/workspace
else
  echo "no valid customer name provided"
  exit 1
fi

# wait for helium to kick off to pulumi before pinging it.
sleep 5

for _ in $(seq 180); do
  STATUS=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/${WORKSPACE}" | jq .Workspace.Status | tr -d '"')
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

pachdIp=$(curl -s -H "Authorization: Bearer ${HELIUM_API_TOKEN}" "https://helium.pachyderm.io/v1/api/workspace/${WORKSPACE}"  | jq .Workspace.PachdIp)
withEnvoy=$(echo "$pachdIp" | sed 's/grpc:\/\//grpc:\/\/pachd-/g' | sed 's/30651/30650/g')
echo "${withEnvoy}"
pachctlCtx=$(echo "{\"pachd_address\": ${withEnvoy}, \"source\": 2}" | tr -d \\)
echo "${pachctlCtx}"
echo "${pachctlCtx}" | pachctl config set context "${WORKSPACE}" --overwrite && pachctl config set active-context "${WORKSPACE}"
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

if [ "${1}" = "wp" ]; then
  # cloning wp-workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  git checkout -b "bosterbuhr/wp-load-test"
  make wp-dag-test
  #make wp-destroy
elif [ "${1}" = "btl" ]; then
  # cloning battelle workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  git checkout -b "workload-hackathon-22"
  cd customer-success/testing/performance/battelle/dag/scripts/
  ./dataload.sh
  ./deploy-all.sh
else
  echo "no valid customer name provided"
  exit 1
fi