#!/bin/bash

# This creates a pachyderm cluster in AWS for testing. This is a thin wrapper around
# etc/deploy/aws.sh, but it uses the same state store bucket for all tests, so that
# kops clusters created for testing can always be enumerated and deleted.

## Parse command-line flags
if [[ "$#" -lt 1 ]]; then
  echo "Must pass --create, --delete, --delete-all or --list to testing/deploy/aws.sh"
  exit 1
fi

REGION=us-west-1
ZONE=us-west-1b
STATE_STORE=s3://pachyderm-travis-state-store-v1

# Process args
new_opt="$( getopt --long="create,delete:,delete-all,list" -- ${0} "${@}" )"
[[ "$?" -eq 0 ]] || exit 1
eval "set -- ${new_opt}"

# No need to authenticate, as auth creds are already in environment variables
# in travis

case "${1}" in
  --delete-all)
    set -x
    kops --state=${STATE_STORE} get clusters | tail -n+2 | awk '{print $1}' \
      | while read name; do
          kops --state=${STATE_STORE} delete cluster --name=${name} --yes
      done
    exit 0
    set +x
    ;;
  --list)
    kops --state=${STATE_STORE} get clusters
    ;;
  --delete)
    set -x
    NAME="${2}"
    kops --state=${STATE_STORE} delete cluster --name=${NAME} --yes
    set +x
    ;;
  --create)
    set -x
    deploy_script="$(realpath "$(dirname "${0}")/../../deploy/aws.sh")"
    sudo "${deploy_script}" --region=${REGION} --zone=${ZONE} --state=${STATE_STORE} --no-metrics
    set +x
    ;;
esac

