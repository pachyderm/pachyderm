# Setup

* [Intro](#intro) 

## Intro
This document is about setting up and troubleshooting Pachyderm installations.
It's meant to cover known good configurations for a number of different platforms.
If you don't see your platform here please open and issue letting us know and
we'll help you find an install path and then document it here.

## Dependencies

- [Go](#go) >= 1.5
- [Docker](#docker) >= 1.9 (must deploy with [`--storage-driver=devicemapper`](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/), supported as described [here](https://docs.docker.com/engine/userguide/storagedriver/device-mapper-driver/))
- [Kubernetes](#kubernetes) and [Kubectl](#kubectl) >= 1.1.7
- [FUSE](#fuse) 2.8.2 (https://osxfuse.github.io/)

## Go
Find the latest version of Go [here](https://golang.org/doc/install).

## Docker

Docker has great docs for installing on any platform, check them out
[here](https://docs.docker.com/engine/installation/).

There is one Pachyderm specific wrinkle, we're not compatible with aufs, we've
found device mapper to be the best backend for Pachyderm. Using device mapper
is should be as simple as passing `--storage-driver=devicemapper` to your
docker invocation. Or adding that same line to `DOCKER_OPTS`. This [blog
post](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/)
discusses more.

## Kubernetes

Kubernetes can be installed in many different ways, if you're looking to setup
a development cluster we recommend running [Dockerized Kubernetes](#dockerized-kubernetes).
Otherwise you should follow instructions specific to your cloud provider.

### Dockerized Kubernetes

From the root of this repo you can deploy Kubernetes with:

```shell
$ make launch-kube
```

### Amazon

CoreOS has [great instructions](https://coreos.com/kubernetes/docs/latest/kubernetes-on-aws.html)
on how to get on how to get Kubernetes working on AWS.

### Google

Google has the best support for Kubernetes, they wrote it, through [Google
Container Engine](https://cloud.google.com/container-engine/).

### Microsoft

Kubernetes on Azure seems to be the least well tested of all the options. This
is the [best guide](https://github.com/kubernetes/kubernetes/blob/master/docs/getting-started-guides/coreos/azure/README.md)
we've found.

## Kubectl

```shell
### Darwin
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.1.7/bin/darwin/amd64/kubectl

### Linux
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.1.7/bin/linux/amd64/kubectl

### Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
```

## Pachctl
From the root of this repo you can install pachctl with:

```shell
$ make install
```

Make sure you add `GOPATH/bin` to your `PATH` env variable:

```shell
$ export PATH=$PATH:$GOPATH/bin
```

## Port Forwarding
Both kubectl and pachctl need a port forwarded so they can talk with their servers.
If docker is running locally you can skip this step. Otherwise do the following:

```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

