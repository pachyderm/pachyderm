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

### News

WE'RE HIRING! Love Docker, Go and distributed systems? Learn more about [our team](http://www.pachyderm.io/jobs.html) and email us at jobs@pachyderm.io.

### What is Pachyderm?

Pachyderm is a complete data analytics solution that lets you efficiently store and analyze your data using containers. We offer the scalability and broad functionality of Hadoop, with the ease of use of Docker.

### Key Features

- Complete version control for your data
- Pipelines are containerized, so you can use any languages and tools you want
- Both batched and streaming analytics
- One-click deploy on AWS without data migration

### Is Pachyderm enterprise production ready?

No, Pachyderm is in beta, but can already solve some very meaningful data analytics problems.  [We'd love your help. :)](#development)

### What is a commit-based file system?

Pfs is implemented as a distributed layer on top of btrfs, the same
copy-on-write file system that powers Docker. Btrfs already offers
[git-like semantics](http://zef.me/6023/who-needs-git-when-you-got-zfs/) on a
single machine; pfs scales these out to an entire cluster. This allows features such as:
- __Commit-based history__: File systems are generally single-state entities. Pfs,
on the other hand, provides a rich history of every previous state of your
cluster. You can always revert to a prior commit in the event of a
disaster.
- __Branching__: Thanks to btrfs's copy-on-write semantics, branching is ridiculously
cheap in pfs. Each user can experiment freely in their own branch without
impacting anyone else or the underlying data. Branches can easily be merged back in the main cluster.
- __Cloning__: Btrfs's send/receive functionality allows pfs to efficiently copy
an entire cluster's worth of data while still maintaining its commit history.

### What are containerized analytics?

Rather than thinking in terms of map or reduce jobs, pps thinks in terms of pipelines expressed within a container. A pipeline is a generic way expressing computation over large datasets and it’s containerized to make it easily portable, isolated, and easy to monitor. In Pachyderm, all analysis runs in containers. You can write them in any language you want and include any libraries.

### Using Pachyderm

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
