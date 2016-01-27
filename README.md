# Pachyderm
[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/github/license/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)

* [News](#news)
* [What is Pachyderm?](#what-is-pachyderm)
* [Key Features](#key-features)
* [Is Pachyderm enterprise production ready?](#is-pachyderm-enterprise-production-ready)
* [What is a commit-based file system?](#what-is-a-commit-based-file-system)
* [What are containerized analytics?](#what-are-containerized-analytics)
* [Using Pachyderm](#using-pachyderm)
    * [Prerequisites](#prerequisites)
    * [Launch a Development Cluster](#launch-a-development-cluster)
    * [Launch a Production Cluster](#launch-a-production-cluster)
    * [Common Problems](#common-problems)
* [Environment Setup](#environment-setup)
    * [Go Setup](#go-setup)
    * [Docker Setup](#docker-setup)
    * [Vagrant](#vagrant)
* [Contributing](#contributing)

### What is Pachyderm?

Pachyderm is a Data Lake. A place to dump and process gigantic data sets.
Pachyderm is inspired by the Hadoop ecosystem but _shares no code_ with it.
Instead we leverage the container ecosystem to provide the broad functionality
of Hadoop with the ease of use of Docker.

Pachyderm offers the following broad functionality:

- Virtually limitless storage for any data.
- Virtually limitless processing power using any tools.
- Tracking of data history, provenance and ownership. (Version Control).
- Automatic processing on new data as it’s ingested. (Streaming).
- Chaining processes together. (Pipelining)

### What's new about Pachyderm? (How is it different from Hadoop?)

There are two bold new ideas in Pachyderm:

- Containers as the processing payload
- Version Control for data

These ideas lead directly to a system that's much easier to use and administer.

To process data you simply create a containerized program which reads and writes
to the local filesystem. Pachyderm will take your container and inject data into
it by way of a FUSE volume. You can use _any_ tools you want! Pachyderm will
automatically replicate your container. It creates multiple copies of the same
container showing each one a different chunk of data in the FUSE volume. With
this technique Pachyderm can scale any code you write up to petabytes of data.

Pachyderm also version controls all data, it's very similar to how git handles
source code. 

#### Prerequisites

Requirements:
- Go 1.5
- Docker 1.9

[More Info](#environment-setup)

#### Launch a Development Cluster
To start a development cluster run:

```shell
make launch
```

This will compile the code on your local machine and launch it as a docker-compose service.
A succesful launch looks like this:

```shell
docker-compose ps
        Name                       Command               State                                 Ports
-----------------------------------------------------------------------------------------------------------------------------------
pachyderm_btrfs_1       sh entrypoint.sh                 Up
pachyderm_etcd_1        /etcd -advertise-client-ur ...   Up      0.0.0.0:2379->2379/tcp, 2380/tcp, 4001/tcp, 7001/tcp
pachyderm_pfs-roler_1   /pfs-roler                       Up
pachyderm_pfsd_1        sh btrfs-mount.sh /pfsd          Up      0.0.0.0:1050->1050/tcp, 0.0.0.0:650->650/tcp, 0.0.0.0:750->750/tcp
pachyderm_ppsd_1        /ppsd                            Up      0.0.0.0:1051->1051/tcp, 0.0.0.0:651->651/tcp
pachyderm_rethink_1     rethinkdb --bind all             Up      28015/tcp, 29015/tcp, 8080/tcp
```

#### Pachyderm CLI
Pachyderm has a CLI called `pach`. To install it:

```shell
make install
```

`pach` should be able to access dev clusters without any additional setup.

#### Launch a Production Cluster
Before you can launch a production cluster you'll need a working Kubernetes deployment.
You can start one locally on Docker using:

```shell
etc/kube/start-kube-docker.sh
```

You can then deploy a Pachyderm cluster on Kubernetes with:

```shell
pachctl create-cluster -n test-cluster -s 1
```

### Environment Setup

#### Go Setup
With golang, it's generally easiest to have your fork match the import paths in the code. We recommend you do it like this:

```
# assuming your github username is alice
rm -rf ${GOPATH}/src/github.com/pachyderm/pachyderm
mkdir -p ${GOPATH}/src/github.com/pachyderm
cd ${GOPATH}/src/github.com/pachyderm
git clone https://github.com/alice/pachyderm.git
cd pachyderm
git remote add upstream https://github.com/pachyderm/pachyderm.git # so you can run 'git fetch upstream' to get upstream changes
```

#### Docker Setup

If you're on a Mac or Windows, easiest way to get up and running is the
[Docker toolbox](https://www.docker.com/docker-toolbox). Linux users should
follow [this guide](http://docs.docker.com/engine/installation/ubuntulinux/).

#### Vagrant

The [Vagrantfile](etc/initdev/Vagrantfile) in this repository will set up a development environment for Pachyderm
that has all dependencies installed.

The easiest way to install Vagrant on your mac is probably:

```
brew install caskroom/cask/brew-cask
brew cask install virtualbox vagrant
```

Basic usage:

```
mkdir -p pachyderm_vagrant
cd pachyderm_vagrant
curl https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/initdev/Vagrantfile > Vagrantfile
curl https://raw.githubusercontent.com/pachyderm/pachyderm/master/etc/initdev/init.sh > init.sh
vagrant up # starts the vagrant box
vagrant ssh # ssh into the vagrant box
```

Once in the vagrant box, set everything up and verify that it works:

```
go get github.com/pachyderm/pachyderm/...
cd ~/go/src/github.com/pachyderm/pachyderm
make test
```

Some other useful vagrant commands:

```
vagrant suspend # suspends the vagrant box, useful if you are not actively developing and want to free up resources
vagrant resume # resumes a suspended vagrant box
vagrant destroy # destroy the vagrant box, this will destroy everything on the box so be careful
```

See [Vagrant's website](https://www.vagrantup.com) for more details.

### Common Problems

*Problem*: Nothing is running after launch.

- Check to make sure the docker daemon is running with `ps -ef | grep docker`.
- Check to see if the container exited with `docker ps -a | grep IMAGE_NAME`.
- Check the container logs with `docker logs`.

*Problem*: Docker commands are failing with permission denied

The bin scripts assume you have your user in the docker group as explained in the [Docker Ubuntu installation docs](https://docs.docker.com/installation/ubuntulinux/#create-a-docker-group).
If this is set up properly, you do not need to use `sudo` to run `docker`. If you do not want this, and want to have to use `sudo` for docker development, wrap all commands like so:

```
sudo -E bash -c 'make test' # original command would have been `make test`
```

### Contributing

To get started, sign the [Contributor License Agreement](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement).

Send us PRs, we would love to see what you do!

### News

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our team](http://www.pachyderm.io/jobs.html) and email us at jobs@pachyderm.io.

