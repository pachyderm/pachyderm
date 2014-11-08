# Pachyderm File System

## What is pfs?
Pfs is a distributed file system alternative built specifically for the Docker
ecosystem. You [https://registry.hub.docker.com/u/pachyderm/pfs/](deploy it
with Docker), just like other applications in your stack. Furthermore,
MapReduce jobs are specified as Docker containers, rather than .jars,
letting you perform distributed computation using any tools you want.

## Key Features
*TODO* make these clickable to lower sections
- Fault-tolerant architecture built on [https://coreos.com](CoreOS) (implemented)
- Git-like distributed file system (implemented)
- Dockerized MapReduce (not implemented)

## Is pfs production ready
No, pfs is at Alpha status.

## Where is this project going?
Long term Pachyderm's goal is to be a viable alternative to Hadoop built on top
of a modern distributed computing toolchain. Hadoop is a mature product so
there's a great deal of work to be done to match its feature set. However we're
finding that thanks to innovative tools like btrfs, Docker and CoreOS we can
build more functionality with less code than was possible 10 years ago.

## What is a git-like file system?
Pfs is implemented as a distributed layer on top of btrfs, the same
copy-on-write(CoW) filesystem that powers Docker. A distributed layer that
horizontally scales btrfs' existing
[http://zef.me/6023/who-needs-git-when-you-got-zfs/](git-like) semantics to
datacenter scale datasets. Pfs brings features like commit based history and
branching, standard primitives for collaborating on code, data engineering.

## What is "dockerized MapReduce?"
The basic interface for MapReduce is a `map` function and a `reduce` function.
In Hadoop this is exposed as a Java interface. In Pachyderm MapReduce jobs are
user submitted Docker containers with http servers inside them. Rather than
calling a `map` method on a class Pachyderm POSTs files to the `/map` route on
a webserver. This completely democratizes MapReduce by decoupling it from a
single platform such as the JVM.

## Quickstart Guide

### Creating a CoreOS cluster
Pfs is designed to run on CoreOS. To start you'll need a working CoreOS
cluster. Currently global containers, which are required by pfs, are only
available in the beta channel (CoreOS 444.5.0.)

- [https://coreos.com/docs/running-coreos/platforms/vagrant/](Vagrant) (reccommended)
- [https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/](Google Compute Engine)
- [https://coreos.com/docs/running-coreos/cloud-providers/ec2/](Amazon EC2)

### Deploy pfs
SSH in to one of your new machines CoreOS machines.

`$ wget https://github.com/pachyderm-io/pfs/raw/master/deploy/static/1Node.tar.gz`

`$ tar -xvf 1Node.tar.gz`

`$ fleetctl start 1Node/*`

The startup process takes a little while the first time your run it because
each node has to pull a Docker image.

### Checking the status of your deploy
The easiest way to see what's going on in your cluster is to use `list-units`
```shell
$ fleetctl list-units
```

If things are working correctly you should see something like:

```
UNIT                            MACHINE                         ACTIVE  SUB
announce-master-0-1.service     3817102d.../10.240.199.203      active  running
announce-replica-0-1.service    3817102d.../10.240.199.203      active  running
master-0-1.service              3817102d.../10.240.199.203      active  running
replica-0-1.service             3817102d.../10.240.199.203      active  running
router.service                  3817102d.../10.240.199.203      active  running
```

### Using pfs
Pfs exposes a git like interface to the file system:

#### Creating a file
```shell
$ curl -XPOST localhost/pfs/file_name -d @local_file
```

#### Read a file
```shell
$ curl localhost/pfs/file_name
```

#### Creating/modifying a file
```shell
$ curl -XPUT localhost/pfs/file_name -d @local_file
```

#### Deleting a file
```shell
$ curl -XDELETE localhost/pfs/file_name
```

#### Committing changes
```shell
$ curl localhost/commit
```

Committing in pfs creates a lightweight snapshot of the file system state and
pushes it to replicas. Where it remains accessible by commit id.

### Accessing previous commits
```shell
$ curl localhost/pfs/file_name?commit=n
```

## Who's building this?
2 guys who love data and communities. Both of whom are named Joe. We'd love
to chat: joey.zwicker@gmail.com jdoliner@gmail.com.

## How do I hack on pfs?
Pfs' only dependency is Docker. You can build it like so:
```shell
pfs$ docker build -t username/pfs .
```
Deploying what you build requires pushing the built container to the central
Docker registry and changing the container name in the .service files from
`pachyderm/pfs` to `username/pfs`.
