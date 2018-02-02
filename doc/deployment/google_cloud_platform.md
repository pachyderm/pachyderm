# Google Cloud Platform

Google Cloud Platform has excellent support for Kubernetes through the [Google Container Engine](https://cloud.google.com/container-engine/).

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 124.0.0
- [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
- [pachctl](#install-pachctl)

If this is the first time you use the SDK, make sure to follow the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts). This may update your `~/.bash_profile` and point your `$PATH` at the location where you extracted `google-cloud-sdk`. We recommend extracting this to `~/bin`.

Note, you can also install `kubectl` installed via the SDK using:

```sh
$ gcloud components install kubectl
```

This will download the `kubectl` binary to `google-cloud-sdk/bin`

### Deploy Kubernetes

To create a new Kubernetes cluster in GKE, run:

```sh
$ CLUSTER_NAME=[any unique name, e.g. pach-cluster]

$ GCP_ZONE=[a GCP availability zone. e.g. us-west1-a]

$ gcloud config set compute/zone ${GCP_ZONE}

$ gcloud config set container/cluster ${CLUSTER_NAME}

$ MACHINE_TYPE=[machine for the k8s nodes. We recommend "n1-standard-4" or larger.]

# By default this spins up a 3-node cluster. You can change the default with `--num-nodes VAL`
$ gcloud container clusters create ${CLUSTER_NAME} --scopes storage-rw --machine-type ${MACHINE_TYPE}
```
Note that you must create the Kubernetes cluster via the gcloud command-line tool rather than the Google Cloud Console, as it's currently only possible to grant the `storage-rw` scope via the command-line tool.

This may take a few minutes to start up. You can check the status on the [GCP Console](https://console.cloud.google.com/compute/instances).  A `kubeconfig` entry will automatically be generated and set as the current context.  As a sanity check, make sure your cluster is up and running via `kubectl`:

```sh
# List all resources in the kube-system namespace
$ kubectl get all -n kube-system
```

If a list of Deployments, Replica Sets, and Pods are *not* displayed after the cluster is up, you can point `kubectl` to the new cluster via:

```sh
# Update your kubeconfig to point at your newly created cluster
$ gcloud container clusters get-credentials ${CLUSTER_NAME}
```

### Deploy Pachyderm

To deploy Pachyderm we will need to:

1. Add some storage resources on Google, 
2. Install the Pachyderm CLI tool, `pachctl`, and
3. Deploy Pachyderm on top of the storage resources.

#### Set up the Storage Resources

Pachyderm needs a [GCS bucket](https://cloud.google.com/storage/docs/) and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.  We can specify the size of the persistent disk, the bucket name, and create the bucket as follows:

```sh
# For a demo you should only need 10 GB. This stores PFS metadata. For reference, 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the volume that you are going to create, in GBs. e.g. "10"]

# BUCKET_NAME needs to be globally unique across the entire GCP region.
$ BUCKET_NAME=[The name of the GCS bucket where your data will be stored]

# Create the bucket.
$ gsutil mb gs://${BUCKET_NAME}
```

To check that everything has been set up correctly, try:

```sh
$ gcloud compute instances list
# should see a number of instances

$ gsutil ls
# should see a bucket
```

#### Install `pachctl`

`pachctl` is a command-line utility for interacting with a Pachyderm cluster.

```sh
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.6

# For Linux (64 bit) or Window 10+ on WSL:
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.6.8/pachctl_1.6.8_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can run `pachctl version --client-only` to check that the installation was successful.

```sh
$ pachctl version --client-only
1.6.6
```

#### Deploy Pachyderm

Now we're ready to deploy Pachyderm itself.  This can be done in one command:

```sh
$ pachctl deploy google ${BUCKET_NAME} ${STORAGE_SIZE} --dynamic-etcd-nodes=1
```

Note, here we are using 1 etcd node to manage Pachyderm metadata. The number of etcd nodes can be adjusted as needed.

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/dash    1         1         1            1           1m
deploy/pachd   1         1         1            1           1m

NAME                  DESIRED   CURRENT   READY     AGE
rs/dash-3149425249    1         1         1         1m
rs/pachd-4216341626   1         1         1         1m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/dash    1         1         1            1           1m
deploy/pachd   1         1         1            1           1m

NAME                DESIRED   CURRENT   AGE
statefulsets/etcd   1         1         1m

NAME                        READY     STATUS    RESTARTS   AGE
po/dash-3149425249-s5kdw    2/2       Running   0          1m
po/etcd-0                   1/1       Running   0          1m
po/pachd-4216341626-zxf5g   1/1       Running   0          1m
```

Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before other components were ready, so it restarted them.

Finally, assuming your `pachd` is running as shown above, we need to set up forward a port so that `pachctl` can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.6.6
pachd               1.6.6
```
