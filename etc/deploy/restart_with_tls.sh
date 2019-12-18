#!/bin/bash

set -e

hostport=$1
cert=$2
key=$3

echo "Turning down any existing pachyderm cluster..."
if pachctl version --timeout=5s; then
  # Turn down old pachd deployment, so that new (TLS-enabled) pachd doesn't try
  # to connect to old, non-TLS pods I'm not sure why this is necessary -- pachd
  # should communicate with itself via an unencrypted, internal port.
  #
  # Empirically, though, new pachd pods crashloop if I don't do this (2018/6/22)
  echo yes | pachctl undeploy

  # Wait 30s for old pachd to go down
  echo -n "Waiting for old pachd to go down"
  retries=15
  while pachctl version --timeout=5s &>/dev/null && (( retries-- > 0 )); do
    echo -n .
    sleep 2
  done
  echo
fi

# Re-deploy pachd with new mount containing TLS key
pachctl deploy local -d --tls="$cert,$key" --dry-run | kubectl apply -f -

# Wait 30s for new pachd to come up
echo -n "Waiting for new pachd to come up"
retries=15
until pachctl version --timeout=5s &>/dev/null || (( retries-- == 0 )); do
  echo -n .
  sleep 2
done
echo
