# Migration

!!! info
    If you need to upgrade Pachyderm from one minor version
    to another, such as from 1.9.4 to 1.9.5, see
    [Upgrade Pachyderm](upgrades.md).

- [Introduction](#introduction)
- [Note about 1.7 to 1.8 migrations](#note-about-1-7-to-1-8-migrations]
- [General migration procedure](#general-migration-procedure)
  - [Before you start: backups](#before-you-start-backups)
  - [Migration steps](#migration-steps)
    - [1. Pause all pipeline and data loading operations](#1-pause-all-pipeline-and-data-loading-operations)
    - [2. Extract a pachyderm backup with the --no-objects flag](#2-extract-a-pachyderm-backup-with-the-no-objects-flag)
    - [3. Clone your object store bucket](#3-clone-your-object-store-bucket)
    - [4. Restart all pipeline and data loading ops](#4-restart-all-pipeline-and-data-loading-ops)
    - [5. Deploy a 1.X Pachyderm cluster with cloned bucket](#5-deploy-a-1x-pachyderm-cluster-with-cloned-bucket)
    - [6. Restore the new 1.X Pachyderm cluster from your backup](#6-restore-the-new-1x-pachyderm-cluster-from-your-backup)
    - [7. Load transactional data from checkpoint into new cluster](#7-load-transactional-data-from-checkpoint-into-new-cluster)
    - [8. Disable the old cluster](#8-disable-the-old-cluster)
    - [9. Reconfigure new cluster as necessary](#9-reconfigure-new-cluster-as-necessary)

## Introduction

As new versions of Pachyderm are released, you may need to update your cluster to get access to bug fixes and new features. 
These updates fall into two categories, upgrades and migrations.

An upgrade is moving between point releases within the same major release, 
like 1.7.2 to 1.7.3.
Upgrades are typically a simple process that require little to no downtime.

Migrations involve moving between major releases, 
like 1.8.6 to 1.9.0.
Migration is covered in this document. 

In general, 
Pachyderm stores all of its state in two places: 
`etcd` 
(which in turn stores its state in one or more persistent volumes,
which were created when the Pachyderm cluster was deployed) 
and an object store bucket 
(something like AWS S3, MinIO, or Azure Blob Storage).

In a migration, 
the data structures stored in those locations need to be read, transformed, and rewritten, so the process involves:

1. bringing up a new Pachyderm cluster adjacent to the old pachyderm cluster
1. exporting the old Pachdyerm cluster's repos, pipelines, and input commits
1. importing the old cluster's repos, commits, and pipelines into the new
   cluster.

*You must perform a migration to move between major releases*,
such as 1.8.7 to 1.9.0.

Whether you're doing an upgrade or migration, it is recommended you [backup Pachyderm](../backup_restore/#general-backup-procedure) prior.
That will guarantee you can restore your cluster to its previous, good state.

## Note about 1.7 to 1.8 migrations

In Pachyderm 1.8,
we rearchitected core parts of the platform to [improve speed and scalability](http://www.pachyderm.io/2018/11/15/performance-improvements.html).
Migrating from 1.7.x to 1.8.x using the procedure below can a fairly lengthy process.
If your requirements fit, it may be easier to create a new 1.8 or greater cluster and reload your latest source data into your input repositories.

You may wish to keep your original 1.7 cluster around in a suspended state, reactivating it in case you need access to that provenance data.

## General migration procedure

### Before you start: backups

Please refer to [the documentation on backing up your cluster](../backup_restore/#general-backup-procedure).

### Migration steps
#### 1. Pause all pipeline and data loading operations

From the directed acyclic graphs (DAG) that define your pachyderm cluster, stop each pipeline step.  You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the `stop pipeline` command.

`pachctl stop pipeline <pipeline-name>`

You can confirm each pipeline is paused using the `list pipeline` command

`pachctl list pipeline`

Alternatively, a useful shell script for running `stop pipeline` on all pipelines is included below.  It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```shell
pachctl list pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl stop pipeline {}
```

It's also a useful practice, for simple to moderately complex deployments, to keep a terminal window up showing the state of all running kubernetes pods.

`watch -n 5 kubectl get pods`

You may need to install the `watch` and `kubectl` commands on your system, and configure `kubectl` to point at the cluster that Pachyderm is running in.

#### Pausing data loading operations
**Input repositories** or **input repos** in Pachyderm are repositories created with the `pachctl create repo` command.
They're designed to be the repos at the top of a directed acyclic graph of pipelines.
Pipelines have their own output repos associated with them, and are not considered input repos.
If there are any processes external to pachyderm that put data into input repos using any method
(the Pachyderm APIs, `pachctl put file`, etc.), 
they need to be paused.  
See [Loading data from other sources into pachyderm](../backup_restore/#loading-data-from-other-sources-into-pachyderm) below for design considerations for those processes that will minimize downtime during a restore or migration.

Alternatively, you can use the following commands to stop all data loading into Pachyderm from outside processes.

```
# Once you have stopped all running pachyderm pipelines, such as with this command,
# $ pachctl list pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl stop pipeline {}

# all pipelines in your cluster should be suspended. To stop all
# data loading processes, we're going to modify the pachd Kubernetes service so that
# it only accepts traffic on port 30649 (instead of the usual 30650). This way,
# any background users and services that send requests to your Pachyderm cluster
# while 'extract' is running will not interfere with the process
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pachd -o json >pach_service_backup_30650.json

# Modify the service to accept traffic on port 30649
# Note that you'll likely also need to modify your cloud provider's firewall
# rules to allow traffic on this port
$ kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on this new port
$ pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30649

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pachctl version
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

### 2. Extract a pachyderm backup with the --no-objects flag

This step and the following step, [3. Clone your object store bucket](#3-clone-your-object-store-bucket), can be run simultaneously.

Using the `pachctl extract` command, create the backup you need.

`pachctl extract --no-objects > path/to/your/backup/file`

You can also use the `-u` or `--url` flag to put the backup directly into an object store.

`pachctl extract --no-objects --url s3://...`

Note that this s3 bucket is different than the s3 bucket will create to clone your object store.
This is merely a bucket you allocated to hold the pachyderm backup without objects.

### 3. Clone your object store bucket

This step and the prior step,
[2. Extract a pachyderm backup with the --no-objects flag](#2-extract-a-pachyderm-backup-with-the-no-objects-flag),
can be run simultaneously.
Run the command that will clone a bucket in your object store.

Below, we give an example using the Amazon Web Services CLI to clone one bucket to another,
[taken from the documentation for that command](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).
Similar commands are available for [Google Cloud](https://cloud.google.com/storage/docs/gsutil/commands/cp)
and [Azure blob storage](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).

`aws s3 sync s3://mybucket s3://mybucket2`

### 4. Restart all pipeline and data loading ops

Once the backup and clone operations are complete,
restart all paused pipelines and data loading operations,
setting a checkpoint for the started operations that you can use in step [7. Load transactional data from checkpoint into new cluster](#7-load-transactional-data-from-checkpoint-into-new-cluster), below.
See [Loading data from other sources into pachyderm](../backup_restore/#loading-data-from-other-sources-into-pachyderm) to understand why designing this checkpoint into your data loading systems is important.

From the directed acyclic graphs (DAG) that define your pachyderm cluster,
start each pipeline.
You can either run a multiline shell command, 
shown below,
or you must,
for each pipeline,
manually run the 'start pipeline' command.

`pachctl start pipeline <pipeline-name>`

You can confirm each pipeline is started using the `list pipeline` command

`pachctl list pipeline`

A useful shell script for running `start pipeline` on all pipelines is included below.
It may be necessary to install several of the utlilies used in the script, like jq, on your system.

```shell
pachctl list pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl start pipeline {}
```

If you used the port-changing technique,
[above](#1-pause-all-pipeline-and-data-loading-operations),
to stop all data loading into Pachyderm from outside processes,
you should change the ports back.

```
# Once you have restarted all running pachyderm pipelines, such as with this command,
# $ pachctl list pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl start pipeline {}

# all pipelines in your cluster should be restarted. To restart all data loading 
# processes, we're going to change the pachd Kubernetes service so that
# it only accepts traffic on port 30650 again (from 30649). 
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pachd -o json >pachd_service_backup_30649.json

# Modify the service to accept traffic on port 30650, again
$ kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on the old port
$ pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

Your old pachyderm cluster can operate while you're creating a migrated one.
It's important that your data loading operations are designed to use the "[Loading data from other sources into pachyderm](../backup_restore/#loading-data-from-other-sources-into-pachyderm)" design criteria below for this to work.

### 5. Deploy a 1.X Pachyderm cluster with cloned bucket

Create a pachyderm cluster using the bucket you cloned in [3. Clone your object store bucket](#3-clone-your-object-store-bucket). 

You'll want to bring up this new pachyderm cluster in a different namespace.
You'll check at the steps below 
to see if there was some kind of problem with the extracted data 
and steps [2](#2-extract-a-pachyderm-backup-with-the-no-objects-flag) and
[3](#3-clone-your-object-store-bucket) need to be run again. 
Once your new cluster is up and you're connected to it, go on to the next step.

Note that there may be modifications needed to Kubernetes ingress to Pachyderm deployment in the new namespace to avoid port conflicts in the same cluster.
Please consult with your Kubernetes administrator for information on avoiding ingress conflicts,
or check with us in your Pachyderm support channel if you need help.

_Important: Use the_ `kubectl config current-config` _command to confirm you're talking to the correct kubernetes cluster configuration for the new cluster._

### 6. Restore the new 1.X Pachyderm cluster from your backup

Using the Pachyderm cluster you deployed in the previous step, [5. Deploy a 1.x Pachyderm cluster with cloned bucket](#5-deploy-a-1x-pachyderm-cluster-with-cloned-bucket), run `pachctl restore` with the backup you created in [2. Extract a pachyderm backup with the --no-objects flag](#2-extract-a-pachyderm-backup-with-the-no-objects-flag).

!!! note "Important"
    Use the_ `kubectl config current-config` _command to confirm you're
    talking to the correct kubernetes cluster configuration_.

`pachctl restore < path/to/your/backup/file`

You can also use the `-u` or `--url` flag to get the backup directly from the object store you placed it in

`pachctl restore --url s3://...`

Note that this s3 bucket is different than the s3 bucket you cloned, above. 
This is merely a bucket you allocated to hold the Pachyderm backup without objects.

### 7. Load transactional data from checkpoint into new cluster

Configure an instance of your data loading systems to point at the new, upgraded pachyderm cluster
and play back transactions from the checkpoint you established in [4. Restart all pipeline and data loading operations](#4-restart-all-pipeline-and-data-loading-ops).

Perform any reconfiguration to data loading or unloading operations.

Confirm that the data output is as expected and the new cluster is operating as expected.


### 8. Disable the old cluster

Once you've confirmed that the new cluster is operating, you can disable the old cluster.

### 9. Reconfigure new cluster as necessary

You may also need to reconfigure

- data loading operations from Pachyderm to processes outside of it to work as expected
- Kubernetes ingress and port changes taken to avoid conflicts with the old cluster
