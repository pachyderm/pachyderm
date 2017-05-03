#!/bin/bash

set -euxo pipefail

which uuid

RESOURCE_GROUP=pachyderm-resource-group-$(uuid | cut -f 1 -d-)
LOCATION=westus

deploy_k8s() {
    az group create --name=$RESOURCE_GROUP --location=$LOCATION
    DNS_PREFIX=$(uuid | cut -f 1 -d-)
    CLUSTER_NAME=$(uuid | cut -f 1 -d-)-pachyderm-cluster-name
    az acs create --orchestrator-type=kubernetes --resource-group $RESOURCE_GROUP --name=$CLUSTER_NAME --dns-prefix=$DNS_PREFIX --generate-ssh-keys
    
    az acs kubernetes get-credentials --resource-group=$RESOURCE_GROUP --name=$CLUSTER_NAME
}

deploy_p7m() {
    # Needs to be globally unique across the entire Azure location
    STORAGE_ACCOUNT=[The name of the storage account where your data will be stored]
    
    CONTAINER_NAME=[The name of the Azure blob container where your data will be stored]
    
    # Needs to end in a ".vhd" extension
    STORAGE_NAME=pach-disk.vhd
    
    # We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
    # should work for 1000 commits on 1000 files.
    STORAGE_SIZE=100
    
    # Create a resource group
    az group create --name=${RESOURCE_GROUP} --location=${LOCATION}
    
    # Create azure storage account
    az storage account create \
      --resource-group="${RESOURCE_GROUP}" \
      --location="${LOCATION}" \
      --sku=Standard_LRS \
      --name="${STORAGE_ACCOUNT}" \
      --kind=Storage
    
    # Build microsoft tool for creating Azure VMs from an image
    STORAGE_KEY="$(az storage account keys list \
                     --account-name="${STORAGE_ACCOUNT}" \
                     --resource-group="${RESOURCE_GROUP}" \
                     --output=json \
                     | jq .[0].value -r
                  )"
    make docker-build-microsoft-vhd 
    VOLUME_URI="$(docker run -it microsoft_vhd \
                    "${STORAGE_ACCOUNT}" \
                    "${STORAGE_KEY}" \
                    "${CONTAINER_NAME}" \
                    "${STORAGE_NAME}" \
                    "${STORAGE_SIZE}G"
                 )"
    
    az storage account list | jq '.[].name'
    az storage blob list \
      --container=${CONTAINER_NAME} \
      --account-name=${STORAGE_ACCOUNT} \
      --account-key=${STORAGE_KEY}
    pachctl deploy microsoft ${CONTAINER_NAME} ${STORAGE_ACCOUNT} ${STORAGE_KEY} ${STORAGE_SIZE} --static-etcd-volume=${VOLUME_URI}
}


deploy_k8s
#deploy_p7m
