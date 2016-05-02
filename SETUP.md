# Setup

## Intro

Pachyderm is built on [Kubernetes](http://kubernetes.io/).  As such, technically Pachyderm can run on any platform that Kubernetes supports.  This guide covers the following commonly used platforms:

* [Local](#local-deployment)
* [Google Cloud Platform](#google-cloud-playform)
* [AWS](#amazon-web-services-aws)

Each section starts with deploying Kubernetes on the said platform, and then moves on to deploying Pachyderm on Kubernetes.  If you have already set up Kubernetes on your platform, you may directly skip to the second part.

## Common Prerequisites

- [Go](#go) >= 1.6
- [FUSE (optional)](#fuse-optional) >= 2.8.2

### Go

Find Go 1.6 [here](https://golang.org/doc/install).

### FUSE (optional)

Having FUSE installed allows you to mount PFS locally, which can be nice if you want to play around with PFS.

FUSE comes pre-installed on most Linux distributions.  For OS X, install [OS X FUSE](https://osxfuse.github.io/)

### Kubectl

```shell
### Darwin
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/darwin/amd64/kubectl

### Linux
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/linux/amd64/kubectl

### Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
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
$ export CLUSTER_NAME=[the name of your Kubernetes cluster]
$ export BUCKET_NAME=[the name of the bucket where your data will be stored]
$ export STORAGE_NAME=[the name of the persistent disk where your pipeline information will be stored]
$ export STORAGE_SIZE=[the size of the persistent disk that you are going to create]
```

Then, simply run:

```shell
$ make google-cluster
```

This creates a Kubernetes cluster, a bucket, and a persistent disk.  To check that everything has been set up correctly, try:

```shell
$ gcloud compute instances list
# should see a number of instances

$ gsutil ls
# should see a bucket

$ gcloud compute disks list
# should see a number of disks, including the one you specified
```

### Deploy Pachyderm

```shell
$ make google-cluster-manifest > manifest
$ MANIFEST=manifest make launch
```

## Amazon Web Services (AWS)

### Prerequisites

- [AWS CLI](https://aws.amazon.com/cli/) 

### Deploy Kubernetes

Deploying Kubernetes on AWS is still a relatively lengthy and manual process comparing to doing it on GCE.  However, here are a few good tutorials that walk through the process:

* https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html
* http://kubernetes.io/docs/getting-started-guides/aws/

## Set up the infrstructure

First of all, set three environment variables:

```shell
$ export BUCKET_NAME=[the name of the bucket where your data will be stored]
$ export STORAGE_SIZE=[the size of the EBS volume that you are going to create]
$ export AWS_REGION=[the AWS region where you want the bucket and EBS volume to reside]
```

Then, simply run:

```shell
$ make amazon-cluster
```

Record the "volume-id" in the output:

```shell
$ export STORAGE_NAME=[volume id]
```

Now you should be able to see the bucket and the EBS volume that are just created:

```shell
aws s3api list-buckets --query 'Buckets[].Name'
aws ec2 describe-volumes --query 'Volumes[].VolumeId'
```

### Deploy Pachyderm

```shell
$ AWS_ID=[access key ID] AWS_KEY=[secret access key] AWS_TOKEN=[security token] make amazon-cluster-manifest > manifest
$ MANIFEST=manifest make launch
```

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

TODO

## Contributing

If you're interested in contributing, you'll need a bit more tooling setup. [Follow the instructions here](https://github.com/pachyderm/pachyderm/blob/master/contributing/setup.md)
