# Pachyderm
![Build status](https://badge.buildkite.com/69c8b8a239b2796b43520ddff1576d8bffeb85fc86adad99d2.svg?branch=master)

[![GitHub release](https://img.shields.io/github/release/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/releases)
[![GitHub license](https://img.shields.io/github/license/pachyderm/pachyderm.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE)

* [News](#news)
* [What is Pachyderm?](#what-is-pachyderm)
* [Key Features](#key-features)
* [Is Pachyderm enterprise production ready?](#is-pachyderm-enterprise-production-ready)
* [What is a commit-based file system?](#what-is-a-commit-based-file-system)
* [What are containerized analytics?](#what-are-containerized-analytics)
* [Development](#development)
    * [Running](#running)
    * [Development Notes](#development-notes)
      * [Logs](#logs)
    * [Environment Setup](#environment-setup)
    * [Common Problems](#common-problems)
* [Contributing](#contributing)

### News

Note the custom golang import path! `go get go.pachyderm.com/pachyderm`.

We are in the midst of a refactor! See the release branch for the current, working release of Pachyderm.

Check out our docker volume driver! https://github.com/pachyderm/pachyderm/tree/master/src/cmd/pfs-volume-driver.

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

## Development

We're hiring! If you like ambitious distributed systems problems and think there should be a better alternative to Hadoop, please reach out. Email jobs@pachyderm.io.

### Running

You need to install docker-compose for the Makefile commands to work.

```shell
curl -L https://github.com/docker/compose/releases/download/1.4.0rc2/docker-compose-$(uname -s)-$(uname -m) > /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

You need to have Go 1.5 installed and have `GO15VENDOREXPERIMENT=1`.

Useful development commands can be seen in the [Makefile](Makefile). Key commands:

```
make test-deps # download all golang dependencies
make build # build the source code (does not build the tests)
make test # run all the tests
make clean # clean up all pachyderm state
RUNARGS="go test -test.v ./..." make run # equivalent to TESTFLAGS=-test.v make test
make launch-pfsd # launch the new pfsd daemon
make install # install all binaries locally
pfs # if ${GOPATH}/bin is on your path, this will run the new pfs cli, this is very experimental and does not check for common errors
```

### Development Notes

##### Logs

We're using [protolog](http://go.pedge.io/protolog) for logging. All new log events should be wrapped in a protobuf message.
A package that has log messages should have a proto file named `protolog.proto` in it.
See [src/pps/run/protolog.proto](src/pps/run/protolog.proto) and [src/pps/run/runner.go](src/pps/run/runner.go) for an example.

### Environment Setup

With golang, it's generally easiest to have your fork match the import paths in the code. We recommend you do it like this:

```
# assuming your github username is alice
rm -rf ${GOPATH}/src/go.pachyderm.com/pachyderm
cd ${GOPATH}/src/go.pachyderm.com
git clone https://github.com/alice/pachyderm.git
git remote add upstream https://github.com/pachyderm/pachyderm.git # so you can run 'git fetch upstream' to get upstream changes
```

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
go get go.pachyderm.com/pachyderm/...
cd ~/go/src/go.pachyderm.com/pachyderm
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
sudo -E bash -c 'bin/run go test ./...' # original command would have been `./bin/run go test ./...`
```

## Contributing

To get started, sign the [Contributor License Agreement](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement).

Send us PRs, we would love to see what you do!
