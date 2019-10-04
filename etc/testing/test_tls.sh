#!/bin/bash

set -euo pipefail

pachctl version
# Don't log our activation code when running this script in Travis
pachctl enterprise activate "$(aws s3 cp s3://pachyderm-engineering/test_enterprise_activation_code.txt -)" && echo

set -x

# Split PACHD_ADDRESS into host and port
IFS=: read -a addr_parts <<<"${PACHD_ADDRESS}"
unset PACHD_ADDRESS # use config value set by gen_pachd_tls.sh

# Generate certs and store a new address and trusted CA in the pachyderm config
etc/deploy/gen_pachd_tls.sh --ip="${addr_parts[0]}" --port="${addr_parts[1]}"

# Restart pachyderm with the given certs
etc/deploy/restart_with_tls.sh --key=${PWD}/pachd.key --cert=${PWD}/pachd.pem

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
pachctl deploy local -d
