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
LOKI_CONTAINER_NAME="${NAME}-loki-container"
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

az storage container create --name "${LOKI_CONTAINER_NAME}" \
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
    tag: "2.2.0"
  lokiDeploy: true
  lokiLogging: true
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
    postgresqlSSL: "verify-full"
    # Below is the CA certificate which signs Azure Database for PostgreSQL
    # server certificates.  For more information, see: https://docs.microsoft.com/en-us/azure/postgresql/single-server/concepts-ssl-connection-security#applications-that-require-certificate-verification-for-tls-connectivity
    postgresqlSSLCACert: |
      -----BEGIN CERTIFICATE-----
      MIIDdzCCAl+gAwIBAgIEAgAAuTANBgkqhkiG9w0BAQUFADBaMQswCQYDVQQGEwJJ
      RTESMBAGA1UEChMJQmFsdGltb3JlMRMwEQYDVQQLEwpDeWJlclRydXN0MSIwIAYD
      VQQDExlCYWx0aW1vcmUgQ3liZXJUcnVzdCBSb290MB4XDTAwMDUxMjE4NDYwMFoX
      DTI1MDUxMjIzNTkwMFowWjELMAkGA1UEBhMCSUUxEjAQBgNVBAoTCUJhbHRpbW9y
      ZTETMBEGA1UECxMKQ3liZXJUcnVzdDEiMCAGA1UEAxMZQmFsdGltb3JlIEN5YmVy
      VHJ1c3QgUm9vdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKMEuyKr
      mD1X6CZymrV51Cni4eiVgLGw41uOKymaZN+hXe2wCQVt2yguzmKiYv60iNoS6zjr
      IZ3AQSsBUnuId9Mcj8e6uYi1agnnc+gRQKfRzMpijS3ljwumUNKoUMMo6vWrJYeK
      mpYcqWe4PwzV9/lSEy/CG9VwcPCPwBLKBsua4dnKM3p31vjsufFoREJIE9LAwqSu
      XmD+tqYF/LTdB1kC1FkYmGP1pWPgkAx9XbIGevOF6uvUA65ehD5f/xXtabz5OTZy
      dc93Uk3zyZAsuT3lySNTPx8kmCFcB5kpvcY67Oduhjprl3RjM71oGDHweI12v/ye
      jl0qhqdNkNwnGjkCAwEAAaNFMEMwHQYDVR0OBBYEFOWdWTCCR1jMrPoIVDaGezq1
      BE3wMBIGA1UdEwEB/wQIMAYBAf8CAQMwDgYDVR0PAQH/BAQDAgEGMA0GCSqGSIb3
      DQEBBQUAA4IBAQCFDF2O5G9RaEIFoN27TyclhAO992T9Ldcw46QQF+vaKSm2eT92
      9hkTI7gQCvlYpNRhcL0EYWoSihfVCr3FvDB81ukMJY2GQE/szKN+OMY3EU/t3Wgx
      jkzSswF07r51XgdIGn9w/xZchMB5hbgF/X++ZRGjD8ACtPhSNzkE1akxehi/oCr0
      Epn3o0WC4zxe9Z2etciefC7IpJ5OCBRLbf1wbWsaY71k5h+3zvDyny67G7fyUIhz
      ksLi4xaNmjICq44Y3ekQEe5+NauQrz4wlHrQMz2nZQ/1/I6eYs9HRCwBXbsdtTLS
      R9I4LtD+gdwyah617jzV/OeBHRnDJELqYzmp
      -----END CERTIFICATE-----

loki-stack:
  loki:
    config:
      schema_config:
        configs:
        - from: 1989-11-09
          object_store: azure
          store: boltdb
          schema: v11
          index:
            prefix: loki_index_
          chunks:
            prefix: loki_chunks_
      storage_config:
        azure:
          container_name: "${LOKI_CONTAINER_NAME}"
          account_name: "${STORAGE_ACCOUNT_NAME}"
          account_key: "${STORAGE_KEY}"
        boltdb:
          directory: /data/loki/indices
    persistence:
      storageClassName: default
  grafana:
    enabled: true
EOF

helm repo add pach https://helm.pachyderm.com
helm repo update
helm install pachyderm -f "./${NAME}.values.yaml" pach/pachyderm

echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite
pachctl config set active-context "${CLUSTER_NAME}"
