#!/bin/bash

# This creates a pachyderm cluster in AWS for testing. This is a thin wrapper around
# etc/deploy/aws.sh, but it uses the same state store bucket for all tests, so that
# kops clusters created for testing can always be enumerated and deleted.

## Parse command-line flags

STATE_STORE=s3://pachyderm-travis-state-store-v1
REGION=-
ZONE=us-west-1b
OP=-
CLOUDFRONT=

# Process args
new_opt="$( getopt --long="create,delete:,delete-all,list,zone:,use-cloudfront" -- ${0} "${@}" )"
[[ "$?" -eq 0 ]] || exit 1
eval "set -- ${new_opt}"

while true; do
  case "${1}" in
    --delete-all)
      OP=delete-all
      shift
      ;;
    --list)
      kops --state=${STATE_STORE} get clusters
      exit 0  # Shortcut
      ;;
    --delete)
      OP=delete
      NAME="${2}"
      shift 2
      ;;
    --create)
      OP=create
      shift
      ;;
    --zone)
      ZONE="${2}"
      shift 2
      ;;
    --use-cloudfront)
      # Default is not to provide the flag
      CLOUDFRONT="--use-cloudfront"
      shift
      ;;
    --)
      shift
      break
      ;;
  esac
done

len_zone_minus_one="$(( ${#ZONE} - 1 ))"
REGION=${ZONE:0:${len_zone_minus_one}}

echo -e "Region: ${REGION}\nZone: ${ZONE}"

# No need to authenticate with kops, as auth creds are already in environment variables
# in travis
set -x
case "${OP}" in
  create)
    aws_sh="$(realpath "$(dirname "${0}")/../../deploy/aws.sh")"
    cmd=("${aws_sh}" --region=${REGION} --zone=${ZONE} --state=${STATE_STORE} --no-metrics)
    if [[ -n "${CLOUDFRONT}" ]]; then
      cmd+=("${CLOUDFRONT}")
    fi
    sudo "${cmd[@]}"
    ;;
  delete)
    kops --state=${STATE_STORE} delete cluster --name=${NAME} --yes
    ;;
  delete-all)
    kops --state=${STATE_STORE} get clusters | tail -n+2 | awk '{print $1}' \
      | while read name; do
          kops --state=${STATE_STORE} delete cluster --name=${name} --yes
      done
    ;;
  *)
    set +x
    echo "Must pass --create, --delete, --delete-all or --list to testing/deploy/aws.sh"
    exit 1
esac

set +x
