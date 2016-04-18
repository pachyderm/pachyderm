# Setup

* [Intro](#intro) 

## Intro
This document is about setting up and troubleshooting Pachyderm installations.
It's meant to cover known good configurations for a number of different platforms.
If you don't see your platform here please open and issue letting us know and
we'll help you find an install path and then document it here.

## Dependencies

- [Go](#go) >= 1.6
- [Docker](#docker) >= 1.10
- [Kubernetes](#kubernetes) and [Kubectl](#kubectl) >= 1.2.0
- [FUSE](#fuse) 2.8.2 (https://osxfuse.github.io/)

## Go
Find Go 1.6 [here](https://golang.org/doc/install).

## Docker

Docker has great docs for installing on any platform, check them out
[here](https://docs.docker.com/engine/installation/).

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
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/darwin/amd64/kubectl

### Linux
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.2.0/bin/linux/amd64/kubectl

### Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
```

## Pachctl
We assume you'll be compiling from within $GOPATH. So to install pachctl do:

```shell
$ go get github.com/pachyderm/pachyderm
$ cd $GOPATH/src/github.com/pachyderm/pachyderm
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

You'll know it works if `kubectl version` runs without error:

```shell
kubectl version
Client Version: version.Info{Major:"1", Minor:"1", GitVersion:"v1.2.0", GitCommit:"e4e6878293a339e4087dae684647c9e53f1cf9f0", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"1", GitVersion:"v1.2.0", GitCommit:"e4e6878293a339e4087dae684647c9e53f1cf9f0", GitTreeState:"clean"}
```

## Launch Pachyderm

From the root of this repo:

```shell
$ kubectl create -f http://pachyderm.io/manifest.json
```

Here's what a functioning cluster looks like:

```shell
$ kubectl get all
CONTROLLER   CONTAINER(S)   IMAGE(S)                               SELECTOR      REPLICAS   AGE
etcd         etcd           gcr.io/google_containers/etcd:2.0.12   app=etcd      1          2m
pachd        pachd          pachyderm/pachd                        app=pachd     1          1m
rethink      rethink        rethinkdb:2.1.5                        app=rethink   1          1m
NAME         CLUSTER_IP   EXTERNAL_IP   PORT(S)                        SELECTOR      AGE
etcd         10.0.0.197   <none>        2379/TCP,2380/TCP              app=etcd      2m
kubernetes   10.0.0.1     <none>        443/TCP                        <none>        10d
pachd        10.0.0.100   nodes         650/TCP,750/TCP                app=pachd     1m
rethink      10.0.0.218   <none>        8080/TCP,28015/TCP,29015/TCP   app=rethink   1m
NAME                   READY     STATUS    RESTARTS   AGE
etcd-4r4hp             1/1       Running   0          2m
k8s-master-127.0.0.1   3/3       Running   0          10d
pachd-u992h            1/1       Running   1          1m
rethink-268hq          1/1       Running   0          1m
NAME      LABELS    STATUS    VOLUME    CAPACITY   ACCESSMODES   AGE
```

You'll know it works if `pachctl version` runs without error:

```shell
pachctl version
COMPONENT           VERSION
pachctl             1.0.0
pachd               1.0.0
```

## Production Clusters
For production clusters you'll want to give Pachyderm access to object storage.
We currently only support s3, to launch a cluster with s3 credentials run:

```shell
$ pachctl manifest amazon-secret bucket id secret token region | kubectl create -f -
```

This encodes your credentials as Kubernetes secrets and gives pachd containers
access to them.


## Contributing

If you're interested in contributing, you'll need a bit more tooling setup. [Follow the instructions here](https://github.com/pachyderm/pachyderm/blob/master/contributing/setup.md)