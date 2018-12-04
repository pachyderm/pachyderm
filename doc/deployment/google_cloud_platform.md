# Google Cloud Platform

Google Cloud Platform has excellent support for Kubernetes, and thus Pachyderm, through the [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE). The following guide will walk you through deploying a Pachyderm cluster on GCP.

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 124.0.0
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [pachctl](#install-pachctl) 

If this is the first time you use the SDK, make sure to follow the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts). Note, this may update your `~/.bash_profile` and point your `$PATH` at the location where you extracted `google-cloud-sdk`. We recommend extracting the SDK to `~/bin`.

Note, you can also install `kubectl` installed via the Google Cloud SDK using:

```sh
$ gcloud components install kubectl
```

### Deploy Kubernetes

To create a new Kubernetes cluster via GKE, run:

```sh
$ CLUSTER_NAME=<any unique name, e.g. "pach-cluster">

$ GCP_ZONE=<a GCP availability zone. e.g. "us-west1-a">

$ gcloud config set compute/zone ${GCP_ZONE}

$ gcloud config set container/cluster ${CLUSTER_NAME}

$ MACHINE_TYPE=<machine type for the k8s nodes, we recommend "n1-standard-4" or larger>

# By default the following command spins up a 3-node cluster. You can change the default with `--num-nodes VAL`.
$ gcloud container clusters create ${CLUSTER_NAME} --scopes storage-rw --machine-type ${MACHINE_TYPE} 
```
Note that you must create the Kubernetes cluster via the gcloud command-line tool rather than the Google Cloud Console, as it's currently only possible to grant the `storage-rw` scope via the command-line tool. Also note, you should deploy a 1.8.x cluster if possible to take full advantage of Pachyderm's latest features.

This may take a few minutes to start up. You can check the status on the [GCP Console](https://console.cloud.google.com/compute/instances).  A `kubeconfig` entry will automatically be generated and set as the current context.  As a sanity check, make sure your cluster is up and running via `kubectl`:

```sh
# List all pods in the kube-system namespace.
$ kubectl get pods -n kube-system
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

If you *don't* see something similar to the above output, you can point `kubectl` to the new cluster manually via:

```sh
# Update your kubeconfig to point at your newly created cluster.
$ gcloud container clusters get-credentials ${CLUSTER_NAME}
```

### Deploy Pachyderm

To deploy Pachyderm we will need to:

1. [Create some storage resources](#set-up-the-storage-resources), 
2. [Install the Pachyderm CLI tool, `pachctl`](#install-pachctl), and
3. [Deploy Pachyderm on the k8s cluster](#deploy-pachyderm-on-the-k8s-cluster)

#### Set up the Storage Resources

Pachyderm needs a [GCS bucket](https://cloud.google.com/storage/docs/) and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.  We can specify the size of the persistent disk, the bucket name, and create the bucket as follows:

```sh
# For the persistent disk, 10GB is a good size to start with. 
# This stores PFS metadata. For reference, 1GB
# should work fine for 1000 commits on 1000 files.
$ STORAGE_SIZE=<the size of the volume that you are going to create, in GBs. e.g. "10">

# The Pachyderm bucket name needs to be globally unique across the entire GCP region.
$ BUCKET_NAME=<The name of the GCS bucket where your data will be stored>

# Create the bucket.
$ gsutil mb gs://${BUCKET_NAME}
```

To check that everything has been set up correctly, try:

```sh
$ gsutil ls
# You should see the bucket you created.
```

#### Install `pachctl`

`pachctl` is a command-line utility for interacting with a Pachyderm cluster. You can install it locally as follows:

```sh
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.8

# For Linux (64 bit) or Window 10+ on WSL:
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.8.0/pachctl_1.8.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can then run `pachctl version --client-only` to check that the installation was successful.

```sh
$ pachctl version --client-only
1.7.0
```

#### Deploy Pachyderm on the k8s cluster

Now we're ready to deploy Pachyderm itself.  This can be done in one command:

```sh
$ pachctl deploy google ${BUCKET_NAME} ${STORAGE_SIZE} --dynamic-etcd-nodes=1
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

Note, here we are using 1 etcd node to manage Pachyderm metadata. The number of etcd nodes can be adjusted as needed. Also, RBAC can be enabled as further documented [here](rbac.html).

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status with `kubectl`, which should output the following when Pachyderm is up and running:

```sh
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-482120938-np8cc     2/2       Running   0          4m
etcd-0                   1/1       Running   0          4m
pachd-3677268306-9sqm0   1/1       Running   0          4m
```

If you see a few restarts on the `pachd` pod, that's totally ok. That simply means that Kubernetes tried to bring up those containers before other components were ready, so it restarted them.

Finally, assuming your `pachd` is running as shown above, we need to set up forward a port so that `pachctl` can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.0
pachd               1.7.0
```
