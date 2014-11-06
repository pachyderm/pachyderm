TODO:
Vision
Contact link
How to build
What does a git like filesystem look like.
What is dockerized mapreduce?

# Pachyderm File System

## What is pfs?
Pfs is an HDFS alternative built specifically for the Docker ecosystem.
You deploy it with Docker, just like everything else in your stack.
Furthermore, MapReduce jobs are specified as Docker containers, rather than
.jars, letting you perform distributed computation using any tools you want.

## Key Features
- Fault tolerant architecture built of CoreOS primitives
- Git like distributed filesystem
- Dockerized Map Reduce (not implemented)

## Is pfs production ready
No, pfs is at Alpha status.

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

`$ fleetctl list-units`

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
`$ curl -XPOST localhost/pfs/file_name -d @local_file`

#### Read a file
`$ curl localhost/pfs/file_name`

#### Creating/modifying a file
`$ curl -XPUT localhost/pfs/file_name -d @local_file`

#### Deleting a file
`$ curl -XDELETE localhost/pfs/file_name`

#### Committing changes
`$ curl localhost/commit`

Committing in pfs creates a lightweight snapshot of the file system state and
pushes it to replicas. Where it remains accessible by commit id.

### Accessing previous commits
`$ curl localhost/pfs/file_name?commit=n`

### What is a git like filesystem?
Pfs is implemented as a distributed layer on top of btrfs, the same
copy-on-write(CoW) filesystem that powers Docker. A distributed layer that
horizontally scales btrfs' existing
[http://zef.me/6023/who-needs-git-when-you-got-zfs/](git-like) semantics to
datacenter scale datasets. Pfs brings features like commit based history and
branching, standard primitives for collaborating on code, data engineering.

### What is "Dockerized MapReduce?"
The basic interface for MapReduce is a `map` function and a `reduce` function.
In Hadoop this is exposed as a Java interface. In Pachyderm MapReduce jobs are
user submitted Docker containers with http servers inside them. Rather than
calling a `map` method on a class Pachyderm POSTs files to the `/map` route on
a webserver. This completely democratizes MapReduce by decoupling it from a
single platform such as the JVM.
