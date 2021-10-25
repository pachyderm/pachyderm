# Google Cloud Platform

Google Cloud Platform provides seamless support for Kubernetes.
Therefore, Pachyderm is fully supported on [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE).
The following section walks you through deploying a Pachyderm cluster on GKE.

## Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 124.0.0
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [pachctl](#install-pachctl)

If this is the first time you use the SDK, follow
the [Google SDK QuickStart Guide](https://cloud.google.com/sdk/docs/quickstarts).

!!! note
    When you follow the QuickStart Guide, you might update your `~/.bash_profile`
    and point your `$PATH` at the location where you extracted
    `google-cloud-sdk`. However, Pachyderm recommends that you extract
    the SDK to `~/bin`.

!!! tip
    You can install `kubectl` by using the Google Cloud SDK and
    running the following command:

    ```shell
    gcloud components install kubectl
    ```

## Deploy Kubernetes

To create a new Kubernetes cluster by using GKE, run:

```shell
CLUSTER_NAME=<any unique name, e.g. "pach-cluster">

GCP_ZONE=<a GCP availability zone. e.g. "us-west1-a">

gcloud config set compute/zone ${GCP_ZONE}

gcloud config set container/cluster ${CLUSTER_NAME}

MACHINE_TYPE=<machine type for the k8s nodes, we recommend "n1-standard-4" or larger>

# By default the following command spins up a 3-node cluster. You can change the default with `--num-nodes VAL`.
gcloud container clusters create ${CLUSTER_NAME} --scopes storage-rw --machine-type ${MACHINE_TYPE}

# By default, GKE clusters have RBAC enabled. To allow 'pachctl deploy' to give the 'pachyderm' service account
# the requisite privileges via clusterrolebindings, you will need to grant *your user account* the privileges
# needed to create those clusterrolebindings.
#
# Note that this command is simple and concise, but gives your user account more privileges than necessary. See
# https://docs.pachyderm.io/en/1.11.x/deploy-manage/deploy/rbac/#rbac for the complete list of privileges that the
# pachyderm serviceaccount needs.
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
```

!!! note "Important"
        - You must create the Kubernetes cluster by using the `gcloud` command-line
        tool rather than the Google Cloud Console, as you can grant the
        `storage-rw` scope through the command-line tool only.
        
        - Adding `--scopes storage-rw` to `gcloud container clusters create ${CLUSTER_NAME} --machine-type ${MACHINE_TYPE}` will grant the rw scope to whatever service account is on the cluster, which if you don’t provide it, is the default compute service account for the project which has Editor permissions. While this is **not recommended in any production settings**, this option can be useful for a quick setup in development. In that scenario, you do not need any service account or additional GCP Bucket permission.

This migth take a few minutes to start up. You can check the status on
the [GCP Console](https://console.cloud.google.com/compute/instances).
A `kubeconfig` entry is automatically generated and set as the current
context. As a sanity check, make sure your cluster is up and running
by running the following `kubectl` command:

```shell
# List all pods in the kube-system namespace.
kubectl get pods -n kube-system
```

**System Response:**

```shell
NAME                                                     READY     STATUS    RESTARTS   AGE
event-exporter-v0.1.7-5c4d9556cf-fd9j2                   2/2       Running   0          1m
fluentd-gcp-v2.0.9-68vhs                                 2/2       Running   0          1m
fluentd-gcp-v2.0.9-fzfpw                                 2/2       Running   0          1m
fluentd-gcp-v2.0.9-qvk8f                                 2/2       Running   0          1m
heapster-v1.4.3-5fbfb6bf55-xgdwx                         3/3       Running   0          55s
kube-dns-778977457c-7hbrv                                3/3       Running   0          1m
kube-dns-778977457c-dpff4                                3/3       Running   0          1m
kube-dns-autoscaler-7db47cb9b7-gp5ns                     1/1       Running   0          1m
kube-proxy-gke-pach-cluster-default-pool-9762dc84-bzcz   1/1       Running   0          1m
kube-proxy-gke-pach-cluster-default-pool-9762dc84-hqkr   1/1       Running   0          1m
kube-proxy-gke-pach-cluster-default-pool-9762dc84-jcbg   1/1       Running   0          1m
kubernetes-dashboard-768854d6dc-t75rp                    1/1       Running   0          1m
l7-default-backend-6497bcdb4d-w72k5                      1/1       Running   0          1m
```

If you *don't* see something similar to the above output,
you can point `kubectl` to the new cluster manually by running
the following command:

```shell
# Update your kubeconfig to point at your newly created cluster.
gcloud container clusters get-credentials ${CLUSTER_NAME}
```

## Deploy Pachyderm

To deploy Pachyderm we will need to:

1. [Create storage resources](#set-up-the-storage-resources), 
2. [Install the Pachyderm CLI tool, `pachctl`](#install-pachctl), and
3. [Deploy Pachyderm on the Kubernetes cluster](#deploy-pachyderm-on-the-kubernetes-cluster)
4. [Point your CLI `pachctl` to your cluster](#have-pachctl-and-your-cluster-communicate)
### Set up the Storage Resources

Pachyderm needs a [GCS bucket](https://cloud.google.com/storage/docs/) (Object store)
and a [persistent disk](https://cloud.google.com/compute/docs/disks/) (Metadata)
to function correctly. You can specify the size of the persistent
disk, the bucket name, and create the bucket by running the following
commands:

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your disk choice, you may need to oversize the volume significantly to ensure enough IOPS.


```shell
# For the persistent disk, this stores PFS metadata. For reference, 1GB should
# work fine for 1000 commits on 1000 files. 10GB is often a sufficient starting
# size, though we recommend provisioning at least 1500 write IOPS, which
# requires at least 50GB of space on SSD-based PDs and 1TB of space on Standard PDs.
# See https://cloud.google.com/compute/docs/disks/performance
STORAGE_SIZE=<the size of the volume that you are going to create, in GBs. e.g. "50">

# The Pachyderm bucket name needs to be globally unique across the entire GCP region.
BUCKET_NAME=<The name of the GCS bucket where your data will be stored>

# Create the bucket.
gsutil mb gs://${BUCKET_NAME}
```

To check that everything has been set up correctly, run:

```shell
gsutil ls
# You should see the bucket you created.
```

### Install `pachctl`

`pachctl` is a command-line utility for interacting with a Pachyderm cluster. You can install it locally as follows:

```shell
# For macOS:
brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}

# For Linux (64 bit) or Window 10+ on WSL:

$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can then run `pachctl version --client-only` to check that the installation was successful.

```shell
pachctl version --client-only
{{ config.pach_latest_version }}
```

### Deploy Pachyderm on the Kubernetes cluster

Now you can deploy a Pachyderm cluster by running this command:

```shell
pachctl deploy google ${BUCKET_NAME} ${STORAGE_SIZE} --dynamic-etcd-nodes=1
```

**System Response:**

```shell
serviceaccount "pachyderm" created
storageclass "etcd-storage-class" created
service "etcd-headless" created
statefulset "etcd" created
service "etcd" created
service "pachd" created
deployment "pachd" created
service "dash" created
deployment "dash" created
secret "pachyderm-storage-secret" created

Pachyderm is launching. Check its status with "kubectl get all"
Once launched, access the dashboard by running "pachctl port-forward"
```

!!! note
    Pachyderm uses one etcd node to manage Pachyderm metadata.

!!! note "Important"
    If RBAC authorization is a requirement or you run into any RBAC
    errors see [Configure RBAC](rbac.md).

It may take a few minutes for the pachd nodes to be running because Pachyderm
pulls containers from DockerHub. You can see the cluster status with
`kubectl`, which should output the following when Pachyderm is up and running:

```shell
kubectl get pods
```

**System Response:**

```shell
NAME                     READY     STATUS    RESTARTS   AGE
dash-482120938-np8cc     2/2       Running   0          4m
etcd-0                   1/1       Running   0          4m
pachd-3677268306-9sqm0   1/1       Running   0          4m
```

If you see a few restarts on the `pachd` pod, you can safely ignore them.
That simply means that Kubernetes tried to bring up those containers
before other components were ready, so it restarted them.

### Have 'pachctl' and your Cluster Communicate

Finally, assuming your `pachd` is running as shown above, 
make sure that `pachctl` can talk to the cluster by:

- Running a port-forward:

```shell
# Background this process because it blocks.
pachctl port-forward   
```

- Exposing your cluster to the internet by setting up a LoadBalancer as follow:

!!! Warning 
    The following setup of a LoadBalancer only applies to pachd.

1. To get an external IP address for a Cluster, edit its k8s service, 
```shell
kubectl edit service pachd
```
and change its `spec.type` value from `NodePort` to `LoadBalancer`. 

1. Retrieve the external IP address of the edited service.
When listing your services again, you should see an external IP address allocated to the service you just edited. 
```shell
kubectl get service
```
1. Update the context of your cluster with their direct url, using the external IP address above:
```shell
echo '{"pachd_address": "grpc://<external-IP-address>:650"}' | pachctl config set context "your-cluster-context-name" --overwrite
```
1. Check that your are using the right context: 
```shell
pachctl config get active-context`
```
Your cluster context name set above should show up. 
    

You are done! You can test to make sure the cluster is working
by running `pachctl version` or even creating a new repo.

```shell
pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

## Advanced Setups
### Increase Ingress Throughput

One way to improve Ingress performance is to restrict Pachd to
a specific, more powerful node in the cluster. This is
accomplished by the use of [node-taints](https://cloud.google.com/kubernetes-engine/docs/how-to/node-taints)
in GKE. By creating a node-taint for `pachd`, you configure the
Kubernetes scheduler to run only the `pachd` pod on that node. After
that’s completed, you can deploy Pachyderm with the `--pachd-cpu-request`
and `--pachd-memory-request` set to match the resources limits of the
machine type. And finally, you need to modify the `pachd` deployment
so that it has an appropriate toleration:

```shell
tolerations:
- key: "dedicated"
  operator: "Equal"
  value: "pachd"
  effect: "NoSchedule"
```

### Increase upload performance

The most straightfoward approach to increasing upload performance is
to [leverage SSD’s as the boot disk](https://cloud.google.com/kubernetes-engine/docs/how-to/custom-boot-disks) in
your cluster because SSDs provide higher throughput and lower latency than
HDD disks. Additionally, you can increase the size of the SSD for
further performance gains because the number of IOPS increases with
disk size.

### Increase merge performance

Performance tweaks when it comes to merges can be done directly in
the [Pachyderm pipeline spec](../../../reference/pipeline_spec/).
More specifically, you can increase the number of hashtrees (hashtree spec)
in the pipeline spec. This number determines the number of shards for the
filesystem metadata. In general this number should be lower than the number
of workers (parallelism spec) and should not be increased unless merge time
(the time before the job is done and after the number of processed datums +
skipped datums is equal to the total datums) is too slow.
