# Installation

## Common Prerequisites

- [FUSE (optional)](#fuse-optional) >= 2.8.2
- [Kubectl (kubernetes CLI)](#kubectl) >= 1.2.2
- [pachctl Command Line Interface](#pachctl)
- [Pachyderm Repository](#pachyderm)

## FUSE (optional)

Having FUSE installed allows you to mount PFS locally, which can be nice if you want to play around with PFS.

FUSE comes pre-installed on most Linux distributions.  For OS X, install [OS X FUSE](https://osxfuse.github.io/)

---

## Kubectl

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

---

## pachctl

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

### Installation

#### Homebrew

```shell
$ brew tap pachyderm/tap && brew install pachctl
```

#### Deb Package

If you're on linux 64 bit amd, you can use our pre-built deb package like so:

```shell
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
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

---

## Pachyderm

Even if you haven't installed pachctl from source, you'll need some make tasks located in the pachyderm repositoriy. If you haven't already cloned the repo, do so:

```shell
$ git clone git@github.com:pachyderm/pachyderm
```

---

## Next

Now that you have the command line interface installed, [deploy your pachyderm cluster](./deploying.html)
