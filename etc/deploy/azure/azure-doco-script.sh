#!/bin/bash

set -xeou pipefail

# This script will create a file called ${NAME}.values.yaml in the current
# directory.
#
# This script will create an Azure cluster, a static ip, the PostgreSQL instance
# and databases, and the storage container. It will also install pachyderm into
# the cluster.  This script assumes you have created your own resource group.
#
# It may be called with the -e flag to create an empty cluster.

# PLEASE CHANGE THESE NEXT 3 VARIABLES AT A MINIMUM
RESOURCE_GROUP="a-resource-group"
NAME="diligent-elephant"
SQL_ADMIN_PASSWORD="correcthorsebatterystaple"

EMPTY_CLUSTER=false
print-help ()
{
    echo "$0: Create an Azure cluster" >&2
    echo
    echo "Options:"
    echo
    echo "-h  display this help text"
    echo "-e  create the cluster without installing Pachyderm"
    echo
    exit
}
while getopts he- OPT; do
    case "$OPT" in
	h )   set +x; print-help ;;
	e )   EMPTY_CLUSTER=true ;;
	??* ) close "Invalid option --$OPT" ;;
	? )   exit 2 ;;
    esac
done

# This group of variables can be changed, but are sane defaults
NODE_SIZE="Standard_DS4_v2"

# The following variables probably shouldn't be changed
CLUSTER_NAME="${NAME}-aks"
# Azure storage account names must be digits and lower-case letters, and must be
# globally unique.
STORAGE_ACCOUNT_NAME=$(echo $NAME | tr -d "-" | tr '[:upper:]' '[:lower:]')storage
CONTAINER_NAME="${NAME}-container"
SQL_INSTANCE_NAME="${NAME}-sql"
STATIC_IP_NAME="${NAME}-ip"


kubectl version --client=true

az version

helm version

jq --version


az aks create --name ${CLUSTER_NAME} \
   --resource-group ${RESOURCE_GROUP} \
   --node-vm-size ${NODE_SIZE}

az aks get-credentials --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME}

az storage account create --name "${STORAGE_ACCOUNT_NAME}" \
   --sku Premium_LRS \
   --kind BlockBlobStorage \
   --resource-group "${RESOURCE_GROUP}"

STORAGE_KEY="$(az storage account keys list \
   		             --account-name="${STORAGE_ACCOUNT_NAME}" \
              		     --resource-group="${RESOURCE_GROUP}" \
			     --output=json \
              | jq '.[0].value' -r
            )"

az storage container create --name "${CONTAINER_NAME}" \
   --account-name "${STORAGE_ACCOUNT_NAME}" \
   --account-key "${STORAGE_KEY}"


az postgres server create \
   --resource-group "${RESOURCE_GROUP}" \
   --name "${SQL_INSTANCE_NAME}" \
   --sku-name GP_Gen5_2 \
   --ssl-enforcement Disabled \
   --admin-password "${SQL_ADMIN_PASSWORD}" \
   --version 11 \
   --admin-user pachyderm

az postgres server firewall-rule create \
   --server-name "${SQL_INSTANCE_NAME}" \
   --resource-group "${RESOURCE_GROUP}" \
   --name AllowAllAzureIps \
   --start-ip-address 0.0.0.0 \
   --end-ip-address 0.0.0.0


az network public-ip create \
   --resource-group "${RESOURCE_GROUP}" \
   --name "${STATIC_IP_NAME}" \
   --sku Standard \
   --allocation-method static

STATIC_IP_ADDR=$(az network public-ip show --resource-group ${RESOURCE_GROUP} --name ${STATIC_IP_NAME} --query ipAddress --output tsv)
#STATIC_IP_ADDR=$(gcloud compute addresses describe ${STATIC_IP_NAME} --region=${GCP_REGION} --format=json --flatten=address | jq .[] )

# allow cluster access to resource group
CLIENT_ID=$(az aks  show --name "${CLUSTER_NAME}" --resource-group "${RESOURCE_GROUP}" | jq .identity.principalId -r)
RESOURCE_GROUP_SCOPE=$(az group show --resource-group "${RESOURCE_GROUP}" | jq .id -r)
az role assignment create \
   --assignee "${CLIENT_ID}" \
   --role "Network Contributor" \
   --scope "${RESOURCE_GROUP_SCOPE}"

az postgres db create \
   --resource-group ${RESOURCE_GROUP} \
   --server-name ${SQL_INSTANCE_NAME} \
   --name pachyderm

az postgres db create \
   --resource-group "${RESOURCE_GROUP}" \
   --server-name "${SQL_INSTANCE_NAME}" \
   --name dex

if [ "$EMPTY_CLUSTER" == "true" ]; then
    exit
fi

SQL_HOSTNAME=$(az postgres server show \
		  --resource-group "${RESOURCE_GROUP}" \
		  --name "${SQL_INSTANCE_NAME}" | jq .fullyQualifiedDomainName -r)
SQL_ADMIN=$(az postgres server show \
		  --resource-group "${RESOURCE_GROUP}" \
		  --name "${SQL_INSTANCE_NAME}" | jq .administratorLogin -r)@"${SQL_INSTANCE_NAME}"

cat <<EOF > ${NAME}.values.yaml
deployTarget: "MICROSOFT"

pachd:
  enabled: true
  externalService:
    enabled: true
    aPIGrpcport:    31400
    loadBalancerIP: ${STATIC_IP_ADDR}
    annotations:
      service.beta.kubernetes.io/azure-load-balancer-resource-group: ${RESOURCE_GROUP}
  image:
    tag: "2.1.1"
  storage:
    microsoft:
      container: "${CONTAINER_NAME}"
      id: "${STORAGE_ACCOUNT_NAME}"
      secret: "${STORAGE_KEY}"
    externalService:
      enabled: true

postgresql:
  enabled: false

global:
  postgresql:
    postgresqlHost: "${SQL_HOSTNAME}"
    postgresqlPort: "5432"
    postgresqlDatabase: "pachyderm"
    postgresqlUsername: "${SQL_ADMIN}"
    postgresqlPassword: "${SQL_ADMIN_PASSWORD}"
    postgresqlSSL: "enable"
EOF

helm repo add pach https://helm.pachyderm.com
helm repo update
helm install pachyderm -f "./${NAME}.values.yaml" pach/pachyderm

echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite
pachctl config set active-context "${CLUSTER_NAME}"
