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

- Containers as the processing primitive
- Version Control for data

These ideas lead directly to a system that's much easier to use and administer.

To process data you simply create a containerized program which reads and writes
to the local filesystem. Pachyderm will take your container and inject data into
it by way of a FUSE volume. You can use _any_ tools you want! Pachyderm will
automatically replicate your container. It creates multiple copies of the same
container showing each one a different chunk of data in the FUSE volume. With
this technique Pachyderm can scale any code you write up to petabytes of data.

Pachyderm also version controls all data using a commit based distributed
filesystem (PFS), it's very similar to what git does with code. Version control
has far reaching consequences in a distributed filesystem. You get the full
history of your data, it's much easier to collaborate with teammates and if
anythng goes wrong you can revert _the entire cluster_ with one click!

Version control is also very synergistic with our containerized processing
engine. Pachyderm understands how your data changes and thus, as new data
is ingested, can run your workload on the _diff_ of the data rather than the
whole thing. This means that there's no difference between a batched job and
a streaming job, the same code will work for both!

### Quickstart

#### Prerequisites
- Docker >= 1.9 (must deploy with [`--storage-driver=devicemapper`](http://muehe.org/posts/switching-docker-from-aufs-to-devicemapper/))
- Go >= 1.5
- Kubernetes and Kubectl >= 1.1.2
- FUSE 2.8.2 (https://osxfuse.github.io/)

#### Create a Pachyderm Cluster

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

