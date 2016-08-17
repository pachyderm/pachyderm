# Deploying On the Cloud

## Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, Pachyderm can run on any platform that supports Kubernetes. This guide covers the following commonly used platforms:

* [Google Cloud Platform](#google-cloud-platform)
* [AWS](#amazon-web-services-aws)
* [OpenShift](#openshift)


## Google Cloud Platform


Google Cloud Platform has excellent support for Kubernetes through the [Google Container Engine](https://cloud.google.com/container-engine/).

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

- [AWS CLI](https://aws.amazon.com/cli/)

### Deploy Kubernetes

Deploying Kubernetes on AWS can be a little harder than GCE, but there are a few good tutorials that walk you through the process:

* CoreOS has a great tool called[kube-aws](https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html)

* If you don't want to use CoreOS, Kubernetes has their [own guides] (http://kubernetes.io/docs/getting-started-guides/aws/)

### Set up the infrastructure

Pachyderm needs a cluster (https://aws.amazon.com/documentation/ec2/), an [S3 bucket](https://aws.amazon.com/documentation/s3/), and a [persistent disk](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html)(EBS) to function correctly.  

We've made this very easy for you by creating the `make amazon-cluster` helper, which will create all of these resources for you.

First of all, set the required environment variables. Choose a name for both the bucket and disk, as well as a capacity for the disk (in GB):


```shell
$ export KUBECTLFLAGS="-s [the IP address of the node where Kubernetes runs]"
$ export BUCKET_NAME=[the name of the bucket where your data will be stored; this name needs to be unique across the entire AWS region]
$ export STORAGE_SIZE=[the size of the EBS volume that you are going to create, in GBs]
$ export AWS_REGION=[the AWS region where you want the bucket and EBS volume to reside]
$ export AWS_AVAILABILITY_ZONE=[the AWS availability zone where you want your EBS volume to reside]
```

Then, simply run:

```shell
$ make amazon-cluster
```

Record the "volume-id" in the output, then export it:

```shell
$ export STORAGE_NAME=[volume id]
```

Now you should be able to see the bucket and the EBS volume that are just created:

```shell
aws s3api list-buckets --query 'Buckets[].Name'
aws ec2 describe-volumes --query 'Volumes[].VolumeId'
```

### Deploy Pachyderm

First get a set of [temporary AWS credentials](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html):

```shell
$ aws sts get-session-token
```

Then run the following commands with the credentials you get:

```shell
$ AWS_ID=[access key ID] AWS_KEY=[secret access key] AWS_TOKEN=[session token] make amazon-cluster-manifest > manifest
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

