# Backup, Restore, and Migrate

## Contents

- [Introduction](#introduction)
   - [Note about releases prior to Pachyderm 1.7](#note-about-releases-prior-to-pachyderm-1-7)
- [Backup & restore concepts](#backup-restore-concepts)
- [General backup procedure](#general-backup-procedure)
  - [1. Pause all pipeline and data loading/unloading operations](#pause-all-pipeline-and-data-loading-unloading-operations)
  - [2. Extract a pachyderm backup](#extract-a-pachyderm-backup)
  - [3. Restart all pipeline and data loading operations](#restart-all-pipeline-and-data-loading-operations)
- [General restore procedure](#general-restore-procedure)
  - [ Restore your backup to a pachyderm cluster, same version](#restore-your-backup-to-a-pachyderm-cluster-same-version)
- [General migration procedure](#general-migration-procedure)
  - [Types of migrations](#types-of-migrations)
  - [Migration steps](#migration-steps)
    - [1. Pause all pipeline and data loading operations](#pause-all-pipeline-and-data-loading-operations)
    - [2. Extract a pachyderm backup with the --no-objects flag](#extract-a-pachyderm-backup-with-the-no-objects-flag)
    - [3. Clone your object store bucket](#clone-your-object-store-bucket)
    - [4. Restart all pipeline and data loading ops](#restart-all-pipeline-and-data-loading-ops)
    - [5. Deploy a 1.X Pachyderm cluster with cloned bucket](#deploy-a-1-x-pachyderm-cluster-with-cloned-bucket)
    - [6. Restore the new 1.X Pachyderm cluster from your backup](#restore-the-new-1-x-pachyderm-cluster-from-your-backup)
    - [7. Load transactional data from checkpoint into new cluster](#load-transactional-data-from-checkpoint-into-new-cluster)
    - [8. Disable the old cluster](#disable-the-old-cluster)
- [Pipeline design & operations considerations](#pipeline-design-operations-considerations)
  - [Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm)

## Introduction

Since release 1.7, Pachyderm provides the commands `pachctl extract` and `pachctl restore` to backup and restore the state of a Pachyderm cluster. (Please see [the note below about releases prior to Pachyderm 1.7](#note-about-releases-prior-to-pachyderm-1-7).)

The `pachctl extract` command requires that all pipeline and data loading activity into Pachyderm stop before the extract occurs.  This enables Pachyderm to create a consistent, point-in-time backup.  In this document, we'll talk about how to create such a backup and restore it to another Pachyderm instance.

Extract and restore commands are currently used to migrate between minor and major releases of Pachyderm, so it's important to understand how to perform them properly.   In addition, there are a few design points and operational techniques that data engineers should take into consideration when creating complex pachyderm deployments to minimize disruptions to production pipelines.

In this document, we'll take you through the steps to backup and restore a cluster, migrate an existing cluster to a newer minor or major release, and elaborate on some of those design and operations considerations.

### Note about releases prior to Pachyderm 1.7

Pachyderm 1.7 is the first version to support `extract` and `restore`.  To bridge the gap to previous Pachyderm versions,
we've made a final 1.6 release, 1.6.10, which backports the `extract` and
`restore` functionality to the 1.6 series of releases. 

Pachyderm 1.6.10 requires no
migration from other 1.6.x versions. You can simply `pachctl undeploy` and then `pachctl
deploy` after upgrading `pachctl` to version 1.6.10. After 1.6.10 is deployed you
should make a backup using `pachctl extract` and then upgrade `pachctl` again,
to 1.7.0. Finally you can `pachctl deploy ... ` with `pachctl` 1.7.0 to trigger the migration.

## Backup & restore concepts
Backing up Pachyderm involves the
persistent volume (PV) that `etcd` uses and the object store bucket that holds
Pachyderm's actual data. Restoring involves populating that PV and object store with data to recreate a Pachyderm cluster.

## General backup procedure

### 1. Pause all pipeline and data loading/unloading operations

#### Pausing pipelines
From the directed acyclic graphs (DAG) that define your pachyderm cluster, stop each pipeline.  You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the `pachctl stop-pipeline` command.

`pachctl stop-pipeline <pipeline name>`

You can confirm each pipeline is paused using the `pachctl list-pipeline` command

`pachctl list-pipeline`

Alternatively, a useful shell script for running stop-pipeline on all pipelines is included below.   It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```
pachctl list-pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl stop-pipeline {}
```

It's also a useful practice, for simple to moderately complex deployments, to keep a terminal window up showing the state of all running kubernetes pods.

`watch -n 5 kubectl get pods`

You may need to install the `watch` and `kubectl` commands on your system, and configure `kubectl` to point at the cluster that Pachyderm is running in.

#### Pausing data loading operations
**Input repositories** or **input repos** in pachyderm are repositories created with the `pachctl create repo` command.  They're designed to be the repos at the top of a directed acyclic graph of pipelines. Pipelines have their own output repos associated with them, and are not considered input repos. If there are any processes external to pachyderm that put data into input repos using any method (the Pachyderm APIs, `pachctl put-file`, etc.), they need to be paused.  See [Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm) below for design considerations for those processes that will minimize downtime during a restore or migration.

Alternatively, you can use the following commands to stop all data loading into Pachyderm from outside processes.

```
# Once you have stopped all running pachyderm pipelines, such as with this command,
# $ pachctl list-pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl stop-pipeline {}

# all pipelines in your cluster should be suspended. To stop all
# data loading processes, we're going to modify the pachd Kubernetes service so that
# it only accepts traffic on port 30649 (instead of the usual 30650). This way,
# any background users and services that send requests to your Pachyderm cluster
# while 'extract' is running will not interfere with the process
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pach -o json >pach_service_backup_30650.json

# Modify the service to accept traffic on port 30649
# Note that you'll likely also need to modify your cloud provider's firewall
# rules to allow traffic on this port
$ kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on this new port
$ export PACHD_ADDRESS="${PACHD_ADDRESS/30650/30649}"

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             1.7.11
pachd               1.7.11
```

### 2. Extract a pachyderm backup

#### Using extract

Using the `pachctl extract` command, create the backup you need.

`pachctl extract > path/to/your/backup/file`

You can also use the `-u` or `--url` flag to put the backup directly into an object store.

`pachctl extract --url s3://...`

If you are planning on backing up the object store using its own built-in clone operation, be sure to add the `--no-objects` flag to the `pachctl extract` command.

#### Using your cloud provider's clone and snapshot services
You should follow your cloud provider's recommendation
for backing up these resources. Here are some pointers to the relevant documentation.

##### Snapshotting persistent volumes
For example, here are official guides on creating snapshots of persistent volumes on Google Cloud Platform, Amazon Web Services (AWS) and Microsoft Azure, respectively:

- [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Creating snapshots of Elastic Block Store (EBS) volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)
- [Creating snapshots of Azure Virtual Hard Disk volumes](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/snapshot-copy-managed-disk)


##### Cloning object stores
Below, we give an example using the Amazon Web Services CLI to clone one bucket to another, [taken from the documentation for that command](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).  Similar commands are available for [Google Cloud](https://cloud.google.com/storage/docs/gsutil/commands/cp) and [Azure blob storage](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).

`aws s3 sync s3://mybucket s3://mybucket2`

#### Combining cloning, snapshots and extract/restore

You can use `pachctl extract` command with the `--no-objects` flag to exclude the object store, and use an object store snapshot or clone command to back up the object store. You can run the two commands at the same time.  For example, on Amazon Web Services, the following commands can be run simultaneously.

`aws s3 sync s3://mybucket s3://mybucket2`

`pachctl extract --no-objects --url s3://anotherbucket`

#### Use case: minimizing downtime during a migration
The above cloning/snapshotting technique is recommended when doing a migration where minimizing downtime is desirable, as it allows the duplicated object store to be the basis of the upgraded, new cluster instead of requiring Pachyderm to extract the data from object store.

### 3. Restart all pipeline and data loading operations

Once the backup is complete, restart all paused pipelines and data loading operations.

From the directed acyclic graphs (DAG) that define your pachyderm cluster, start each pipeline.    You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the `pachctl start-pipeline` command.

`pachctl start-pipeline pipeline-name`

You can confirm each pipeline is started using the pachctl list-pipeline command

`pachctl list-pipeline`

A useful shell script for running `start-pipeline` on all pipelines is included below.  It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```
pachctl list-pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl start-pipeline {}
```

If you used the port-changing technique, [above](#pausing-data-loading-operations), to stop all data loading into Pachyderm from outside processes, you should change the ports back.

```
# Once you have restarted all running pachyderm pipelines, such as with this command,
# $ pachctl list-pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl start-pipeline {}

# all pipelines in your cluster should be restarted. To restart all data loading 
# processes, we're going to change the pachd Kubernetes service so that
# it only accepts traffic on port 30650 again (from 30649). 
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pach -o json >pach_service_backup_30649.json

# Modify the service to accept traffic on port 30650, again
$ kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on the old port
$ export PACHD_ADDRESS="${PACHD_ADDRESS/30649/30650}"

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             1.7.11
pachd               1.7.11
```
## General restore procedure
### Restore your backup to a pachyderm cluster, same version

Spin up a Pachyderm cluster and run `pachctl restore` with the backup you created earlier.

`pachctl restore < path/to/your/backup/file`

You can also use the `-u` or `--url` flag to get the backup directly from the object store you placed it in

`pachctl restore --url s3://...`

## General migration procedure

### Types of migrations
You'll need to follow one of two procedures when upgrading a Pachyderm cluster:

1. Simple migration path (for minor versions, e.g. v1.8.2 to v1.8.3)
1. Full migration (for major versions, e.g. v1.7.11 to v1.8.0)

The simple migration path is shorter and faster than the full migration path
because it can reuse Pachyderm's existing internal state.

#### Difference in detail
In general, Pachyderm stores all of its state in two places: `etcd` (which in
turn stores its state in one or more persistent volumes, which were created when
the Pachyderm cluster was deployed) and an object store bucket (in e.g. S3).

In the simple migration path, you only need to deploy new Pachyderm pods that
reference the existing data, and the existing data structures residing there
will be re-used. 

In the full migration path, however, those data structures need
to be read, transformed, and rewritten, so the process involves:

1. bringing up a new Pachyderm cluster adjacent to the old pachyderm cluster
1. exporting the old Pachdyerm cluster's repos, pipelines, and input commits
1. importing the old cluster's repos, commits, and pipelines into the new
   cluster, and then re-running all of the old cluster's pipelines from scratch
   once.
   
#### Note about 1.7 to 1.8 migrations

Most Pachyderm 1.8 data structures are substantively different from their 1.7
counterparts, and in particular, Pachyderm v1.8+ stores more information in its
output commits than its predecessors, which enables useful performance
improvements. Old (v1.7) Pachyderm output commits don't have this extra
information, so Pachyderm 1.8 must re-create its output commits.

#### Before you start: backups
We recommend, before starting any migration effort, that you back up Pachyderm's
pre-migration state. Specifically, you must back up the
persistent volume that etcd uses and the object store bucket that holds
pachyderm's actual data, as shown above in [ General backup & restore procedure](#general-backup-procedure). 

### Migration steps
#### 1. Pause all pipeline and data loading operations

From the directed acyclic graphs (DAG) that define your pachyderm cluster, stop each pipeline step.  You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the 'pachctl stop-pipeline' command.

`pachctl stop-pipeline pipeline_name`

You can confirm each pipeline is paused using the `pachctl list-pipeline` command

`pachctl list-pipeline`

Alternatively, a useful shell script for running stop-pipeline on all pipelines is included below.  It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```
pachctl list-pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl start-pipeline {}
```

It's also a useful practice, for simple to moderately complex deployments, to keep a terminal window up showing the state of all running kubernetes pods.

`watch -n 5 kubectl get pods`

You may need to install the `watch` and `kubectl` commands on your system, and configure `kubectl` to point at the cluster that Pachyderm is running in.

#### Pausing data loading operations
**Input repositories** or **input repos** in Pachyderm are repositories created with the `pachctl create repo` command.  They're designed to be the repos at the top of a directed acyclic graph of pipelines. Pipelines have their own output repos associated with them, and are not considered input repos. If there are any processes external to pachyderm that put data into input repos using any method (the Pachyderm APIs, `pachctl put-file`, etc.), they need to be paused.  See [Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm) below for design considerations for those processes that will minimize downtime during a restore or migration.

Alternatively, you can use the following commands to stop all data loading into Pachyderm from outside processes.

```
# Once you have stopped all running pachyderm pipelines, such as with this command,
# $ pachctl list-pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl stop-pipeline {}

# all pipelines in your cluster should be suspended. To stop all
# data loading processes, we're going to modify the pachd Kubernetes service so that
# it only accepts traffic on port 30649 (instead of the usual 30650). This way,
# any background users and services that send requests to your Pachyderm cluster
# while 'extract' is running will not interfere with the process
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pach -o json >pach_service_backup_30650.json

# Modify the service to accept traffic on port 30649
# Note that you'll likely also need to modify your cloud provider's firewall
# rules to allow traffic on this port
$ kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on this new port
$ export PACHD_ADDRESS="${PACHD_ADDRESS/30650/30649}"

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             1.7.11
pachd               1.7.11
```

### 2. Extract a pachyderm backup with the --no-objects flag

This step and the following step, [3. Clone your object store bucket](#clone-your-object-store-bucket), can be run simultaneously.

Using the `pachctl extract` command, create the backup you need.

`pachctl extract --no-objects > path/to/your/backup/file`

You can also use the `-u` or `--url` flag to put the backup directly into an object store.

`pachctl extract --no-objects --url s3://...`

Note that this s3 bucket is different than the s3 bucket will create to clone your object store.  This is merely a bucket you allocated to hold the pachyderm backup without objects.

### 3. Clone your object store bucket

This step and the prior step, [2. Extract a pachyderm backup with the --no-objects flag](#extract-a-pachyderm-backup-with-the-no-objects-flag), can be run simultaneously.  Run the command that will clone a bucket in your object store.

Below, we give an example using the Amazon Web Services CLI to clone one bucket to another, [taken from the documentation for that command](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).  Similar commands are available for [Google Cloud](https://cloud.google.com/storage/docs/gsutil/commands/cp) and [Azure blob storage](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).

`aws s3 sync s3://mybucket s3://mybucket2`

### 4. Restart all pipeline and data loading ops

Once the backup and clone operations are complete, restart all paused pipelines and data loading operations, setting a checkpoint for the started operations that you can use in step [7. Load transactional data from checkpoint into new cluster](#load-transactional-data-from-checkpoint-into-new-cluster), below.  See [Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm) below  to understand why designing this checkpoint into your data loading systems is important.

From the directed acyclic graphs (DAG) that define your pachyderm cluster, start each pipeline.    You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the 'pachctl start-pipeline' command.

`pachctl start-pipeline pipeline_name`

You can confirm each pipeline is started using the `pachctl list-pipeline` command

`pachctl list-pipeline`

A useful shell script for running `start-pipeline` on all pipelines is included below.  It may be necessary to install several of the utlilies used in the script, like jq, on your system.

```
pachctl list-pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl start-pipeline {}
```

If you used the port-changing technique, [above](#pause-all-pipeline-and-data-loading-operations), to stop all data loading into Pachyderm from outside processes, you should change the ports back.

```
# Once you have restarted all running pachyderm pipelines, such as with this command,
# $ pachctl list-pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl start-pipeline {}

# all pipelines in your cluster should be restarted. To restart all data loading 
# processes, we're going to change the pachd Kubernetes service so that
# it only accepts traffic on port 30650 again (from 30649). 
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pach -o json >pach_service_backup_30649.json

# Modify the service to accept traffic on port 30650, again
$ kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on the old port
$ export PACHD_ADDRESS="${PACHD_ADDRESS/30649/30650}"

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             1.7.11
pachd               1.7.11
```

Your old pachyderm cluster can operate while you're creating a migrated one.  It's important that your data loading operations are designed to use the "[Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm)" design criteria below for this to work.

### 5. Deploy a 1.X Pachyderm cluster with cloned bucket

Create a pachyderm cluster using the bucket you cloned in [3. Clone your object store bucket](#clone-your-object-store-bucket). 

You'll want to bring up this new pachyderm cluster in a different
namespace.  You'll check at the steps below to see if there was some
kind of problem with the extracted data and steps
[2](#extract-a-pachyderm-backup-with-the-no-objects-flag) and
[3](#clone-your-object-store-bucket) need to be run again. Once your
new cluster is up and you're connected to it, go on to the next step.

_Important: Use the_ `kubectl config current-config` _command to confirm you're talking to the correct kubernetes cluster configuration for the new cluster._

### 6. Restore the new 1.X Pachyderm cluster from your backup

Using the Pachyderm cluster you deployed in the previous step, [5. Deploy a 1.X pachyderm cluster with cloned bucket](#deploy-a-1.X-pachyderm-cluster-with-cloned-bucket), run `pachctl restore` with the backup you created in [2. Extract a pachyderm backup with the --no-objects flag](#extract-a-pachyderm-backup-with-the-no-objects-flag).

_Important: Use the_ `kubectl config current-config` _command to confirm you're talking to the correct kubernetes cluster configuration_

`pachctl restore < path/to/your/backup/file`

You can also use the `-u` or `--url` flag to get the backup directly from the object store you placed it in

`pachctl restore --url s3://...`

Note that this s3 bucket is different than the s3 bucket you cloned, above.  This is merely a bucket you allocated to hold the Pachyderm backup without objects.

### 7. Load transactional data from checkpoint into new cluster

Configure an instance of your data loading systems to point at the new, upgraded pachyderm cluster and play back transactions from the checkpoint you established in [4. Restart all pipeline and data loading operations](#restart-all-pipeline-and-data-loading-operations). Confirm that the data output is as expected and the new cluster is operating as expected.

You may also need to reconfigure data loading operations from Pachyderm to processes outside of it to work as expected.

### 8. Disable the old cluster

Once you've confirmed that the new cluster is operating, you can disable the old cluster.

## Pipeline design & operations considerations

### Loading data from other sources into pachyderm

When writing systems that place data into Pachyderm input repos (see [above](#pausing-data-loading-operations) for a definition of 'input repo'), it is important to provide ways of 'pausing' output while queueing any data output requests to be output when the systems are 'resumed'.  This allows all pachyderm processing to be stopped while the extract takes place.

In addition, it is desirable for systems that load data into Pachyderm have a mechanism for replaying a queue from any checkpoint in time.  This is useful when doing migrations from one release to another, where you would like to minimize downtime of a production Pachyderm system.  After an extract, the old system is kept running with the checkpoint established while a duplicate, upgraded pachyderm cluster is migrated with duplicated data. Transactions that occur while the migrated, upgraded cluster is being brought up are not lost, and can be replayed into this new cluster to reestablish state and minimize downtime.
