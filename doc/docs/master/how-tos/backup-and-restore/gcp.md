# Backing up and restoring Pachyderm

## Setup

Retain (ideally in version control) a copy of the original values file
used to install the cluster.

## Backing up

### Prerequisites

 - A Pachyderm cluster
 - `kubectl` configured to access that cluster
 - `gcloud` configured appropriately
 - A bucket for backup use

### Procedure

 1. Scale `pachd` down: `kubectl scale deployment pachd --replicas 0`
 2. Scale workers down: `kubectl scale rc --replicas 0 -l suite=pachyderm,component=worker`
 3. Backup pachyderm PostgreSQL database: `gcloud sql export sql $INSTANCE gs://$BACKUP_BUCKET/sql/NAME/pachyderm.sql -d pachyderm`
	1.  If this is the first time, you may need to run `gsutil iam ch serviceAccount:$(gcloud sql instances describe $INSTANCE --project=$PROJECT --format="value(serviceAccountEmailAddress)"):objectAdmin gs://$BACKUP_BUCKET`
 4. Backup dex PostgreSQL database: `gcloud sql export sql $INSTANCE gs://$BACKUP_BUCKET/sql/NAME/dex.sql -d dex`
 5. Backup postgres PostgreSQL database: `gcloud sql export sql $INSTANCE gs://$BACKUP_BUCKET/sql/NAME/postgres.sql -d postgres`
 6. Backup object store: `gsutil -m cp -r s3://$BUCKET s3://$BACKUP_BUCKET`
 7. Scale `pachd` back up: `kubectl scale deployment pachd --replicas 1`

## Restoring

### Prerequisites

 - A new SQL instance
 - A new bucket
 - An empty Kubernetes cluster
 - `kubectl` configured to access that cluster
 - `gcloud` configured appropriately

Note that the SQL instance, bucket and cluster may all be created with
the script below.

## Procedure

 1. Give service account permissions: `gsutil iam ch serviceAccount:$(gcloud sql instances describe $INSTANCE --project=$PROJECT --format="value(serviceAccountEmailAddress)"):objectViewer gs://$BUCKET`
 1. Restore postgres: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/postgres.sql -d postgres`
 2. Restore dex: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/dex.sql -d dex`
 3. Restore pachyderm: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/pachyderm.sql -d pachyderm`
 3. Copy objects from backup to restoration bucket: `gsutil -m cp -r s3://$BACKUP_BUCKET s3://$BUCKET`
 4. Update a copy of the original `values.yaml`:
    - `pachd.externalService.loadBalancerIP`: use `gcloud compute addressses list` to get
      the IP for the restoration cluster
    - `pachd.storage.google.bucket`: update to point to restoration
      bucket
    - `pachd.serviceAccount.additionalAnnotations.iam.gke.io/gcp-service-account`:
      update
    - `pachd.worker.serviceAccount.AdditionalAnnotations.iam.gke.io/gcp-service-account`:
      update
    - `cloudsqlAuthProxy.connectionName`: update
    - `cloudsqlAuthProxy.serviceAccount`: update
	- `global.postgresql.postgresqlPassword`: update if changed
 5. Install Pachyderm: `helm install pachyderm -f $VALUES_YAML pach/pachyderm`

## Connecting to your restored cluster

Pachctl must still be configured to connect to a restored cluster.

  1.  Use `gcloud compute addresses list` to get the IP address for your new
instance.
  1.  `STATIC_IP_ADDR=…`
  2.  `echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite`
  3.  `pachctl config set active-context ${CLUSTER_NAME}`

## Appendix

### Creating an empty cluster

The following procedure may be used to create an empty cluster.

```shell
PROJECT_ID=…
NAME=…
SQL_ADMIN_PASSWORD=…

# change these if desired
GCP_ZONE="us-central1-a"
CLUSTER_MACHINE_TYPE="n1-standard-4"

CLUSTER_NAME="${NAME}-gke"
BUCKET_NAME="${NAME}-gcs"
CLOUDSQL_INSTANCE_NAME="${NAME}-sql"
GSA_NAME="${NAME}-gsa"
STATIC_IP_NAME="${NAME}-ip"

gcloud config set project ${PROJECT_ID}
gcloud config set compute/zone ${GCP_ZONE}
gcloud config set container/cluster ${CLUSTER_NAME}

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

gcloud sql databases create pachyderm -i ${CLOUDSQL_INSTANCE_NAME}

gcloud sql databases create dex -i ${CLOUDSQL_INSTANCE_NAME}
```
