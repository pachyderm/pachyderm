#!/bin/bash

set -euxo pipefail

tar -xvzf /tmp/workspace/dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp

sudo mv /tmp/pachctl_*/pachctl /usr/local/bin && chmod +x /usr/local/bin/pachctl

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

if [ "${1}" = "wp" ]; then
  ENV="qa2"
elif [ "${1}" = "btl" ]; then
  ENV="qa3"
elif [ "${1}" = "krs" ]; then
  ENV="qa4"
else
  echo "no valid customer name provided"
  exit 1
fi

echo "{\"pachd_address\": \"grpcs://${ENV}.workspace.pachyderm.com:443\", \"session_token\": \"test\"}" | tr -d \\ | pachctl config set context "test" --overwrite
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

if [ "${1}" = "wp" ]; then
  # cloning wp-workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  cd ci-load-tests/
  make wp-test
  pachctl create pipeline -f ~/project/etc/testing/circle/workloads/aws-wp/metrics.json
elif [ "${1}" = "btl" ]; then
  # cloning btl workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  git checkout -b "workload-hackathon-22"
  cd customer-success/testing/performance/battelle/dag/scripts/
  ./dataload.sh
  ./deploy-most.sh
  pachctl create pipeline -f ~/project/etc/testing/circle/workloads/aws-btl/metrics.json
elif [ "${1}" = "krs" ]; then
  # cloning krs workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  git checkout -b "workload-hackathon-22"
  cd ./testing/performance/karius/FD-0388/
  make pipeline
  pachctl create pipeline -f ~/project/etc/testing/circle/workloads/aws-krs/metrics.json
else
  echo "no valid customer name provided"
  exit 1
fi