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
Furthermore, jobs are also specified as Docker containers. Now you can use any
tools you want for distributed computation.

## Key Features
- Fault tolerant system built around CoreOS primitives (implemented)
- Rich commit history (implemented)
- Branching (not implemented)
- Dockerized Map Reduce (not implemented)

## Is pfs production ready
Absolutely not, pfs has only recently hit MVP status.

## Quickstart Guide

### Creating a CoreOS cluster
Pfs is designed to run on CoreOS. To start you'll need a working CoreOS
cluster. Currently global containers, which are required by pfs, are only
available in the beta channel (CoreOS 444.5.0.)

- Google Compute Engine (recommended): [https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/]
- Amazon EC2: [https://coreos.com/docs/running-coreos/cloud-providers/ec2/]

### Deploy pfs
SSH in to one of your new machines CoreOS machines.

`$ wget https://github.com/pachyderm-io/pfs/raw/master/deploy/static/3Node.tar.gz`

`$ tar -xvf 3Node.tar.gz`

`$ fleetctl start 3Node/*`

The startup process takes a little while the first time your run it because
each node has to pull a Docker image.

### Checking the status of your deploy
The easiest way to see what's going on in your cluster is to use `list-units`

`$ fleetctl list-units`

If things are working correctly you should see something like:

```
UNIT                            MACHINE                         ACTIVE  SUB
announce-master-0-3.service     3817102d.../10.240.199.203      active  running
announce-master-1-3.service     06c6dba9.../10.240.177.113      active  running
announce-master-2-3.service     3817102d.../10.240.199.203      active  running
announce-replica-0-3.service    f652105a.../10.240.229.124      active  running
announce-replica-1-3.service    06c6dba9.../10.240.177.113      active  running
announce-replica-2-3.service    f652105a.../10.240.229.124      active  running
master-0-3.service              3817102d.../10.240.199.203      active  running
master-1-3.service              06c6dba9.../10.240.177.113      active  running
master-2-3.service              3817102d.../10.240.199.203      active  running
replica-0-3.service             f652105a.../10.240.229.124      active  running
replica-1-3.service             06c6dba9.../10.240.177.113      active  running
replica-2-3.service             f652105a.../10.240.229.124      active  running
router.service                  06c6dba9.../10.240.177.113      active  running
router.service                  3817102d.../10.240.199.203      active  running
router.service                  f652105a.../10.240.229.124      active  running
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
