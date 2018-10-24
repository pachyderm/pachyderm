#!/bin/bash

eval "set -- $( getopt -l "key:,cert:" "--" "${0}" "${@}" )"
while true; do
  case "${1}" in
    --cert)
      export PACH_CA_CERTS="${2}"
      shift 2
      ;;
    --key)
      PACH_TLS_KEY="${2}"
      shift 2
      ;;
    --)
      shift
      break
      ;;
  esac
done

# Turn down old pachd deployment, so that new (TLS-enabled) pachd doesn't try to connect to old, non-TLS pods
# I'm not sure why this is necessary -- pachd should communicate with itself via an unencrypted, internal port
# Empirically, though, the new pachd pod crashloops if I don't do this (2018/6/22)
kubectl get deploy/pachd -o json | jq '.spec.replicas = 0' | kubectl apply -f -

# Re-deploy pachd with new mount containing TLS key
pachctl deploy local -d --tls="${PACH_CA_CERTS},${PACH_TLS_KEY}" --dry-run | kubectl apply -f -

echo "######################################"
echo -e "Run:\nexport PACH_CA_CERTS=${PACH_CA_CERTS}\nto talk to the new tls-enabled pachd cluster"
echo "######################################"
# Wait for new pachd pod to start


echo "Waiting for old pachd to go down..."
WHEEL="\|/-"
retries=5
while pachctl version &>/dev/null && (( retries-- > 0 )); do
  echo -en "\e[G${WHEEL::1} (retries: ${retries})"
  WHEEL="${WHEEL:1}${WHEEL::1}"
  sleep 1
done
echo

echo "Waiting for new pachd to come up..."
retries=10
until pachctl version &>/dev/null || (( retries-- == 0 )); do
  echo -en "\e[G${WHEEL::1} (retries: ${retries})"
  WHEEL="${WHEEL:1}${WHEEL::1}"
  sleep 1
done
echo

# Delete old replicaset with no replicas (which kubernetes doesn't for some reason)
set -x
old_rs=$(kubectl get rs -l app=pachd,suite=pachyderm -o json | jq -r '.items[] | select(.spec.replicas == 0) | .metadata.name')
echo "old replicaset: ${old_rs}"
if [[ -n "${old_rs}" ]]; then
  kubectl delete rs/${old_rs}
fi
