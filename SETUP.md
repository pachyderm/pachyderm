# Setup

* [Intro](#intro) 

## Intro

This document is about setting up and troubleshooting Pachyderm installations.
It's meant to cover known good configurations for a number of different platforms.
If you don't see your platform here please open and issue letting us know and
we'll help you find an install path and then document it here.

## Common Prerequisites

- [Go](#go) >= 1.6
- [FUSE (optional)](#fuse) >= 2.8.2 ()

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

### Launch Kubernetes

From the root of this repo you can deploy Kubernetes with:

```shell
$ make launch-kube
```

This step can take a while the first time you run it, since some Docker images need to be pulled. 

### Launch Pachyderm

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
$ make cluster
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

### Launch Pachyderm

```shell
$ make google-cluster-manifest > manifest
$ MANIFEST=manifest make launch
```

## Amazon Web Services (AWS)

TODO

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

### Port Forwarding

(TODO: is this still relevant?)

Both kubectl and pachctl need a port forwarded so they can talk with their servers.
If docker is running locally you can skip this step. Otherwise do the following:

```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

You'll know it works if `kubectl version` runs without error:

```shell
kubectl version
Client Version: version.Info{Major:"1", Minor:"1", GitVersion:"v1.2.0", GitCommit:"e4e6878293a339e4087dae684647c9e53f1cf9f0", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"1", GitVersion:"v1.2.0", GitCommit:"e4e6878293a339e4087dae684647c9e53f1cf9f0", GitTreeState:"clean"}
```

### Usage

TODO

## Contributing

If you're interested in contributing, you'll need a bit more tooling setup. [Follow the instructions here](https://github.com/pachyderm/pachyderm/blob/master/contributing/setup.md)
