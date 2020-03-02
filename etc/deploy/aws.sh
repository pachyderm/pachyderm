#!/bin/bash

# Argument Defaults
DEPLOY_PACHD="true"  # By default, aws.sh deploys pachyderm in its k8s cluster
USE_CLOUDFRONT="false"

# Other defaults
CLOUDFRONT_DOMAIN=

set -euxo pipefail

parse_flags() {
  # Check prereqs
  command -v aws
  command -v jq
  command -v uuid
  # Common config
  export AWS_REGION=us-west-1
  export AWS_AVAILABILITY_ZONE=us-west-1a
  export KOPS_BUCKET=s3://k8scom-state-store-pachyderm-${RANDOM}
  local use_existing_kops_bucket='false'

  # Parse flags
  eval "set -- $( getopt -l "state:,region:,zone:,no-metrics,use-cloudfront,no-pachyderm" "--" "${0}" "${@:-}" )"
  while true; do
      case "${1}" in
          --state)
            export KOPS_BUCKET="${2}"
            use_existing_kops_bucket='true'
            shift 2
            ;;
          --zone)
            export AWS_AVAILABILITY_ZONE="${2}"
            local len_zone_minus_one="$(( ${#AWS_AVAILABILITY_ZONE} - 1 ))"
            export AWS_REGION=${AWS_AVAILABILITY_ZONE:0:${len_zone_minus_one}}
            shift 2
            ;;
          --no-pachyderm)
            DEPLOY_PACHD="false" # default is true, see top of file
            shift
            ;;
          --use-cloudfront)
            USE_CLOUDFRONT="true" # default is false, see top of file
            shift
            ;;
          --)
            shift
            break
            ;;
          *)
            echo "Unrecognized argument: \"${1}\""
            echo "Must be one of --state, --zone, --use-cloudfront, --no-pachyderm"
            exit 1
            ;;
      esac
  done

  echo "Availability zone: ${AWS_AVAILABILITY_ZONE}"
  if [[ ! ( "${KOPS_BUCKET}" =~ s3://* ) ]]; then
    echo "kops state bucket must start with \"s3://\" but is \"${KOPS_BUCKET}\""
    echo "Exiting to be safe..."
    exit 1
  fi

  if [ "${use_existing_kops_bucket}" == 'false' ]; then
    create_s3_bucket "${KOPS_BUCKET}" false || exit 1
  fi
}

# Takes 2 args
# $1 : bucket name (required)
# $2 : boolean to use cloudfront or not (required)
create_s3_bucket() {
  if [[ "$#" -lt 1 ]] || [[ "${1}" =~ s3://* ]] ; then
    echo "Error: create_s3_bucket needs a bucket name"
    return 1
  fi
  if [[ "$#" -lt 2 ]]; then
    echo "Error: must specify whether cloudfront has access"
    return 1
  fi
  if [[ "$#" -eq 3 ]]; then
    KOPS_STATE_BUCKET="${3}"
  fi
  BUCKET="${1#s3://}"
  GIVE_CLOUDFRONT_ACCESS="${2}"

  # For some weird reason, s3 emits an error if you pass a location constraint when location is "us-east-1"
  if [[ "${AWS_REGION}" == "us-east-1" ]]; then
    aws s3api create-bucket --bucket "${BUCKET}" --region "${AWS_REGION}"
  else
    aws s3api create-bucket --bucket "${BUCKET}" --region "${AWS_REGION}" --create-bucket-configuration "LocationConstraint=${AWS_REGION}"
  fi
  if [[ -n "${KOPS_STATE_BUCKET}" ]]; then
    # Put bucket name in kops state store bucket right away, in case the
    # cluster doesn't finish coming up
    #
    # TODO(msteffen) pass $NAME in as an argument--relying on the fact that it
    # was already exported by deploy_k8s_on_aws is brittle and hard to read
    jq -n \
      --arg pachyderm_bucket "${BUCKET}" \
      --arg timestamp "$(date)" \
      '{"pachyderm_bucket": $pachyderm_bucket, "created": $timestamp}' \
    | aws s3 cp - "${KOPS_STATE_BUCKET}/${NAME}-info.json"
  fi

  if [ "${GIVE_CLOUDFRONT_ACCESS}" != "false" ]; then
    mkdir -p tmp
    sed "s/BUCKET_NAME/$BUCKET/" etc/deploy/cloudfront/bucket-policy.json.template > tmp/bucket-policy.json
    aws s3api put-bucket-policy --bucket "$BUCKET" --policy file://tmp/bucket-policy.json --region="${AWS_REGION}"
    create_cloudfront_distribution "${BUCKET}" || exit 1
  fi
}

create_cloudfront_distribution() {
  if [[ "$#" -lt 1 ]]; then
    echo "Error: create_cloudfront_distribution needs a bucket name"
    return 1
  fi
  BUCKET="${1#s3://}"

  someuuid=$(uuid | cut -f 1 -d-)
  mkdir -p tmp
  sed "s/XXCallerReferenceXX/$someuuid/" etc/deploy/cloudfront/distribution.json.template > tmp/cloudfront-distribution.json
  sed -i "s/XXBucketNameXX/$BUCKET/" tmp/cloudfront-distribution.json

  aws cloudfront create-distribution --distribution-config file://tmp/cloudfront-distribution.json > tmp/cloudfront-distribution-info.json
  CLOUDFRONT_ID=$(jq -r ".Distribution.Id" < tmp/cloudfront-distribution-info.json)
  export CLOUDFRONT_ID
  CLOUDFRONT_DOMAIN=$(jq -r ".Distribution.DomainName" < tmp/cloudfront-distribution-info.json | cut -f 1 -d .)
  aws cloudfront wait distribution-deployed --id "$CLOUDFRONT_ID"
}

deploy_k8s_on_aws() {
    # Verify authorization
    aws configure list
    aws iam list-users

    NODE_SIZE=r4.xlarge
    export NODE_SIZE
    MASTER_SIZE=r4.xlarge
    export MASTER_SIZE
    NUM_NODES=2
    export NUM_NODES
    NAME=$(uuid | cut -f 1 -d-)-pachydermcluster.kubernetes.com
    export NAME
    echo "kops state store: ${KOPS_BUCKET}"
    local kops_flags=(
        "--state=${KOPS_BUCKET}"
        "--cloud=aws"
        "--zones=${AWS_AVAILABILITY_ZONE}"
        "--node-count=${NUM_NODES}"
        "--master-zones=${AWS_AVAILABILITY_ZONE}"
        "--dns=private"
        "--dns-zone=kubernetes.com"
        "--node-size=${NODE_SIZE}"
        "--master-size=${MASTER_SIZE}"
        "--name=${NAME}"
        "--kubernetes-version=1.10.0"
        --yes
    )
    if [[ -n "${KUBECONFIG:-}" ]]; then
      kops_flags+=("--config=${KUBECONFIG}")
    fi
    kops create cluster "${kops_flags[@]}"
    kops update cluster "${NAME}" --yes --state="${KOPS_BUCKET}"

    # Record state store bucket in temp file.
    # This will allow us to cleanup the cluster afterwards
    set +euxo pipefail
    mkdir tmp
    echo "KOPS_STATE_STORE=${KOPS_BUCKET}" >> "tmp/${NAME}.sh"
    echo "${NAME}" > tmp/current-benchmark-cluster.txt
    set -euxo pipefail

    wait_for_k8s_master_ip
    update_sec_group
    wait_for_nodes_to_come_online
    remove_default_limit
}

update_sec_group() {
    SECURITY_GROUP_ID="$(
        aws ec2 describe-instances --filters "Name=instance-type,Values=${MASTER_SIZE}" --region "${AWS_REGION}" --output=json \
          | jq --raw-output ".Reservations[].Instances[] | select([.Tags[]?.Value | contains(\"masters.${NAME}\")] | any) | .SecurityGroups[0].GroupId"
    )"
    export SECURITY_GROUP_ID
    # For k8s access
    aws ec2 authorize-security-group-ingress --group-id "${SECURITY_GROUP_ID}" --protocol tcp --port 8080 --cidr "0.0.0.0/0" --region "${AWS_REGION}"
    # For pachyderm direct access:
    aws ec2 authorize-security-group-ingress --group-id "${SECURITY_GROUP_ID}" --protocol tcp --port 30650 --cidr "0.0.0.0/0" --region "${AWS_REGION}"
}

remove_default_limit() {
  # Kops, by default, creates a LimitRange that applies a 100m CPU Request to
  # all pods in the "default" namespace. Because we don't turn down most of the
  # pipelines spawned by our tests, this default request prevents our test
  # suite from finishing, as nodes fill up and later tests can't schedule
  # pipelines. Therefore we remove the LimitRange so all pipelines have no
  # resource request by default.
  kubectl delete --namespace=default limits/limits
}

# Prints a spinning wheel. Every time you call it, the wheel advances 1/4 turn
WHEEL="-\|/"
spin() {
    echo -en "\e[D${WHEEL:0:1}"
    WHEEL=${WHEEL:1}${WHEEL:0:1}
}

wait_for_k8s_master_ip() {
    # Get the IP of the k8s master node and hack /etc/hosts so we can connect
    # Need to retry this in a loop until we see the instance appear
    set +euxo pipefail
    echo "Retrieving ec2 instance list to get k8s master domain name (may take a minute)"
    get_k8s_master_domain
    while [ $? -ne 0 ]; do
        spin
        sleep 1
        get_k8s_master_domain
    done
    echo "Master k8s node is up and lives at ${K8S_MASTER_DOMAIN}"
    set -euxo pipefail
    masterk8sip="$(dig +short "${K8S_MASTER_DOMAIN}")"
    # This is the only operation that requires sudo privileges
    echo " " | sudo tee -a /etc/hosts # Some files dont contain newlines ... I'm looking at you travisCI
    echo "${masterk8sip} api.${NAME}" | sudo tee -a /etc/hosts
    echo "state of /etc/hosts:"
    cat /etc/hosts
}

wait_for_nodes_to_come_online() {
    # Wait until all nodes show as ready, and we have as many as we expect
    set +euxo pipefail
    echo "Waiting for nodes to come online (may take a few minutes)"
    check_all_nodes_ready >/dev/null 2>&1
    while [ $? -ne 0 ]; do
        spin
        sleep 1
        check_all_nodes_ready >/dev/null 2>&1
    done
    set -euxo pipefail
    rm nodes.txt
}

check_all_nodes_ready() {
    echo "Checking k8s nodes are ready"
    kubectl get nodes > nodes.txt
    if [ $? -ne 0 ]; then
        return 1
    fi

    total_nodes=$((${NUM_NODES}+1))
    ready_nodes=$(grep -v NotReady < nodes.txt | grep -c Ready)
    echo "total ${total_nodes}, ready ${ready_nodes}"
    if [ "${ready_nodes}" == "${total_nodes}" ]; then
        echo "all nodes ready"
        return 0
    fi
    return 1
}

get_k8s_master_domain() {
    K8S_MASTER_DOMAIN="$(
        aws ec2 describe-instances --filters "Name=instance-type,Values=${MASTER_SIZE}" --region "${AWS_REGION}" --output=json \
          | jq --raw-output ".Reservations[].Instances[] | select([.Tags[]?.Value | contains(\"masters.${NAME}\")] | any) | .PublicDnsName"
    )"
    export K8S_MASTER_DOMAIN
    if [ -n "${K8S_MASTER_DOMAIN}" ]; then
        return 0
    fi
    return 1
}

check_kops_version() {
  command -v kops
  KOPS_VERSION="$( kops version | awk '{print $2}' )"
  echo "${KOPS_VERSION#1.} >= 8.0" | bc
  if [[ "$( echo "${KOPS_VERSION#1.} >= 8.0" | bc )" -ne 1 ]]; then
    set +x
    echo "Your kops version is too old--must have at least 1.8.0"
    exit 1
  fi
}

##################################
###### Deploy Pach cluster #######
##################################

deploy_pachyderm_on_aws() {
    echo "deploying pachyderm on AWS"
    # shared with k8s deploy script:
    export STORAGE_SIZE=100
    export PACHYDERM_BUCKET=${RANDOM}-pachyderm-store
    create_s3_bucket "${PACHYDERM_BUCKET}" ${USE_CLOUDFRONT} "${KOPS_BUCKET}"

    # Since my user should have the right access:
    AWS_ID=$(tr -s "[:space:]" < ~/.aws/credentials | grep -m1 ^aws_access_key_id  | cut -d " " -f 3)
    AWS_KEY=$(tr -s "[:space:]" < ~/.aws/credentials | grep -m1 ^aws_secret_access_key | cut -d " " -f 3)


    # Omit token since im using my personal creds
    cmd=( pachctl deploy amazon "${PACHYDERM_BUCKET}" "${AWS_REGION}" "${STORAGE_SIZE}" "--dynamic-etcd-nodes=1" --no-dashboard "--credentials=${AWS_ID},${AWS_KEY},")
    if [[ "${USE_CLOUDFRONT}" == "true" ]]; then
      cmd+=( "--cloudfront-distribution" "${CLOUDFRONT_DOMAIN}" )
    fi
    "${cmd[@]}"  # Run pachctl deploy
}

if [ "${EUID}" -ne 0 ]; then
  echo "Cowardly refusing to deploy cluster. Please run as root"
  echo "Please run this command like 'sudo -E make launch-bench'"
  exit 1
fi
parse_flags "${@:-}"

command -v pachctl
check_kops_version

deploy_k8s_on_aws
if [[ "${DEPLOY_PACHD}" == "true" ]]; then
  deploy_pachyderm_on_aws
fi

if [[ "${USE_CLOUDFRONT}" == "true" ]]; then
  echo "To upgrade cloudfront to use security credentials, e.g.:"
  echo ""
  echo "    $./etc/deploy/cloudfront/secure-cloudfront.sh --zone us-east-1b --bucket 2642-pachyderm-store --distribution E3DPJE36K8O9U7 --keypair-id APKAXXXXXXXXXX --private-key-file pk-APKXXXXXXXXXXXX.pem"
  echo ""
  echo "Please save this deploy output to a file for your future reference,"
  echo "You'll need some of the values reported here"
  # They'll need this ID to run the secure script
  echo "Created cloudfront distribution with ID: ${CLOUDFRONT_ID}"
fi
echo "Cluster created:"
echo "${NAME}"

# Put the cluster address in the pachyderm config
config_path="${HOME}/.pachyderm/config.json"
[[ -d "${HOME}/.pachyderm" ]] || mkdir "${HOME}/.pachyderm"
[[ -e "${config_path}" ]] || {
  echo '{}' >"${config_path}"
}
tmpfile="$(mktemp "$(pwd)/tmp.XXXXXXXXXX")"
jq --monochrome-output \
  ".v1.pachd_address=\"${K8S_MASTER_DOMAIN}:30650\"" \
  "${config_path}" \
  >"${tmpfile}"
mv "${tmpfile}" "${config_path}"
chmod 777 "${config_path}"

echo "Cluster address has been written to ${config_path}"
