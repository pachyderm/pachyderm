# Setup

## Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, technically Pachyderm can run on any platform that Kubernetes supports.  This guide covers the following commonly used platforms:

* [Local](#local-deployment)
* [Google Cloud Platform](#google-cloud-platform)
* [AWS](#amazon-web-services-aws)

Each section starts with deploying Kubernetes on the said platform, and then moves on to deploying Pachyderm on Kubernetes.  If you have already set up Kubernetes on your platform, you may directly skip to the second part.

## Common Prerequisites

- [Go](#go) >= 1.6
- [FUSE (optional)](#fuse-optional) >= 2.8.2
- [Kubectl (kubernetes CLI)](#kubectl) >= 1.2.2
- [Pachyderm Repository](#pachyderm)

### Go

Find Go 1.6 [here](https://golang.org/doc/install).

### FUSE (optional)

Having FUSE installed allows you to mount PFS locally, which can be nice if you want to play around with PFS.

FUSE comes pre-installed on most Linux distributions.  For OS X, install [OS X FUSE](https://osxfuse.github.io/)

### Kubectl

Make sure you have version 1.2.2 or higher.

```shell
### Darwin (OS X)
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.2/bin/darwin/amd64/kubectl

### Linux
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.2/bin/linux/amd64/kubectl

### Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
```

### Pachyderm

Clone this repo under your `GOPATH`:

```shell
# this will put the repo under $GOPATH/src/github.com/pachyderm/pachyderm
$ go get github.com/pachyderm/pachyderm
```


## Local Deployment

### Prerequisites

- [Docker](https://docs.docker.com/engine/installation) >= 1.10

### Port Forwarding

Both kubectl and pachctl need a port forwarded so they can talk with their servers.  If your Docker daemon is running locally you can skip this step.  Otherwise (e.g. you are running Docker through [Docker Machine](https://docs.docker.com/machine/)), do the following:


```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

### Deploy Kubernetes

From the root of this repo you can deploy Kubernetes with:

```shell
$ make launch-kube
```

This step can take a while the first time you run it, since some Docker images need to be pulled. 

### Deploy Pachyderm

From the root of this repo you can deploy Pachyderm on Kubernetes with:

```shell
$ make launch
```

This step can take a while the first time you run it, since a lot of Docker images need to be pulled. 

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

Pachyderm needs a [container cluster](https://cloud.google.com/container-engine/), a [GCS bucket](https://cloud.google.com/storage/docs/), and a [persistent disk](https://cloud.google.com/compute/docs/disks/) to function correctly.  We've made this very easy for you.

First of all, set three environment variables:

```shell
$ export BUCKET_NAME=[the name of the bucket where your data will be stored; this name needs to be unique across the entire Google Cloud Storage namespace]
$ export STORAGE_NAME=[the name of the persistent disk where your pipeline information will be stored]
$ export STORAGE_SIZE=[the size of the persistent disk that you are going to create, in GBs]
```

Then, simply run:

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

### Format Volume

Unfortunately, your persistent disk is not immediately available for use upon creation.  You will need to manually format it.  Follow [these instructions](https://cloud.google.com/compute/docs/disks/add-persistent-disk#formatting), then clear all files on the disk by:

```shell
rm -rf [path-to-disk]/*
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

This is so we can use [`pachctl`](#pachctl) to talk to our cluster later.

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

Deploying Kubernetes on AWS is still a relatively lengthy and manual process comparing to doing it on GCE.  However, here are a few good tutorials that walk through the process:

* http://kubernetes.io/docs/getting-started-guides/aws/
* https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html

### Set up the infrastructure

First of all, set these environment variables:

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

### Format Volume

Unfortunately, your EBS volume is not immediately available for use upon creation.  You will need to manually format it.  Follow [these instructions](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-using-volumes.html), then clear all files on the volume by:

```shell
rm -rf [path-to-disk]/*
```

### Deploy Pachyderm

First of all, get a set of [temporary AWS credentials](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp.html):

```shell
$ aws sts get-session-token
```

Then run the following commands with the credentials you get:

```shell
$ AWS_ID=[access key ID] AWS_KEY=[secret access key] AWS_TOKEN=[session token] make amazon-cluster-manifest > manifest
$ make MANIFEST=manifest launch
```

It may take a while to complete for the first time, as a lot of Docker images need to be pulled.

## pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

### Installation

#### Homebrew

```shell
$brew tap pachyderm/tap && brew install pachctl
```

#### From Source

To install pachctl from source, we assume you'll be compiling from within $GOPATH. So to install pachctl do:

```shell
$ go get github.com/pachyderm/pachyderm
$ cd $GOPATH/src/github.com/pachyderm/pachyderm
$ make install
```

Make sure you add `GOPATH/bin` to your `PATH` env variable:

```shell
$ export PATH=$PATH:$GOPATH/bin
```

### Usage

If Pachyderm is running locally, you are good to go.  Otherwise, you need to make sure that `pachctl` can find the node on which you deployed Pachyderm:

```shell
$ export ADDRESS=[the IP address of the node where Pachyderm runs]:30650
# for example:
# export ADDRESS=104.197.179.185:30650
```

Now, create an empty repo to make sure that everything has been set up correctly:

```shell
pachctl create-repo test
pachctl list-repo
# should see "test"
```

## Next Step

Ready to jump into data analytics with Pachyderm?  Head to our [quick start guide](examples/fruit_stand/README.md).
