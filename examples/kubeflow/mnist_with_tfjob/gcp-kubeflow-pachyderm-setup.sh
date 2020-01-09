# NOTE: GKE Resources cost Money. Proceed with care.
# This script is intended for guidance only and may not
# execute correctly in all environments.  Try
# setting all the environment variables here and
# executing the commands one-by-one to debug issues.
export CLUSTER_NAME="data-lineage-example2"
export GCP_ZONE="us-east1-b"
export MACHINE_TYPE="n1-standard-8" 

# For the persistent disk, 10GB is a good size to start with.
# This stores PFS metadata. For reference, 1GB
# should work fine for 1000 commits on 1000 files.
export STORAGE_SIZE=10

# The Pachyderm bucket name needs to be globally unique across the entire GCP region.
export BUCKET_NAME=${CLUSTER_NAME}-bucket

export KUBEFLOW_USERNAME='pachyderm'
export KUBEFLOW_PASSWORD='pachyderm'

# The following command is optional, to make kfctl binary easier to use.
#export PATH=$PATH:<path to kfctl in your kubeflow installation>
#export ZONE=${GCP_ZONE}  #where the deployment will be created

export PROJECT=$(gcloud config get-value project)
export KFAPP=$CLUSTER_NAME

# from macOS, this would allow you to run "open $KF_URL" from this
# command prompt to see the Kubeflow dashboard.  In Linux, the
# command might be "xdg-open $KF_URL".
export KF_URL="https://${KFAPP}.endpoints.${PROJECT}.cloud.goog"

gcloud auth login

gcloud config set compute/zone ${GCP_ZONE}

gcloud config set container/cluster ${CLUSTER_NAME}


# By default the following command spins up a 3-node cluster. You can change the default with `--num-nodes VAL`.
gcloud container clusters create ${CLUSTER_NAME} --scopes storage-rw --machine-type ${MACHINE_TYPE}

# todo: investigate these issues
# WARNING: In June 2019, node auto-upgrade will be enabled by default for newly created clusters and node pools. To disable it, use the `--no-enable-autoupgrade` flag.
# WARNING: Starting in 1.12, new clusters will have basic authentication disabled by default. Basic authentication can be enabled (or disabled) manually using the `--[no-]enable-basic-auth` flag.
# WARNING: Starting in 1.12, new clusters will not have a client certificate issued. You can manually enable (or disable) the issuance of the client certificate using the `--[no-]issue-client-certificate` flag.
# WARNING: Starting in 1.12, default node pools in new clusters will have their legacy Compute Engine instance metadata endpoints disabled by default. To create a cluster with legacy instance metadata endpoints disabled in the default node pool, run `clusters create` with the flag `--metadata disable-legacy-endpoints=true`.
# WARNING: The Pod address range limits the maximum size of the cluster. Please refer to https://cloud.google.com/kubernetes-engine/docs/how-to/flexible-pod-cidr to learn how to optimize IP address allocation.
# This will enable the autorepair feature for nodes. Please see https://cloud.google.com/kubernetes-engine/docs/node-auto-repair for more information on node autorepairs.


# By default, GKE clusters have RBAC enabled. To allow 'pachctl deploy' to give the 'pachyderm' service account
# the requisite privileges via clusterrolebindings, you will need to grant *your user account* the privileges
# needed to create those clusterrolebindings.
#
# Note that this command is simple and concise, but gives your user account more privileges than necessary. See
# https://docs.pachyderm.com/latest/deploy-manage/deploy/rbac/ for the complete list of privileges that the
# pachyderm serviceaccount needs.
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)


gcloud container clusters get-credentials ${CLUSTER_NAME}


# Create the bucket.
gsutil mb gs://${BUCKET_NAME}

kubectl create namespace kubeflow

pachctl deploy google ${BUCKET_NAME} ${STORAGE_SIZE} --dynamic-etcd-nodes=1 --namespace kubeflow
# Default uses Cloud IAP:
#kfctl init ${KFAPP} --platform gcp --project ${PROJECT}
# Alternatively, use this command if you want to use basic authentication:
#kfctl init ${KFAPP} --platform gcp --project ${PROJECT} --use_basic_auth -V
export CONFIG_URI="https://raw.githubusercontent.com/kubeflow/manifests/v0.7-branch/kfdef/kfctl_gcp_basic_auth.0.7.0.yaml"
kfctl init ${KFAPP} --platform gcp --project ${PROJECT} --config=${CONFIG} --skip-init-gcp-project --use_basic_auth -V

cd ${KFAPP}
kfctl generate all -V --zone ${GCP_ZONE}
kfctl apply all -V

