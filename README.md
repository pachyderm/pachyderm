## News
Pachyderm v0.7 is out. v0.7 includes replication, automatic failover, and a rigorous testing suite. 

We're hiring! Pachyderm is looking for our first hire. If you'd like to get involved, email us at jobs@pachyderm.io

## What is pfs?
Pachyderm is a distributed file system and analytics engine built specifically for containerized infrastructures. 
You [deploy it with Docker](https://registry.hub.docker.com/u/pachyderm/pfs/),
just like the other applications in your stack. Furthermore,
analysis jobs are specified as containers, rather than .jars,
letting you perform distributed computation using any tools you want.

## Key Features
- Fault-tolerant architecture built on [CoreOS](https://coreos.com)
- [Git-like distributed file system](#what-is-a-git-like-file-system)
- [Containerized analytics engine](#what-is-containerized-analytics)

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

## What is "containerized analytics?"
Rather than thinking in terms of map or reduce jobs, pps thinks in terms of pipelines expressed within a container. A pipeline is a generic way expressing computation over large datasets and it’s containerized to make it easily portable, isolated, and easy to monitor. In Pachyderm, all analysis runs in containers. You can write them in any language you want and include any libraries. 
For example, suppose you want to perform computer vision on a large set of images. Creating this job is as simple as
running `npm install opencv` inside a Docker container.

## Quickstart Guide
### Tutorial -- Analyzing chess games
#### Run Pachyderm locally on a small sample dataset
```shell
# launch a local pfs shard
$ curl www.pachyderm.io/launch | sh

# clone the chess pipeline
$ git clone https://github.com/pachyderm/chess.git && cd chess

# install the pipeline locally and run it
$ install/pachyderm/local
```
#####Step 1: Launch a local pfs shard
Download and run the Pachyderm launch script to get a local instance running.
#####Step 2: Clone the chess pipeline
Clone the chess git repo we’ve provided. You can check out the full map code [here](https://github.com/pachyderm/chess).
#####Step 3: Install and run the pipeline locally
Run the local install script to start the pipeline. It should take around 6 minutes to complete the analysis.

#### Creating a CoreOS cluster
Pfs is designed to run on CoreOS. To start, you'll need a working CoreOS
cluster. Here's links on how to set one up:

- [Amazon EC2](https://coreos.com/docs/running-coreos/cloud-providers/ec2/) (recommended)
- [Google Compute Engine](https://coreos.com/docs/running-coreos/cloud-providers/google-compute-engine/)
- [Vagrant](https://coreos.com/docs/running-coreos/platforms/vagrant/) (requires setting up DNS)

#### Deploy pfs
SSH in to one of your new CoreOS machines.

```shell
$ curl pachyderm.io/deploy | sh
```

The startup process takes a little while the first time you run it because
each node has to pull a Docker image.

####  Settings
By default the deploy script will create a cluster with 3 shards and 3
replicas. However you can pass it flags to change this behavior:

```shell
$ ./deploy -h
Usage of /go/bin/deploy:
  -container="pachyderm/pfs": The container to use for the deploy.
  -replicas=3: The number of replicas of each shard.
  -shards=3: The number of shards in the deploy.
```

### Integrating with s3
As of v0.4 pfs can leverage s3 as a source of data for MapReduce jobs. Pfs also
uses s3 as the backend for its local Docker registry. To get s3 working you'll
need to provide pfs with credentials by setting them in etcd like so:

```
etcdctl set /pfs/creds/AWS_ACCESS_KEY_ID <AWS_ACCESS_KEY_ID>
etcdctl set /pfs/creds/AWS_SECRET_ACCESS_KEY <AWS_SECRET_ACCESS_KEY>
etcdctl set /pfs/creds/IMAGE_BUCKET <IMAGE_BUCKET>
```

### Checking the status of your deploy
The easiest way to see what's going on in your cluster is to use `list-units`,
this is what a healthy 1 Node cluster looks like.
```
UNIT                            MACHINE                         ACTIVE          SUB
announce-master-0-1.service     0b0625cf.../172.31.9.86         active          running
announce-registry.service       0e7cf611.../172.31.27.115       active          running
gitdaemon.service               0b0625cf.../172.31.9.86         active          running
gitdaemon.service               0e7cf611.../172.31.27.115       active          running
gitdaemon.service               ed618559.../172.31.9.87         active          running
master-0-1.service              0b0625cf.../172.31.9.86         active          running
registry.service                0e7cf611.../172.31.27.115       active          running
router.service                  0b0625cf.../172.31.9.86         active          running
router.service                  0e7cf611.../172.31.27.115       active          running
router.service                  ed618559.../172.31.9.87         active          running
```
If you startup a new cluster and `registry.service` fails to start it's
probably an issue with s3 credentials. See the section above.

## Using pfs
Pfs exposes a git-like interface to the file system:

#### Creating files
```shell
# Write <file> to <branch>. Branch defaults to "master".
$ curl -XPOST pfs/file/<file>?branch=<branch> -T local_file
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
##Containerized Analytics

####Creating a new pipeline descriptor

Pipelines and jobs are specified as JSON files in the following format:

```
{
    "type"  : either "map" or "reduce"
    "input" : a directory in pfs, S3 URL, or the output from another job
    "image" : the Docker image to use 
    "command" : the command to start your web server
}
```

**NOTE**: You do not need to specify the output location for a job. The output of a job, often referred to as a _materialized view_, is automatically stored in pfs `/job/<jobname>`.

####POSTing a job to pfs

Post a local JSON file with the above format to pfs:

```sh
$ curl -XPOST <host>/job/<jobname> -T <localfile>.json
```

**NOTE**: POSTing a job doesn't run the job. It just records the specification of the job in pfs. 

####Running a job
Jobs are only run on a commit. That way you always know exactly the state of
the file system that is used in a computation. To run all committed jobs, use
the `commit` keyword with the `run` parameter.

```sh
$ curl -XPOST <host>/commit?run
```

Think of adding jobs as constructing a
[DAG](http://en.wikipedia.org/wiki/Directed_acyclic_graph) of computations that
you want performed. When you call `/commit?run`, Pachyderm automatically
schedules the jobs such that a job isn't run until the jobs it depends on have
completed.

####Getting the output of a job
Each job records its output in its own read-only file system. You can read the output of the job with:

```sh
$ curl <host>/job/<jobname>/file/*?commit=<commit>
```
or get just a specific file with:
```sh
$ curl -XGET <host>/job/<job>/file/*?commit=<commit>
```

**NOTE**: You must specify the commit you want to read from and that commit
needs to have been created with the run parameter. We're planning to expand
this API to make it not have this requirement in the near future.
####Creating a job:


#### Deleting jobs

```shell
# Delete <job>
$ curl -XDELETE <host>/job/<job>
```

#### Getting the job descriptor

```shell
# Read <job>
$ curl -XGET <host>/job/<job>
```

## Who's building this?
Two guys who love data and communities and both happen to be named Joe. We'd love
to chat: joey@pachyderm.io jdoliner@pachyderm.io.

## How do I hack on pfs?
We're hiring! If you like ambitious distributed systems problems and think there should be a better alternative to Hadoop, please reach out.  Email jobs@pachyderm.io

Want to hack on pfs for fun?
You can run pfs locally using:

```shell
scripts/dev-launch
```

This will build a docker image from the working directory, tag it as `pfs` and
launch it locally using `scripts/launch`.  The only dependencies are Docker >=
1.5 and btrfs-tools >= 3.14. The script checks for this and gives you
directions on how to fix it.
