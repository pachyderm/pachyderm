#!/bin/bash

set -euxo pipefail

tar -xvzf ./dist-pach/pachctl/pachctl_*_linux_amd64.tar.gz -C /tmp

sudo mv /tmp/pachctl_*/pachctl /usr/local/bin && chmod +x /usr/local/bin/pachctl

pachctl version --client-only

# Set version for docker builds.
VERSION="$(pachctl version --client-only)"
export VERSION

if [ "${1}" = "wp" ]; then
  # cloning wp-workload test repo
  git clone https://github.com/pachyderm/customer-success.git customer-success
  git checkout -b "bosterbuhr/wp-load-test"
  make wp-dag-test
  pachctl create pipeline -f ~/project/etc/testing/circle/workloads/aws-wp/metrics.json
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