# Deploying On the Cloud

## Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, Pachyderm can run on any platform that supports Kubernetes. This guide covers the following commonly used platforms:

* [Google Cloud Platform](#google-cloud-platform)
* [AWS](#amazon-web-services-aws)
* [OpenShift](#openshift)


## Google Cloud Platform


Google Cloud Platform has excellent support for Kubernetes through [Google Container Engine](https://cloud.google.com/container-engine/).

### Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 106.0.0

If this is the first time you use the SDK, make sure to follow through the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts).

After the SDK is installed, run:

```shell
$ gcloud components install kubectl
```

### Set up the infrastructure

Pachyderm needs a [container cluster](https://cloud.google.com/container-engine/), a [GCS bucket](https://cloud.google.com/storage/docs/), and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.  

We've made this very easy for you by creating the `make google-cluster` helper, which will create all of these resources for you.

First of all, set the required environment variables. Choose a name for both the bucket and disk, as well as a capacity for the disk (in GB):

```shell
$ export BUCKET_NAME=some-unique-bucket-name
$ export STORAGE_NAME=pach-disk
$ export STORAGE_SIZE=200
```

You may need to visit the GCP console to fully initialize Container Engine in a new project. Then, simply run the following command:

```shell
$ make google-cluster
```

This creates a Kubernetes cluster named "pachyderm", a bucket, and a persistent disk.  To check that everything has been set up correctly, try:

```shell
$ gcloud compute instances list
# should see a number of instances

$ gsutil ls
# should see a bucket

$ gcloud compute disks list
# should see a number of disks, including the one you specified
```

### Deploy Pachyderm


First of all, record the external IP address of one of the nodes in your Kubernetes cluster:

```shell
$ gcloud compute instances list
```

Then export it with port 30650:

```shell
$ export ADDRESS=[the external address]:30650
# for example:
# export ADDRESS=104.197.179.185:30650
```

This is so we can use `pachctl` to talk to our cluster later.

Now you can deploy Pachyderm with:

```shell
$ make google-cluster-manifest > manifest
$ make MANIFEST=manifest launch
```

It may take a while to complete for the first time, as a lot of Docker images need to be pulled.

## Amazon Web Services (AWS)

### Prerequisites

- Make sure you have the [AWS CLI](https://aws.amazon.com/cli/) installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.


### Deploy Kubernetes

The rest of the guide assumes you already have Kubernetes running in AWS. If you do, jump to [deploying Pachyderm](#Deploy_Pachyderm).

If you don't have a Kubernetes cluster yet you can either set up a [demo cluster](http://kubernetes.io/docs/getting-started-guides/aws/) to get started quickly or [deploy a production-grade cluster](TODO). 

WARNING: As of 9/1/16, the demo cluster guide has a [minor bug](https://github.com/kubernetes/kubernetes/issues/30495). TLDR: You need to add three lines in `~/kubernetes/aws/cluster/kubernetes/cluster/common.sh` starting at line 524. 

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
  fi
 ```

 NOTE: If you already had kubectl set up from the minikube demo, kubectl will now be talking to your aws cluster. You can switch back to talking to minikube with:
 ```
 kubectl config use-context minikube
 ``` 

 Now we've got Kubernetes up and running, it's time to deploy Pachyderm!

## Deploy Pachyderm

Before we deploy Pachyderm, we need to add some storage resources to our cluster so that Pachyderm has a place to put data. 

### Set up the storage infrastructure

Pachyderm needs an [S3 bucket](https://aws.amazon.com/documentation/s3/), and a [persistent disk](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html)(EBS) to function correctly.  

Here are the parameters to set up these resources:

```shell
$ export KUBECTLFLAGS="-s [The public IP of the Kubernetes master:`kubectl cluster-info`]"

# BUCKET_NAME needs to be globally unique across the entire AWS region
$ export BUCKET_NAME=[The name of the S3 bucket where your data will be stored] 

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files. 
$ export STORAGE_SIZE=[the size of the EBS volume that you are going to create, in GBs. e.g. "10"]

$ export AWS_REGION=[the AWS region of your Kubernetes cluster. e.g. "us-west-2"]

$ export AWS_AVAILABILITY_ZONE=[the AWS availability zone of your Kubernetes cluster. e.g. "us-west-2a"]

```

And then run:
```
$ aws s3api create-bucket --bucket ${BUCKET_NAME} --region ${AWS_REGION}

$ aws ec2 create-volume --size ${STORAGE_SIZE} --region ${AWS_REGION} --availability-zone ${AWS_AVAILABILITY_ZONE} --volume-type gp2
```

Record the "volume-id" that is output (e.g. "vol-8050b807"). You can also view it in the aws console or with  `aws ec2 describe-volumes`. Export the volume-id:

```shell
$ export STORAGE_NAME=[volume id]
```

Now you should be able to see the bucket and the EBS volume that are just created:

```shell
aws s3api list-buckets --query 'Buckets[].Name'
aws ec2 describe-volumes --query 'Volumes[].VolumeId'
```


### Install Pachctl 

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

### Start Pachyderm

First get a set of [temporary AWS credentials](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html) by using this command:

```shell
$ aws sts get-session-token
```

Then run the following commands with the credentials you get:

```shell
$ AWS_ID=[access key ID] 

$ AWS_KEY=[secret access key]

$ AWS_TOKEN=[session token] make amazon-cluster-manifest > manifest

$ make MANIFEST=manifest launch
```

It may take a while to complete for the first time, as a lot of Docker images need to be pulled.

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
