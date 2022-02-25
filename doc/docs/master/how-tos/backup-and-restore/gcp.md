# Backing up and restoring Pachyderm with Velero

## Setup

### Per-project, one-time setup

```shell
BUCKET=…
PROJECT_ID=$(gcloud config get-value project)
SECRETS_PATH=…
gcloud iam service-accounts create velero --display-name "Velero service account"
SERVICE_ACCOUNT_EMAIL=$(gcloud iam service-accounts list --filter="displayName:Velero service account" --format 'value(email)')
gsutil mb gs://$BUCKET/
gsutil iam ch serviceAccount:$SERVICE_ACCOUNT_EMAIL:objectAdmin gs://${BUCKET}
gcloud iam service-accounts keys create $SECRETS_PATH --iam-account $SERVICE_ACCOUNT_EMAIL
ROLE_PERMISSIONS=(compute.disks.get compute.disks.create compute.disks.createSnapshot compute.snapshots.get compute.snapshots.create compute.snapshots.useReadOnly compute.snapshots.delete compute.zones.get)
gcloud iam roles create velero.server --project $PROJECT_ID --title "Velero Server" --permissions "$(IFS=","; echo "${ROLE_PERMISSIONS[*]}")"
```

### Per cluster

```shell
$ SERVICE_ACCOUNT=$(gcloud sql instances describe $INSTANCE --project=$PROJECT --format="value(serviceAccountEmailAddress)")
$ gsutil iam ch serviceAccount:${SERVICE_ACCOUNT}:objectCreator gs://$BUCKET
$ velero install --provider gcp --plugins velero/velero-plugin-for-gcp:v1.3.0 --bucket $BUCKET --secret-file $SECRETS_PATH
```

## Backing up

### Prerequisites

 - A Pachyderm cluster
 - `kubectl` configured to access that cluster
 - `gcloud` configured appropriately
 - execution of the setup steps above for both the GCP project and the
cluster

### Procedure

 1. Scale `pachd` down: `kubectl scale deployment pachd --replicas 0`
 2. Scale workers down: `kubectl scale rc --replicas 0 -l suite=pachyderm,component=worker`
 3. Take a backup: `velero backup create NAME --include-namespaces PACHD-NAMESPACE`
 4. Backup pachyderm PostgreSQL database: `gcloud sql export sql $INSTANCE gs://$BUCKET/sql/NAME/pachyderm.sql -d pachyderm`
 5. Backup dex PostgreSQL database: `gcloud sql export sql $INSTANCE gs://$BUCKET/sql/NAME/dex.sql -d dex`
 6. Backup postgres PostgreSQL database: `gcloud sql export sql
    $INSTANCE gs://$BUCKET/sql/NAME/postgres.sql -d postgres`
 7. Backup objext: `gsutil cp -r $BUCKET $BACKUP_BUCKET`
 7. Scale `pachd` back up: `kubectl scale deployment pachd --replicas 1`

## Restoring

### Prerequisites

 - An empty cluster
 - `kubectl` configured to access that cluster
 - `gcloud` configured appropriately

## Procedure

 1. Install Velero: `velero install --provider gcp --plugins
    velero/velero-plugin-for-gcp:v1.3.0 --bucket $BUCKET --secret-file
    $SECRETS_PATH`
 2. Give service account permissions: `gsutil iam ch serviceAccount:$(gcloud sql instances describe $INSTANCE --project=$PROJECT --format="value(serviceAccountEmailAddress)"):objectViewer gs://$BUCKET`
 3. If using a new database:
    1. Restore postgres: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/postgres.sql -d postgres`
	2. Restore dex: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/dex.sql -d dex`
	3. Restore pachyderm: `gcloud sql import sql $INSTANCE gs://$BUCKET/sql/NAME/pachyderm.sql -d pachyderm`
 4. Restore Pachyderm: `velero restore create --from-backup NAME`
 5. If using a new database:
	1. Update to point to new database: `kubectl edit deployment cloudsql-auth-proxy` (update line beginning with `- -instances=` to reference new DB)
	2. Update to point to new service account: `kubectl edit serviceaccount k8s-cloudsql-auth-proxy` (update
       `iam.gke.io/gcp-service-account` and delete `secrets`)
 6. Update pachyderm-worker service account: `kubectl edit serviceaccount pachyderm` (update
       `iam.gke.io/gcp-service-account` and delete `secrets`)
 7. Update pachyderm-worker service account: `kubectl edit serviceaccount pachyderm-worker` (update
       `iam.gke.io/gcp-service-account` and delete `secrets`)
 8. `kubectl edit secret pachyderm-storage-secret` and update Google bucket
 9. Scale `pachd` back up: `kubectl scale deployment pachd --replicas 1`
 10. Use `gcloud compute addresses list` to get the IP address for your new
 instance
 11. `kubectl edit service pachd-lb` (set `loadBalancerIP`)

## Connecting to your restored cluster

Pachctl must still be configured to connect to a restored cluster.

  1.  Use `gcloud compute addresses list` to get the IP address for your new
instance.
  2.  `STATIC_IP_ADDR=…`
  3.  `echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite`
  4.  `pachctl config set active-context ${CLUSTER_NAME}`

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
