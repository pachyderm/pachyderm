
If you're evaluating Pachyderm and want the quickest hosting option, we recommend using Google's GCE. You already have credentials (if you have a gmail account), and our setup script is quick.

# Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, technically Pachyderm can run on any platform that Kubernetes supports.  This guide covers the following commonly used platforms:

* [Local](#local-deployment)
* [Google Cloud Platform](#google-cloud-platform)
* [AWS](#amazon-web-services-aws)
* [OpenShift](#openshift)

# Local Deployment

If you have docker running, you can run kubernetes right off of docker. If you don't, consider checking out:

- minikube (TODO / link)
- or use one of the [Deployment Options](./deploying.md) to host the Docker daemon.

## Prerequisites

- [Docker](https://docs.docker.com/engine/installation) >= 1.10

## Port Forwarding

Both kubectl and pachctl need a port forwarded so they can talk with their servers.  If your Docker daemon is running locally you can skip this step.  Otherwise (e.g. you are running Docker through [Docker Machine](https://docs.docker.com/machine/)), do the following:


```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

## Deploy Kubernetes

From the root of this repo you can deploy Kubernetes with:

```shell
$ make launch-kube
```

This step can take a while the first time you run it, since some Docker images need to be pulled.

## Deploy Pachyderm

From the root of this repo you can deploy Pachyderm on Kubernetes with:

```shell
$ make launch
```

This step can take a while the first time you run it, since a lot of Docker images need to be pulled.


# Google Cloud Platform

Google Cloud Platform has excellent support for Kubernetes through the [Google Container Engine](https://cloud.google.com/container-engine/).

## Prerequisites

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 106.0.0

If this is the first time you use the SDK, make sure to follow through the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts).

After the SDK is installed, run:

```shell
$ gcloud components install kubectl
```

## Set up the infrastructure

Pachyderm needs a [container cluster](https://cloud.google.com/container-engine/), a [GCS bucket](https://cloud.google.com/storage/docs/), and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.  We've made this very easy for you by creating the `make google-cluster` helper, which will create all of these resources for you.

First of all, set the required environment variables. Choose a name for both the bucket and disk, as well as a capacity for the disk (in GB):

```shell
$ export BUCKET_NAME=some-unique-bucket-name
$ export STORAGE_NAME=pach-disk
$ export STORAGE_SIZE=200
```

You may need to visit the [Console] to fully initialize Container Engine in a new project. Then, simply run the following command:

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

## Deploy Pachyderm

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

This is so we can use [`pachctl`](#pachctl) to talk to our cluster later.

Now you can deploy Pachyderm with:

```shell
$ make google-cluster-manifest > manifest
$ make MANIFEST=manifest launch
```

It may take a while to complete for the first time, as a lot of Docker images need to be pulled.

# Amazon Web Services (AWS)

## Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/)

## Deploy Kubernetes

Deploying Kubernetes on AWS is still a relatively lengthy and manual process comparing to doing it on GCE.  However, here are a few good tutorials that walk through the process:

* http://kubernetes.io/docs/getting-started-guides/aws/
* https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html

## Set up the infrastructure

First of all, set these environment variables:

```shell
$ export KUBECTLFLAGS="-s [the Public IP address of the node where Kubernetes-master runs]"
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

## Deploy Pachyderm

First of all, get a set of [temporary AWS credentials](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html):

```shell
$ aws sts get-session-token
```

Then run the following commands with the credentials you get:

```shell
$ AWS_ID=[access key ID] AWS_KEY=[secret access key] AWS_TOKEN=[session token] make amazon-cluster-manifest > manifest
$ make MANIFEST=manifest launch
```

It may take a while to complete for the first time, as a lot of Docker images need to be pulled. User can tear down the Pachyderm service by running:

```shell
$ make amazon-clean
```

It may take about 30 seconds to delete the pachyderm components.

# OpenShift

[OpenShift](https://www.openshift.com/) is a popular enterprise Kubernetes distribution.  Pachyderm can run on OpenShift with two additional steps:

1. Make sure that priviledge containers are allowed (they are not allowed by default):  `oc edit scc` and set `allowPrivilegedContainer: true` everywhere.
2. Remove `hostPath` everywhere from your cluster manifest (e.g. `etc/kube/pachyderm-versioned.json` if you are deploying locally).

Problems related to OpenShift deployment are tracked in this issue: https://github.com/pachyderm/pachyderm/issues/336

# Usage Metrics

Pachyderm automatically reports anonymized usage metrics. These metrics help us
understand how people are using Pachyderm and make it better.  They can be
disabled by setting the env variable `METRICS` to `false` in the pachd
container.

# Pachctl

Instructions for setting up pachctl can be found [here](http://pachyderm.readthedocs.io/en/latest/installation.html#)

# Next

Ready to jump into data analytics with Pachyderm?  Head to our [quick start guide](../examples/fruit_stand/README.md).

