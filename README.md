# Pachyderm

[![Apache 2.0 License](https://img.shields.io/badge/license-Apache-2.0.svg?style=flat-square)](https://github.com/pachyderm/pachyderm/blob/master/LICENSE.md)

* [News](#news)
* [What is Pachyderm?](#what-is-pachyderm)
* [Key Features](#key-features)
* [Is Pachyderm enterprise production ready?](#is-pachyderm-enterprise-production-ready)
* [What is a commit-based file system?](#what-is-a-commit-based-file-system)
* [What are containerized analytics?](#what-are-containerized-analytics)
* [Documentation](#documentation)
  * [Deploying a Pachyderm cluster](#deploying-a-pachyderm-cluster)
    * [Deploy Pachyderm manually](#deploy-pachyderm-manually)
    * [Settings](#settings)
    * [Integrating with s3](#integrating-with-s3)
    * [Checking the status of your deploy](#checking-the-status-of-your-deploy)
  * [The Pachyderm HTTP API](#the-pachyderm-http-api)
    * [Creating files](#creating-files)
    * [Reading files](#reading-files)
    * [Deleting files](#deleting-files)
    * [Committing changes](#committing-changes)
    * [Branching](#branching)
  * [Containerized Analytics](#containerized-analytics)
    * [Creating a new pipeline with a Pachfile](#creating-a-new-pipeline-with-a-pachfile)
    * [POSTing a Pachfile to pfs](#posting-a-pachfile-to-pfs)
    * [Running a pipeline](#running-a-pipeline)
    * [Getting the output of a pipelines](#getting-the-output-of-a-pipelines)
    * [Deleting pipelines](#deleting-pipelines)
    * [Getting the Pachfile](#getting-the-pachfile)
* [Development](#development)
    * [Running](#running)
    * [Environment Setup](#environment-setup)
    * [Common Problems](#common-problems)
* [Contributing](#contributing)

### News

Pachyderm v0.8 is out and includes a brand new pipelining system! [Read more](http://pachyderm.io/pps.html) about it or check out our [web scraper demo](https://medium.com/pachyderm-data/build-your-own-wayback-machine-in-10-lines-of-code-99884b2ff95c).

We are in the midst of a refactor! More news to follow, but new code is being committed along-side existing code and is contained primarily within [src/pfs](src/pfs).

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

## Documentation

### Deploying a Pachyderm cluster

Pachyderm is designed to run on CoreOS so we'll need to deploy a CoreOs cluster. We've created an AWS cloud template to make this insanely easy.
- [Deploy on Amazon EC2](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=Pachyderm&templateURL=https:%2F%2Fs3-us-west-1.amazonaws.com%2Fpachyderm-templates%2Ftemplate) using cloud templates (recommended)
- [Amazon EC2](https://coreos.com/docs/running-coreos/cloud-providers/ec2/) (manual)
- [Google Compute Engine](https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/) (manual)
- [Vagrant](https://coreos.com/docs/running-coreos/platforms/vagrant/) (requires setting up DNS)

#### Deploy Pachyderm manually

If you chose any of the manual options above, you'll neeed to SSH in to one of your new CoreOS machines and start Pachyderm.

```shell
$ wget -qO- pachyderm.io/deploy | sh
```
The startup process takes a little while the first time you run it because
each node has to pull a Docker image.

#### Deploy Pachyderm using Kubernetes

As of v0.9 Pachyderm supports Deploying on Kubernetes.  The relevant files can
be found in [etc/kube](etc/kube). The start the service running on
Kubernetes do:

```shell
$ kubectl create -f etc/kube/storage-controller.yml
$ kubectl create -f etc/kube/router-controller.yml
$ kubectl create -f etc/kube/pachyderm-service.yml
```

Or just do:

```shell
make kube-create
```

Pachyderm expects etcd to be running on the host machine. 

See [Kubernetes' Getting Started Guide](http://kubernetes.io/gettingstarted) for how to deploy Kubernetes to various platforms.
If you are developing on linux and want to test locally, here's a cheat sheet to local deployment.
This won't be generally updated, it's just how we got a local cluster on our linux boxes, and
some of the commands/functions may not apply if you do not use bash as your shell, so just
use this as a reference:

```shell
mkdir -p ~/git # or wherever you clone to
mkdir -p ~/other # a place to download etcd to

# https://github.com/coreos/etcd/releases
cd ~/other
curl -L  https://github.com/coreos/etcd/releases/download/v2.1.0-rc.0/etcd-v2.1.0-rc.0-linux-amd64.tar.gz -o etcd-v2.1.0-rc.0-linux-amd64.tar.gz
tar xzvf etcd-v2.1.0-rc.0-linux-amd64.tar.gz

# http://kubernetes.io/v1.0/docs/getting-started-guides/locally.html
cd ~/git
git clone https://github.com/GoogleCloudPlatform/kubernetes.git

# in your bash_aliases, put the following:
# export PATH=${PATH}:~/other/etcd-v2.1.0-rc.0-linux-amd64:~/git/kubernetes/cluster
# kubernetes_up() {
#   ~/git/kubernetes/hack/local-up-cluster.sh
#   ~/git/kubernetes/cluster/kubectl.sh config set-cluster local --server=http://127.0.0.1:8080 --insecure-skip-tls-verify=true
#   ~/git/kubernetes/cluster/kubectl.sh config set-context local --cluster=local
#   ~/git/kubernetes/cluster/kubectl.sh config use-context local
# }

# in a separate terminal:
source ~/.bashrc
kubernetes_up

# in your main terminal:
cd ${GOPATH}/src/github.com/pachyderm/pachyderm
make kube-create
kubectl.sh get pods
```

####  Settings

By default the deploy script will create a cluster with 3 shards and 3
replicas. However you can pass it flags to change this behavior:

```shell
$ make install && ${GOPATH}/bin/deploy -h
Usage of deploy:
  -disk="/var/lib/pfs/data.img": The disk to use for pfs' storage.
  -replicas=3: The number of replicas of each shard.
  -shards=3: The number of shards in the deploy.
```

#### Integrating with s3

If you'd like to populate your Pachyderm cluster with your own data, [jump ahead](https://github.com/pachyderm/pachyderm#using-pfs) to learn how. If not, we've created a public s3 bucket with chess data for you and we can run the chess pipeline in the full cluster.

As of v0.4 pfs can leverage s3 as a source of data for pipelines. Pfs also
uses s3 as the backend for its local Docker registry. To get s3 working you'll
need to provide pfs with credentials by setting them in etcd like so:

```
etcdctl set /pfs/creds/AWS_ACCESS_KEY_ID <AWS_ACCESS_KEY_ID>
etcdctl set /pfs/creds/AWS_SECRET_ACCESS_KEY <AWS_SECRET_ACCESS_KEY>
etcdctl set /pfs/creds/IMAGE_BUCKET <IMAGE_BUCKET>
```

#### Checking the status of your deploy

The easiest way to see what's going on in your cluster is to use `list-units`,
this is what a healthy 3 Node cluster looks like.
```
UNIT                            MACHINE                         ACTIVE          SUB
router.service      8ce43ef5.../10.240.63.167   active  running
router.service      c1ecdd2f.../10.240.66.254   active  running
router.service      e0874908.../10.240.235.196  active  running
shard-0-3:0.service e0874908.../10.240.235.196  active  running
shard-0-3:1.service 8ce43ef5.../10.240.63.167   active  running
shard-0-3:2.service c1ecdd2f.../10.240.66.254   active  running
shard-1-3:0.service c1ecdd2f.../10.240.66.254   active  running
shard-1-3:1.service 8ce43ef5.../10.240.63.167   active  running
shard-1-3:2.service e0874908.../10.240.235.196  active  running
shard-2-3:0.service c1ecdd2f.../10.240.66.254   active  running
shard-2-3:1.service 8ce43ef5.../10.240.63.167   active  running
shard-2-3:2.service e0874908.../10.240.235.196  active  running
storage.service     8ce43ef5.../10.240.63.167   active  exited
storage.service     c1ecdd2f.../10.240.66.254   active  exited
storage.service     e0874908.../10.240.235.196  active  exited
```

### The Pachyderm HTTP API

Pfs exposes a "git-like" interface to the file system -- you can add files and then create commits, branches, etc.

#### Creating files

```shell
# Write <file> to <branch>. Branch defaults to "master".
$ curl -XPOST <hostname>/file/<file>?branch=<branch> -T local_file
```

#### Reading files

```shell
# Read <file> from <master>.
$ curl <hostname>/file/<file>

# Read all files in a <directory>.
$ curl <hostname>/file/<directory>/*

# Read <file> from <commit>.
$ curl <hostname>/file/<file>?commit=<commit>
```

#### Deleting files

```shell
# Delete <file> from <branch>. Branch defaults to "master".
$ curl -XDELETE <hostname>/file/<file>?branch=<branch>
```

#### Committing changes

```shell
# Commit dirty changes to <branch>. Defaults to "master".
$ curl -XPOST <hostname>/commit?branch=<branch>

# Getting all commits.
$ curl -XGET <hostname>/commit
```

#### Branching

```shell
# Create <branch> from <commit>.
$ curl -XPOST <hostname>/branch?commit=<commit>&branch=<branch>

# Commit to <branch>
$ curl -XPOST <hostname>/commit?branch=<branch>

# Getting all branches.
$ curl -XGET <hostname>/branch
```

### Containerized Analytics

#### Creating a new pipeline with a Pachfile

Pipelines are described as Pachfiles. The Pachfile specifies a Docker image, input data, and then analysis logic (run, shuffle, etc). Pachfiles are somewhat analogous to how Docker files specify how to build a Docker image. 

```shell
{
  # Specify the Docker image you want to run your analsis in. You can pull from any registry you want. 
  image <image_name> 
  # Example: image ubuntu
  
  # Specify the input data for your analysis.  
  input <data directory>
  # Example: input my_data/users
  
  # Specify Your analysis logic and the output directory for the results. You can use they keywords `run`, `shuffle` 
  # or any shell commands you want.
  run <output directory>
  run <analysis logic>
  # Example: see the wordcount demo:                   https://github.com/pachyderm/pachyderm/examples/WordCount.md#step-3-create-the-wordcount-pipeline
}
```

#### POSTing a Pachfile to pfs

POST a text-based Pachfile with the above format to pfs:

```sh
$ curl -XPOST <hostname>/pipeline/<pipeline_name> -T <name>.Pachfile
```

**NOTE**: POSTing a Pachfile doesn't run the pipeline. It just records the specification of the pipeline in pfs. The pipeline will get run when a commit is made.

#### Running a pipeline

Pipelines are only run on a commit. That way you always know exactly the state of
the data that is used in the computation. To run all pipelines, use
the `commit` keyword.

```sh
$ curl -XPOST <hostname>/commit
```

Think of adding pipelines as constructing a
[DAG](http://en.wikipedia.org/wiki/Directed_acyclic_graph) of computations that
you want performed. When you call `/commit`, Pachyderm automatically
schedules the pipelines such that a pipeline isn't run until the pipelines it depends on have
completed.

#### Getting the output of a pipelines

Each pipeline records its output in its own read-only file system. You can read the output of the pipeline with:

```sh
$ curl -XGET <hostname>/pipeline/<piplinename>/file/*?commit=<commit>
```
or get just a specific file with:
```sh
$ curl -XGET <hostname>/pipeline/<piplinename>/file/<filename>?commit=<commit>
```

**NOTE**: You don't  need to  specify the commit you want to read from. If you use `$ curl -XGET <hostname>/pipeline/<piplinename>/file/<filename>` Pachyderm will return the most recently completed output of that pipeline. If the current pipeline is still in progress, the command will wait for it to complete before returning. We plan to update this API soon to handle these situations better. 

#### Deleting pipelines

```shell
# Delete <pipelinename>
$ curl -XDELETE <hostname>/pipeline/<pipelinename>
```

#### Getting the Pachfile

```shell
# Get the Pachfile for <pipelinename>
$ curl -XGET <hostname>/pipeline/<pipelinename>
```

## Development

We're hiring! If you like ambitious distributed systems problems and think there should be a better alternative to Hadoop, please reach out. Email jobs@pachyderm.io.

### Running

Want to hack on pachyderm for fun? You can run pachyderm locally using:

```shell
make launch-shard
```

This will build a docker image from the working directory, tag it as `pachyderm` and launch it locally.

**Note that all development must be done on linux due to the dependency on btrfs.**

And more specifically, btrfs version >= 3.14. We recommend Ubuntu 15.04 for this. See the [Environment Setup](#environment-setup)
section for a Vagrant setup that works.

Other useful development commands can be seen in the [Makefile](Makefile) and the
[bin](bin) directory. Key commands:

```
make test-deps # download all golang dependencies
make test # run all the tests
make clean # clean up all pachyderm state
make shell # go into a shell inside a running pachyderm container
./bin/run ARGS... # run a command inside a fresh pachyderm container
./bin/wrap # if run inside a ./bin/run or make shell, this will change PFS_ environment variables for a separate state from global
./bin/test ./src/PACKAGE # run tests for a specific package
./bin/test -run REGEX ./... # run all tests that match the regex
./bin/wrap ./bin/test -test.short ./... # will allow tests to be run repeatedly inside make shell
./bin/wrap go test -test.short ./... # the default settings in ./bin/test  do not have to be used
make launch-shard # launch pachyderm, as outlined above
make launch-pfsd # launch the new pfsd daemon
make install # install all binaries locally
pfs # if ${GOPATH}/bin is on your path, this will run the new pfs cli, this is very experimental and does not check for common errors
```

### Environment Setup

With golang, it's generally easiest to have your fork match the import paths in the code. We recommend you do it like this:

```
# assuming your github username is alice
rm -rf ${GOPATH}/src/github.com/pachyderm/pachyderm
mkdir -p ${GOPATH}/src/github.com/pachyderm
cd ${GOPATH}/src/github.com/pachyderm
git clone https://github.com/alice/pachyderm.git
git remote add upstream https://github.com/pachyderm/pachyderm.git # so you can run 'git fetch upstream' to get upstream changes
```

The [Vagrantfile](Vagrantfile) in this repository will set up a development environment for Pachyderm
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
curl https://raw.githubusercontent.com/pachyderm/pachyderm/master/Vagrantfile > Vagrantfile
vagrant up # starts the vagrant box
vagrant ssh # ssh into the vagrant box
```

Once in the vagrant box, set everything up and verify that it works:

```
go get github.com/pachyderm/pachyderm
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
- Check to see if the container exited with `docker ps -a | grep pfs`.
- Check the container logs with `docker logs`.

*Problem*: Docker commands are failing with permission denied

The bin scripts assume you have your user in the docker group as explained in the [Docker Ubuntu installation docs](https://docs.docker.com/installation/ubuntulinux/#create-a-docker-group).
If this is set up properly, you do not need to use `sudo` to run `docker`. If you do not want this, and want to have to use `sudo` for docker development, wrap all commands like so:

```
sudo -E bash -c 'bin/run bin/test ./...' # original command would have been `./bin/run bin/test ./...`
```

## Contributing

To get started, sign the [Contributor License Agreement](https://pachyderm.wufoo.com/forms/pachyderm-contributor-license-agreement).

Send us PRs, we would love to see what you do!
