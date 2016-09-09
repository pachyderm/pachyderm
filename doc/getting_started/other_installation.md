# Other Installation Methods (k8s in Docker, from source)

Local Installation

## Prerequisites

- [Docker](https://docs.docker.com/engine/installation) >= 1.10
- [Kubectl (Kubernetes CLI)](#kubectl) >= 1.3.5
- [Pachyderm Command Line Interface](#pachctl)
- [FUSE (optional)](#fuse-optional) >= 2.8.2


TODO Last stable release
## OSX
### Kubectl


Make sure you have version 1.3.5 or higher.

```shell
# Darwin (OS X)
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.3.5/bin/darwin/amd64/kubectl
```
Note: If you don't have wget, just copy the link into your browser.


```sh
# Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
```
You can try running `kubectl version` to check that this worked correctly. 


### Pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


```shell
$ brew tap pachyderm/tap && brew install pachctl
```
You can try running `pachctl version` to check that this worked correctly. 

### Port Forwarding

Both kubectl and pachctl need a port forwarded so they can talk with their servers
If your Docker daemon is running locally you can skip this step.  Otherwise (e.g. you are running Docker through [Docker Machine](https://docs.docker.com/machine/)), do the following:


```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```


## Linux

```sh
# Linux
$ wget https://storage.googleapis.com/kubernetes-release/release/v1.3.5/bin/linux/amd64/kubectl
```

```sh
# Copy kubectl to your path
chmod +x kubectl
mv kubectl /usr/local/bin/
```
You can try running `kubectl version` to check that this worked correctly. 

---

### Pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.


If you're on linux 64 bit amd, you can use our pre-built deb package like so:

```shell
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

### From Source

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






### Kubernetes in Docker
- [Docker](https://docs.docker.com/engine/installation) >= 1.10





### Port Forwarding

Both kubectl and pachctl need a port forwarded so they can talk with their servers
If your Docker daemon is running locally you can skip this step.  Otherwise (e.g. you are running Docker through [Docker Machine](https://docs.docker.com/machine/)), do the following:


```shell
$ ssh <HOST> -fTNL 8080:localhost:8080 -L 30650:localhost:30650
```

### Deploy Kubernetes

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


## Installation

### Homebrew

```shell
$ brew tap pachyderm/tap && brew install pachctl
```

### Deb Package

If you're on linux 64 bit amd, you can use our pre-built deb package like so:

```shell
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```



## Usage

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

---

## Pachyderm

Even if you haven't installed pachctl from source, you'll need some make tasks located in the pachyderm repositoriy. If you haven't already cloned the repo, do so:

```shell
$ git clone git@github.com:pachyderm/pachyderm
```


### FUSE

Having FUSE installed allows you to mount PFS locally, which can be nice if you want to play around with PFS and is used in the beginner tutorial.

FUSE comes pre-installed on most Linux distributions.  For OS X, install [OS X FUSE](https://osxfuse.github.io/)


## Deploying Kubernetes


There are two easy ways to get kubernetes running locally. If you have Docker set up already, you can run kubernetes in Docker. If not, we recommend you use minikube.

## Install from source (absolute latest)

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

---

## Next

Now that you have everything installed and working, check out our [beginner_tutorial](LINK) to learn the basics of Pachyderm.

