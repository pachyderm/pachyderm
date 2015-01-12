# Pachyderm File System

## What is pfs?
Pfs is a distributed file system built specifically for the Docker
ecosystem. You [deploy it with Docker](https://registry.hub.docker.com/u/pachyderm/pfs/),
just like other applications in your stack. Furthermore,
MapReduce jobs are specified as Docker containers, rather than .jars,
letting you perform distributed computation using any tools you want.

## Key Features
- Fault-tolerant architecture built on [CoreOS](https://coreos.com) (implemented)
- [Git-like distributed file system](#what-is-a-git-like-file-system) (implemented)
- [Dockerized MapReduce](#what-is-dockerized-mapreduce) (not implemented)

## Is pfs production ready
No, pfs is at Alpha status. [We'd love your help. :)](#how-do-i-hack-on-pfs)

## Where is this project going?
Pachyderm will eventually be a complete replacement for Hadoop, built on top of
a modern toolchain instead of the JVM. Hadoop is a mature ecosystem, so there's
a long way to go before pfs will fully match its feature set. However, thanks to innovative tools like btrfs, Docker, and CoreOS, we can build an order of magnitude more functionality with much less code.

## What is a "git-like file system"?
Pfs is implemented as a distributed layer on top of btrfs, the same
copy-on-write file system that powers Docker. Btrfs already offers
[git-like semantics](http://zef.me/6023/who-needs-git-when-you-got-zfs/) on a
single machine; pfs scales these out to an entire cluster. This allows features such as:
- Commit-based history: File systems are generally single-state entities. Pfs,
on the other hand, provides a rich history of every previous state of your
cluster. You can always revert to a prior commit in the event of a
disaster.
- Branching: Thanks to btrfs's copy-on-write semantics, branching is ridiculously
cheap in pfs. Each user can experiment freely in their own branch without
impacting anyone else or the underlying data. Branches can easily be merged back in the main cluster.
- Cloning: Btrfs's send/receive functionality allows pfs to efficiently copy
an entire cluster's worth of data while still maintaining its commit history.

## What is "dockerized MapReduce?"
The basic interface for MapReduce is a `map` function and a `reduce` function.
In Hadoop this is exposed as a Java interface. In Pachyderm, MapReduce jobs are
user-submitted Docker containers with http servers inside them. Rather than
calling a `map` method on a class, Pachyderm POSTs files to the `/map` route on
a webserver. This completely democratizes MapReduce by decoupling it from a
single platform, such as the JVM.

Thanks to Docker, Pachyderm can seamlessly integrate external libraries. For example, suppose you want to perform computer
vision on a large set of images. Creating this job is as simple as
running `npm install opencv` inside a Docker container and creating a node.js server, which uses this library on its `/map` route.

## Quickstart Guide
The easiest way to try out pfs is to point curl at the live instance we have
running here: http://release.pachyderm.io. We'll try to keep it up and running throughout
the day.

### Creating a CoreOS cluster
Pfs is designed to run on CoreOS. To start, you'll need a working CoreOS
cluster. Currently global containers, which are required by pfs, are only
available in the beta channel (CoreOS 444.5.0)

- [Google Compute Engine](https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/) (recommended)
- [Amazon EC2](https://coreos.com/docs/running-coreos/cloud-providers/ec2/)
- [Vagrant](https://coreos.com/docs/running-coreos/platforms/vagrant/) (requires setting up DNS)

### Deploy pfs
SSH in to one of your new machines CoreOS machines.

```shell
$ wget https://github.com/pachyderm-io/pfs/raw/master/deploy/static/1Node.tar.gz
$ tar -xvf 1Node.tar.gz
$ fleetctl start 1Node/*
```

The startup process takes a little while the first time you run it because
each node has to pull a Docker image.

### Checking the status of your deploy
The easiest way to see what's going on in your cluster is to use `list-units`
```shell
$ fleetctl list-units
```

If things are working correctly, you should see something like:

```
UNIT                            MACHINE                         ACTIVE  SUB
announce-master-0-1.service     3817102d.../10.240.199.203      active  running
master-0-1.service              3817102d.../10.240.199.203      active  running
router.service                  3817102d.../10.240.199.203      active  running
```

### Using pfs
Pfs exposes a git-like interface to the file system:

#### Creating files
```shell
# Write <file> to <branch>. Branch defaults to "master".
$ curl -XPOST pfs/file/<file>?branch=<branch> -d @local_file
```

#### Reading files
```shell
# Read <file> from <master>.
$ curl pfs/file/<file>
# Read all files in a <directory>.
$ curl pfs/file/<directory>/*
# Read <file> from <commit>.
$ curl pfs/file/<file>?commit=<commit>
```

#### Deleting files
```shell
# Delete <file> from <branch>. Branch defaults to "master".
$ curl -XDELETE pfs/file/<file>?branch=<branch>
```

#### Committing changes
```shell
# Commit dirty changes to <branch>. Defaults to "master".
$ curl -XPOST pfs/commit?branch=<branch>
# Getting all commits.
$ curl -XGET pfs/commit
```

#### Branching
```shell
# Create <branch> from <commit>.
$ curl -XPOST pfs/branch?commit=<commit>&branch=<branch>
# Commit to <branch>
$ curl -XPOST pfs/commit?branch=<branch>
# Getting all branches.
$ curl -XGET pfs/branch
```
###MapReduce

####Creating a job:

```
# Job format:
{
    "type"  : either "map" or "reduce"
    "input" : a file in pfs or the output from another job
    "image" : the Docker image
    "command" : the command to start your web server
}
```

```shell
# Write a job to pfs:
$ curl -XPOST <host>/job/<jobname> -T <localfile>.json
```
#### Deleting jobs

```shell
# Delete <job>
$ curl -XDELETE <host>/job/<job>
```

#### Getting jobs

```shell
# Read <job>
$ curl -XGET <host>/job/<job>
```

#### Running jobs

```shell
# Commit and run all jobs:
$ curl -XPOST <host>/commit?run
```

#### Getting output from jobs
```shell
# Read <file> from the output of <job> at <commit>:
$ curl -XGET <host>/job/<job>/file/<file>?commit=<commit>

# Read the output of <job> at <commit>:
$ curl -XGET <host>/job/<job>/file/*?commit=<commit>
```
```shell
# Get the output of jobs:
$ curl -XGET <host>/job/<jobname>/file/<filename>
```



## Who's building this?
Two guys who love data and communities and both happen to be named Joe. We'd love
to chat: joey@pachyderm.io jdoliner@pachyderm.io.

## How do I hack on pfs?
Pfs's only dependency is Docker. You can build it with:
```shell
pfs$ docker build -t username/pfs .
```
Deploying what you build requires pushing the built container to the central
Docker registry and changing the container name in the .service files from
`pachyderm/pfs` to `username/pfs`.
