#!/bin/bash

set -euo pipefail

active_context=`pachctl config get active-context`
address=`pachctl config get context $active_context | jq -r .pachd_address`

if [[ "${address}" = "null" ]]; then
  echo "pachd_address must be set on the active context"
  exit 1
fi

IFS=':' read -ra parts <<< "$address"
host=${parts[0]}
port=${parts[1]:-30650}
echo "testing TLS against host=${host}, port=${port}"

# Validate the host (that it doesn't start with a protocol)
if [[ "${host}" =~ :// ]]; then
  echo "${host} should not start with <protocol>://" >/dev/stderr
  exit 1
fi

# Generate self-signed cert and private key
if [[ "${host}" =~ [0-9]+\.[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  etc/deploy/gen_pachd_tls.sh --ip="${host}" --port="${port}"
else
  etc/deploy/gen_pachd_tls.sh --dns="${host}" --port="${port}"
fi

# Restart pachyderm with the given certs
etc/deploy/restart_with_tls.sh --key=${PWD}/pachd.key --cert=${PWD}/pachd.pem

# Use new cert in pachctl
echo "Backing up Pachyderm config to \$HOME/.pachyderm/config.json.backup"
echo "New config with address and cert is at \$HOME/.pachyderm/config.json"
cp ~/.pachyderm/config.json ~/.pachyderm/config.json.backup
pachctl config update context \
  --pachd-address="grpcs://${host}:${port}" \
  --server-cas="$(cat ./pachd.pem | base64)"

set +x
# Don't log our activation code when running this script in Travis
pachctl enterprise activate "$(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)" && echo
set -x

# Make sure the pachyderm client can connect, write data, and create pipelines
go test -v ./src/server -run TestPipelineWithParallelism

# Make sure that config's pachd_address isn't disfigured by pachctl cmds (bug
# fix)
echo admin | pachctl auth activate
otp="$(pachctl auth get-otp admin)"
echo "${otp}" | pachctl auth login --one-time-password
pachctl auth whoami | grep -q admin # will fail if pachctl can't connect
echo yes | pachctl auth deactivate

# Undeploy TLS
yes | pachctl undeploy || true
pachctl deploy local -d
