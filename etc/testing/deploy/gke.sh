#!/bin/bash

# This creates a pachyderm cluster in AWS for testing. This is a thin wrapper around
# etc/deploy/aws.sh, but it uses the same state store bucket for all tests, so that
# kops clusters created for testing can always be enumerated and deleted.

## Parse command-line flags
if [[ "$#" -lt 1 ]]; then
  echo "Must pass at least one argument (command) to testing/deploy/aws.sh"
  exit 1
fi

PROJECT="pach-travis"
MACHINE_TYPE="n1-standard-4"
GCP_ZONE="us-west1-a"
NUM_NODES=3
STORAGE_SIZE=10

PREFIX=pachyderm-batch-test
gcloud config set compute/zone ${GCP_ZONE}

# Process args
new_opt="$( getopt --long="create,delete:,delete-all" -- "${0}" "${@}" )"
[[ "$?" -eq 0 ]] || exit 1
eval "set -- ${new_opt}"

KEY_FILE="$(dirname "${0}")/../pach-travis-86b9f180aa16.json"
set -x
gcloud auth activate-service-account --key-file="${KEY_FILE}"
gcloud config set project "${PROJECT}"
set +x

case "${1}" in
  --delete-all)
    set -x
    gcloud container clusters list | grep ${PREFIX} | awk '{print $1}' \
        | while read -r c; do
            ID=${c#${PREFIX}-}
            ID=${ID%-cluster}
            CLUSTER_NAME=${PREFIX}-${ID}-cluster
            STORAGE_NAME=${PREFIX}-${ID}-disk
            BUCKET_NAME=${PREFIX}-${ID}-bucket
            echo "Y" | gcloud container clusters delete "${CLUSTER_NAME}"
            echo "Y" | gcloud compute disks delete "${STORAGE_NAME}"
            gsutil rb "gs://${BUCKET_NAME}"
        done
    gcloud compute disks list | grep ${PREFIX} | awk '{print $1}' \
      | while read -r d; do
          echo "y" | gcloud compute disks delete "${d}"
        done
    gsutil ls | grep ${PREFIX} \
      | while read -r b; do
          gsutil rb "${b}";
        done
    set +x
    ;;
  --delete)
    set -x
    ID="${2}"
    CLUSTER_NAME=${PREFIX}-${ID}-cluster
    STORAGE_NAME=${PREFIX}-${ID}-disk
    BUCKET_NAME=${PREFIX}-${ID}-bucket
    echo "Y" | gcloud container clusters delete "${CLUSTER_NAME}"
    echo "Y" | gcloud compute disks delete "${STORAGE_NAME}"
    gsutil rb "gs://${BUCKET_NAME}"
    set +x
    ;;
  --create)
    set -x
    ID=$(uuid | cut -d- -f1)
    CLUSTER_NAME=${PREFIX}-${ID}-cluster
    STORAGE_NAME=${PREFIX}-${ID}-disk
    BUCKET_NAME=${PREFIX}-${ID}-bucket

    gcloud config set container/cluster "${CLUSTER_NAME}"
    gcloud container clusters create "${CLUSTER_NAME}" --scopes storage-rw --machine-type="${MACHINE_TYPE}" --num-nodes="${NUM_NODES}"
    gcloud compute disks create --size="${STORAGE_SIZE}GB" "${STORAGE_NAME}"
    gsutil mb "gs://${BUCKET_NAME}"
    pachctl deploy google "${BUCKET_NAME}" "${STORAGE_SIZE}" --static-etcd-volume="${STORAGE_NAME}"
    set +x
    echo "${ID}"
    ;;
esac

