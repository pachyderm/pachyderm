# Deploying On the Cloud

## Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, Pachyderm can run on any platform that supports Kubernetes. This guide covers the following commonly used platforms:

* [Google Cloud Platform](#google-cloud-platform)
* [AWS](#amazon-web-services-aws)
* [OpenShift](#openshift)


## Google Cloud Platform


Google Cloud Platform has excellent support for Kubernetes through [Google Container Engine](https://cloud.google.com/container-engine/).

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 124.0.0

If this is the first time you use the SDK, make sure to follow through the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts). This may update your `~/.bash_profile` and point your `$PATH` at the location where you extracted `google-cloud-sdk`. We recommend extracting this to `~/bin`.

If you do not already have `kubectl` installed: after the SDK is installed, run:

```shell
$ gcloud components install kubectl
```

This will download the `kubectl` binary to `google-cloud-sdk/bin`

### Deploy Kubernetes

To create a new Kubernetes cluster in GKE, just run:

```
$ CLUSTER_NAME=[any unique name, e.g. pach-cluster]

$ GCP_ZONE=[a GCP availability zone. e.g. us-west1-a]

$ gcloud config set compute/zone ${GCP_ZONE}

$ gcloud config set container/cluster ${CLUSTER_NAME}

# By default this spins up a 3-node cluster. You can change the default with `--num-nodes VAL`
$ gcloud container clusters create ${CLUSTER_NAME} --scopes storage-rw
```
This may take a few minutes to start up. You can check the status on the [GCP Console](https://console.cloud.google.com/compute/instances). 

```
# Update your kubeconfig to point at your newly created cluster
$ gcloud container clusters get-credentials ${CLUSTER_NAME}
```

Check to see that your cluster is up and running:
```
$ kubectl get all
NAME         CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   10.3.240.1   <none>        443/TCP   10m
```

### Deploy Pachyderm

#### Set up the storage infrastructure

Pachyderm needs a [GCS bucket](https://cloud.google.com/storage/docs/) and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.

Here are the parameters to create these resources:

```shell
# BUCKET_NAME needs to be globally unique across the entire GCP region
$ BUCKET_NAME=[The name of the GCS bucket where your data will be stored] 

# Name this whatever you want, we chose pach-disk as a default
$ STORAGE_NAME=pach-disk 

# For a demo you should only need 10 GB. This stores PFS metadata. For reference, 1GB
# should work for 1000 commits on 1000 files. 
$ STORAGE_SIZE=[the size of the volume that you are going to create, in GBs. e.g. "10"]
```

And then run:
```shell
$ gsutil mb gs://${BUCKET_NAME}
$ gcloud compute disks create --size=${STORAGE_SIZE}GB ${STORAGE_NAME}
```
To check that everything has been set up correctly, try:

```shell
$ gcloud compute instances list
# should see a number of instances

$ gsutil ls
# should see a bucket

$ gcloud compute disks list
# should see a number of disks, including the one you specified
```

#### Install Pachctl 

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version. 

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.2.0
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

#### Start Pachyderm

Now we're ready to boot up Pachyderm:
```
$ pachctl deploy google ${BUCKET_NAME} ${STORAGE_NAME} ${STORAGE_SIZE}
```

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME                   DESIRED        CURRENT          AGE
etcd                   1              1                1m
pachd                  2              2                1m
rethink                1              1                1m
NAME                   CLUSTER-IP     EXTERNAL-IP      PORT(S)                        AGE
etcd                   10.3.253.161   <none>           2379/TCP,2380/TCP              1m
kubernetes             10.3.240.1     <none>           443/TCP                        47m
pachd                  10.3.254.31    <nodes>          650/TCP,651/TCP                1m
rethink                10.3.241.56    <nodes>          8080/TCP,28015/TCP,29015/TCP   1m
NAME                   READY          STATUS           RESTARTS                       AGE
etcd-1mv3v             1/1            Running          0                              1m
pachd-6vjpc            1/1            Running          3                              1m
pachd-nxj54            1/1            Running          3                              1m
rethink-e4v60          1/1            Running          0                              1m
NAME                   STATUS         VOLUME           CAPACITY                       ACCESSMODES   AGE
rethink-volume-claim   Bound          rethink-volume   10Gi                           RWO           1m
```
Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.  

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks. 
$ pachctl portforward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.2.0
pachd               1.2.0
```

## Amazon Web Services (AWS)

### Prerequisites

- Make sure you have the [AWS CLI](https://aws.amazon.com/cli/) installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.


### Deploy Kubernetes

The easiest way to deploy a Kubernetes cluster is to use the [official Kubernetes guide](http://kubernetes.io/docs/getting-started-guides/aws/).

WARNING: As of 9/1/16, the guide has a [minor bug](https://github.com/kubernetes/kubernetes/issues/30495). TLDR: You need to add three lines in `~/kubernetes/aws/cluster/kubernetes/cluster/common.sh` starting at line 524. 

```
function build-kube-env {
  local master=$1
  local file=$2

  # Add these lines
  KUBE_MANIFESTS_TAR_URL="${SERVER_BINARY_TAR_URL/server-linux-amd64/manifests}"
  MASTER_OS_DISTRIBUTION="${KUBE_MASTER_OS_DISTRIBUTION}"
  NODE_OS_DISTRIBUTION="${KUBE_NODE_OS_DISTRIBUTION}"

  local server_binary_tar_url=$SERVER_BINARY_TAR_URL
  local salt_tar_url=$SALT_TAR_URL
  local kube_manifests_tar_url="${KUBE_MANIFESTS_TAR_URL:-}"
  if [[ "${master}" == "true" && "${MASTER_OS_DISTRIBUTION}" == "coreos" ]] || \
     [[ "${master}" == "false" && "${NODE_OS_DISTRIBUTION}" == "coreos" ]] ; then
    # TODO: Support fallback .tar.gz settings on CoreOS
    server_binary_tar_url=$(split_csv "${SERVER_BINARY_TAR_URL}")
    salt_tar_url=$(split_csv "${SALT_TAR_URL}")
    kube_manifests_tar_url=$(split_csv "${KUBE_MANIFESTS_TAR_URL}")
 ```

 NOTE: If you already had kubectl set up from the minikube demo, kubectl will now be talking to your aws cluster. You can switch back to talking to minikube with:
 ```
 kubectl config use-context minikube
 ``` 

 Now we've got Kubernetes up and running, it's time to deploy Pachyderm!

### Deploy Pachyderm

Before we deploy Pachyderm, we need to add some storage resources to our cluster so that Pachyderm has a place to put data. 

#### Set up the storage infrastructure

Pachyderm needs an [S3 bucket](https://aws.amazon.com/documentation/s3/), and a [persistent disk](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html)(EBS) to function correctly.  

Here are the parameters to set up these resources:

```shell
$ KUBECTLFLAGS="-s [The public IP of the Kubernetes master:`kubectl cluster-info`]"

# BUCKET_NAME needs to be globally unique across the entire AWS region
$ BUCKET_NAME=[The name of the S3 bucket where your data will be stored] 

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files. 
$ STORAGE_SIZE=[the size of the EBS volume that you are going to create, in GBs. e.g. "10"]

$ AWS_REGION=[the AWS region of your Kubernetes cluster. e.g. "us-west-2"]

$ AWS_AVAILABILITY_ZONE=[the AWS availability zone of your Kubernetes cluster. e.g. "us-west-2a"]

```

And then run:
```
$ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}

$ aws ec2 create-volume --size ${STORAGE_SIZE} --region ${AWS_REGION} --availability-zone ${AWS_AVAILABILITY_ZONE} --volume-type gp2
```

Record the "volume-id" that is output (e.g. "vol-8050b807"). You can also view it in the aws console or with  `aws ec2 describe-volumes`. Export the volume-id:

```shell
$ STORAGE_NAME=[volume id]
```

Now you should be able to see the bucket and the EBS volume that are just created:

```shell
aws s3api list-buckets --query 'Buckets[].Name'
aws ec2 describe-volumes --query 'Volumes[].VolumeId'
```

#### Install Pachctl 

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version. 

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.1.0
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.

#### Start Pachyderm

First get a set of [temporary AWS credentials](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html) by using this command:

```shell
$ aws sts get-session-token
```
Then set these variables:
```sh
$ AWS_ID=[access key ID] 

$ AWS_KEY=[secret access key]

$ AWS_TOKEN=[session token] 
```
Run the following command to deploy your Pachyderm cluster:

```shell
$ pachctl deploy amazon ${BUCKET_NAME} ${AWS_ID} ${AWS_KEY} ${AWS_TOKEN} ${AWS_REGION} ${STORAGE_NAME} ${STORAGE_SIZE}
```

It may take a few minutes for the pachd nodes to be running because it's pulling containers from DockerHub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME                   DESIRED        CURRENT          AGE
etcd                   1              1                17m
pachd                  2              2                17m
rethink                1              1                17m
NAME                   CLUSTER-IP     EXTERNAL-IP      PORT(S)                        AGE
etcd                   10.0.255.155   <none>           2379/TCP,2380/TCP              17m
kubernetes             10.0.0.1       <none>           443/TCP                        4h
pachd                  10.0.43.148    <nodes>          650/TCP,651/TCP                17m
rethink                10.0.249.8     <nodes>          8080/TCP,28015/TCP,29015/TCP   17m
NAME                   READY          STATUS           RESTARTS                       AGE
etcd-04jbq             1/1            Running          0                              17m
pachd-7a8sp            1/1            Running          2                              17m
pachd-9egd7            1/1            Running          3                              17m
rethink-xd7sc          1/1            Running          0                              17m
NAME                   STATUS         VOLUME           CAPACITY                       ACCESSMODES   AGE
rethink-volume-claim   Bound          rethink-volume   10Gi                           RWO           17m
```
Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.  

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks. 
$ pachctl portforward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.
```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.2.0
pachd               1.2.0
```

## OpenShift

[OpenShift](https://www.openshift.com/) is a popular enterprise Kubernetes distribution.  Pachyderm can run on OpenShift with two additional steps:

1. Make sure that priviledge containers are allowed (they are not allowed by default):  `oc edit scc` and set `allowPrivilegedContainer: true` everywhere.
2. Remove `hostPath` everywhere from your cluster manifest (e.g. `etc/kube/pachyderm-versioned.json` if you are deploying locally).

Problems related to OpenShift deployment are tracked in this issue: https://github.com/pachyderm/pachyderm/issues/336

## Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.
