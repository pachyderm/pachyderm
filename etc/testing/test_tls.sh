#!/bin/bash

set -euo pipefail

which match || {
  here="$(dirname "${0}")"
  GO111MODULE=on go install -mod=vendor -v "${here}/../../src/testing/match"
}

address=$(pachctl config get context `pachctl config get active-context` | jq -r .pachd_address)
if [[ "${address}" = "null" ]]; then
  echo "pachd_address must be set on the active context"
  exit 1
fi
hostport=$(echo $address | sed -e 's/grpcs:\/\///g' -e 's/grpc:\/\///g')

set -x

# Generate self-signed cert and private key
etc/deploy/gen_pachd_tls.sh $hostport ""

# Restart pachyderm with the given certs
etc/deploy/restart_with_tls.sh $hostport ${PWD}/pachd.pem ${PWD}/pachd.key

set +x # Do not log our activation code when running this script in Travis
pachctl enterprise activate "$(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)" && echo
set -x

# Make sure the pachyderm client can connect, write data, and create pipelines
go test -v -count=1 ./src/server -run TestSimplePipeline

# Make sure that config's pachd_address isn't disfigured by pachctl cmds that
# modify the pachctl config (bug fix)
echo admin | pachctl auth activate
otp="$(pachctl auth get-otp admin)"
echo "${otp}" | pachctl auth login --one-time-password
pachctl auth whoami | match admin # will fail if pachctl can't connect
echo yes | pachctl auth deactivate

# Undeploy TLS
yes | pachctl undeploy || true
pachctl deploy local -d
