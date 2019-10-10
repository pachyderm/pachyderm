#!/bin/bash

set -e

eval "set -- $( getopt -l "key:,cert:" "--" "${0}" "${@}" )"
while true; do
  case "${1}" in
    --cert)
      TLS_CERT="${2}"
      shift 2
      ;;
    --key)
      TLS_KEY="${2}"
      shift 2
      ;;
    --)
      shift
      break
      ;;
  esac
done

echo "Turning down any existing pachyderm cluster..."
if pachctl version --timeout=5s; then
  # Turn down old pachd deployment, so that new (TLS-enabled) pachd doesn't try
  # to connect to old, non-TLS pods I'm not sure why this is necessary -- pachd
  # should communicate with itself via an unencrypted, internal port.
  #
  # Empirically, though, new pachd pods crashloop if I don't do this (2018/6/22)
  echo yes | pachctl undeploy

  # Wait 30s for old pachd to go down
  echo "Waiting for old pachd to go down..."
  WHEEL='\|/-'
  echo "${WHEEL}"
  retries=15
  while pachctl version --timeout=5s &>/dev/null && (( retries-- > 0 )); do
    echo -en "\e[G\e[K ${WHEEL::1} (retries: ${retries})"
    WHEEL="${WHEEL:1}${WHEEL::1}"
    sleep 2
  done
  echo
fi

# Re-deploy pachd with new mount containing TLS key
pachctl deploy local -d --tls="${TLS_CERT},${TLS_KEY}" --dry-run | kubectl apply -f -

# Wait 30s for new pachd to come up
echo "Waiting for new pachd to come up..."
retries=15
until pachctl version --timeout=5s &>/dev/null || (( retries-- == 0 )); do
  echo -en "\e[G\e[K ${WHEEL::1} (retries: ${retries})"
  WHEEL="${WHEEL:1}${WHEEL::1}"
  sleep 2
done

echo ""
echo "Remember to configure pachctl to trust pachd's new cert!"
echo ""
