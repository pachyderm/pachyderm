#!/bin/bash

# This creates a pachyderm cluster in AWS for testing. This is a thin wrapper around
# etc/deploy/aws.sh, but it uses the same state store bucket for all tests, so that
# kops clusters created for testing can always be enumerated and deleted.

set -euxo pipefail

command -v jq

delete_resources() {
  local name="${1}"
  if [[ ! -f ${HOME}/.pachyderm/${name}-info.json ]]; then
    # Try to copy cluster info from s3 to local file (may not exist if cluster didn't finish coming up)
    S3_CLUSTER_INFO="${KOPS_BUCKET}/${name}-info.json"
    if aws s3 ls "${S3_CLUSTER_INFO}"; then
      aws s3 cp "${S3_CLUSTER_INFO}" "${HOME}/.pachyderm/${name}-info.json"
    fi
  fi
  # Now that we might have fetched the file, try to delete the pachyderm bucket
  if [[ -f ${HOME}/.pachyderm/${name}-info.json ]]; then
    aws s3 rb \
      --region "${REGION}" \
      --force \
      "s3://$(jq --raw-output .pachyderm_bucket "${HOME}/.pachyderm/${name}-info.json")" \
      >/dev/null
    rm "${HOME}/.pachyderm/${name}-info.json"
  fi
  kops --state="${KOPS_BUCKET}" delete cluster --name="${name}" --yes
  aws s3 rm "${KOPS_BUCKET}/${name}-info.json"
}

ZONE="${ZONE:-us-west-1b}"
KOPS_BUCKET=${KOPS_BUCKET:-s3://pachyderm-travis-state-store-v1}
OP=-
CLOUDFRONT=
DEPLOY_PACHD="true"  # By default, aws.sh deploys pachyderm in its k8s cluster
len_zone_minus_one="$(( ${#ZONE} - 1 ))"
REGION=${ZONE:0:${len_zone_minus_one}}

# Process args
new_opt="$( getopt --long="create,delete:,delete-all,list,zone:,use-cloudfront,no-pachyderm" -- "${0}" "${@}" )"
eval "set -- ${new_opt}"

while true; do
  case "${1}" in
    --list)
      kops --state="${KOPS_BUCKET}" get clusters
      exit 0  # Shortcut
      ;;
    --delete)
      OP=delete
      CLUSTER_NAME="${2}"
      shift 2
      ;;
    --delete-all)
      OP=delete-all
      shift
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
    --no-pachyderm)
      DEPLOY_PACHD="false" # default is true, see top of file
      shift
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "Unrecognized argument: \"${1}\""
      echo "Must be one of --list, --delete=<cluster>, --delete-all, --create [--zone=<zone>] [--use-cloudfront] [--no-pachyderm]"
      exit 1
      ;;
  esac
done

echo -e "Zone: ${ZONE}"

# No need to authenticate with kops, as auth creds are already in environment variables
# in travis
set -x
case "${OP}" in
  create)
    pachctl config set metrics false
    aws_sh="$(dirname "${0}")/../../deploy/aws.sh"
    aws_sh="$(realpath "${aws_sh}")"
    cmd=("${aws_sh} --zone=\"${ZONE}\" --state=\"${KOPS_BUCKET}\"")
    if [[ "${DEPLOY_PACHD}" == "false" ]]; then
      cmd+=("--no-pachyderm")
    fi
    if [[ -n "${CLOUDFRONT}" ]]; then
      cmd+=("${CLOUDFRONT}")
    fi
    sudo env "PATH=${PATH}" "GOPATH=${GOPATH}" "KUBECONFIG=${KUBECONFIG}" "${cmd[@]}"
    check_ready="$(dirname "${0}")/../../kube/check_ready.sh"
    check_ready="$(realpath "${check_ready}")"
    sudo env "PATH=${PATH}" "GOPATH=${GOPATH}" "KUBECONFIG=${KUBECONFIG}" bash -c \
      "until timeout 1s ${check_ready} app=pachd; do sleep 1; echo -en \"\\033[F\"; done"
    ;;
  delete)
    delete_resources "${CLUSTER_NAME}"
    ;;
  delete-all)
    kops --state="${KOPS_BUCKET}" get clusters | tail -n+2 | awk '{print $1}' \
      | while read -r name; do
        delete_resources "${name}"
      done
    ;;
  *)
    set +x
    echo "Must pass --create, --delete, --delete-all or --list to testing/deploy/aws.sh"
    exit 1
esac

set +x
