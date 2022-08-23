#!/bin/bash

set -xeou pipefail

# This script will create a file called ${NAME}.values.yaml in the current directory
# This script will create a GKE cluster, the workload identity service accounts and permissions,
# a static ip, the cloudsql instance and databases, and the cloud storage bucket. It will also install pachyderm into the cluster.
# This script assumes you have created your own GCP project

# PLEASE CHANGE THESE NEXT 3 VARIABLES AT A MINIMUM
PROJECT_ID="pachyderm-gcp-test"
NAME="fuzzy-alpaca"
SQL_ADMIN_PASSWORD="batteryhorsestaple"

# This group of variables can be changed, but are sane defaults
GCP_REGION="us-central1"
GCP_ZONE="us-central1-a"
K8S_NAMESPACE="default"
CLUSTER_MACHINE_TYPE="n1-standard-4"
SQL_CPU="2"
SQL_MEM="7680MB"

# The following variables probably shouldn't be changed
CLUSTER_NAME="${NAME}-gke"
BUCKET_NAME="${NAME}-gcs"
LOKI_BUCKET_NAME="${NAME}-logs-gcs"
CLOUDSQL_INSTANCE_NAME="${NAME}-sql"
GSA_NAME="${NAME}-gsa"
LOKI_GSA_NAME="${NAME}-loki-gsa"
STATIC_IP_NAME="${NAME}-ip"

ROLE1="roles/cloudsql.client"
ROLE2="roles/storage.admin"

SERVICE_ACCOUNT="${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
LOKI_SERVICE_ACCOUNT="${LOKI_GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
PACH_WI="serviceAccount:${PROJECT_ID}.svc.id.goog[${K8S_NAMESPACE}/pachyderm]"
SIDECAR_WI="serviceAccount:${PROJECT_ID}.svc.id.goog[${K8S_NAMESPACE}/pachyderm-worker]"
CLOUDSQLAUTHPROXY_WI="serviceAccount:${PROJECT_ID}.svc.id.goog[${K8S_NAMESPACE}/k8s-cloudsql-auth-proxy]"


kubectl version --client=true

gcloud version

helm version

jq --version

gcloud config set project ${PROJECT_ID}

gcloud config set compute/zone ${GCP_ZONE}

gcloud config set container/cluster ${CLUSTER_NAME}

gcloud services enable container.googleapis.com

gcloud services enable sqladmin.googleapis.com

gcloud container clusters create ${CLUSTER_NAME} \
 --machine-type=${CLUSTER_MACHINE_TYPE} \
 --workload-pool=${PROJECT_ID}.svc.id.goog \
 --enable-ip-alias \
 --create-subnetwork="" \
 --enable-stackdriver-kubernetes \
 --enable-dataplane-v2 \
 --enable-shielded-nodes \
 --release-channel="regular" \
 --workload-metadata="GKE_METADATA" \
 --enable-autorepair \
 --enable-autoupgrade \
 --disk-type="pd-ssd" \
 --image-type="COS_CONTAINERD"

gcloud container clusters get-credentials ${CLUSTER_NAME}

gsutil mb -l ${GCP_REGION} gs://${BUCKET_NAME}
gsutil mb -l ${GCP_REGION} gs://${LOKI_BUCKET_NAME}

gcloud sql instances create ${CLOUDSQL_INSTANCE_NAME} \
  --database-version=POSTGRES_13 \
  --cpu=${SQL_CPU} \
  --memory=${SQL_MEM} \
  --zone=${GCP_ZONE} \
  --availability-type=ZONAL \
  --storage-size=50GB \
  --storage-type=SSD \
  --storage-auto-increase \
  --root-password=${SQL_ADMIN_PASSWORD}

gcloud compute addresses create ${STATIC_IP_NAME} --region=${GCP_REGION}

STATIC_IP_ADDR=$(gcloud compute addresses describe ${STATIC_IP_NAME} --region=${GCP_REGION} --format=json --flatten=address | jq .[] )

gcloud iam service-accounts create ${GSA_NAME}

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="${ROLE1}"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="${ROLE2}"

gcloud iam service-accounts add-iam-policy-binding ${SERVICE_ACCOUNT} \
    --role roles/iam.workloadIdentityUser \
    --member "${PACH_WI}"

gcloud iam service-accounts add-iam-policy-binding ${SERVICE_ACCOUNT} \
    --role roles/iam.workloadIdentityUser \
    --member "${SIDECAR_WI}"

gcloud iam service-accounts add-iam-policy-binding ${SERVICE_ACCOUNT} \
    --role roles/iam.workloadIdentityUser \
    --member "${CLOUDSQLAUTHPROXY_WI}"

gcloud iam service-accounts create ${LOKI_GSA_NAME}

gcloud iam service-accounts keys create "${LOKI_GSA_NAME}-key.json" --iam-account="$LOKI_SERVICE_ACCOUNT"

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${LOKI_SERVICE_ACCOUNT}" \
    --role="${ROLE2}"

kubectl create secret generic loki-service-account --from-file="${LOKI_GSA_NAME}-key.json"

gcloud sql databases create pachyderm -i "${CLOUDSQL_INSTANCE_NAME}"

gcloud sql databases create dex -i ${CLOUDSQL_INSTANCE_NAME}

CLOUDSQL_CONNECTION_NAME=$(gcloud sql instances describe ${CLOUDSQL_INSTANCE_NAME} --format=json | jq ."connectionName")

cat <<EOF > ${NAME}.values.yaml
deployTarget: "GOOGLE"

pachd:
  enabled: true
  externalService:
    enabled: true
    aPIGrpcport:    31400
    loadBalancerIP: ${STATIC_IP_ADDR}
  image:
    tag: "2.2.0"
  lokiDeploy: true
  lokiLogging: true
  storage:
    google:
      bucket: "${BUCKET_NAME}"
  serviceAccount:
    additionalAnnotations:
      iam.gke.io/gcp-service-account: "${SERVICE_ACCOUNT}"
    create: true
    name:   "pachyderm"
  worker:
    serviceAccount:
      additionalAnnotations:
        iam.gke.io/gcp-service-account: "${SERVICE_ACCOUNT}"
      create: true
      name:   "pachyderm-worker"

cloudsqlAuthProxy:
  enabled: true
  connectionName: ${CLOUDSQL_CONNECTION_NAME}
  serviceAccount: "${SERVICE_ACCOUNT}"
  resources:
    requests:
      memory: "500Mi"
      cpu:    "250m"

postgresql:
  enabled: false

global:
  postgresql:
    postgresqlHost: "cloudsql-auth-proxy.default.svc.cluster.local."
    postgresqlPort: "5432"
    postgresqlSSL: "disable"
    postgresqlUsername: "postgres"
    postgresqlPassword: "${SQL_ADMIN_PASSWORD}"

loki-stack:
  loki:
    env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /etc/secrets/${LOKI_GSA_NAME}-key.json
    extraVolumes:
      - name: loki-service-account
        secret:
          secretName: loki-service-account
    extraVolumeMounts:
      - name: loki-service-account
        mountPath: /etc/secrets
    config:
      schema_config:
        configs:
        - from: 1989-11-09
          object_store: gcs
          store: boltdb
          schema: v11
          index:
            prefix: loki_index_
          chunks:
            prefix: loki_chunks_
      storage_config:
        gcs:
          bucket_name: "${LOKI_BUCKET_NAME}"
        # https://github.com/grafana/loki/issues/256
        bigtable:
          project: project
          instance: instance
        boltdb:
          directory: /data/loki/indices
  grafana:
    enabled: true
EOF

helm repo add pach https://helm.pachyderm.com
helm repo update
helm install pachyderm -f ./${NAME}.values.yaml pach/pachyderm

STATIC_IP_ADDR_NO_QUOTES=$(echo "$STATIC_IP_ADDR" | tr -d '"')
echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR_NO_QUOTES}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite
pachctl config set active-context ${CLUSTER_NAME}
