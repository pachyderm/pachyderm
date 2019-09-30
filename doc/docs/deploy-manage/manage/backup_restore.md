# Backup and Restore

Pachyderm provides the `pachctl extract` and `pachctl restore` commands to backup and restore the state of a Pachyderm cluster.

The `pachctl extract` command requires that all pipeline and data loading activity into Pachyderm stop before the extract occurs.  This enables Pachyderm to create a consistent, point-in-time backup.  In this document, we'll talk about how to create such a backup and restore it to another Pachyderm instance.

Extract and restore commands are currently used to migrate between minor and major releases of Pachyderm, so it's important to understand how to perform them properly.   In addition, there are a few design points and operational techniques that data engineers should take into consideration when creating complex pachyderm deployments to minimize disruptions to production pipelines.

In this document, we'll take you through the steps to backup and restore a cluster, migrate an existing cluster to a newer minor or major release, and elaborate on some of those design and operations considerations.

## Backup and restore concepts
Backing up Pachyderm involves the persistent volume (PV) that `etcd` uses for administrative data
and the object store bucket that holds Pachyderm's actual data. 
Restoring involves populating that PV and object store with data to recreate a Pachyderm cluster.

## General backup procedure

### 1. Pause all pipeline and data loading/unloading operations

Before you begin, you need to pause all pipelines and data operations.

#### Pausing pipelines
From the directed acyclic graphs (DAG) that define your pachyderm cluster, stop each pipeline.  You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the `pachctl stop pipeline` command.

`pachctl stop pipeline <pipeline-name>`

You can confirm each pipeline is paused using the `pachctl list pipeline` command

`pachctl list pipeline`

Alternatively, a useful shell script for running `stop pipeline` on all pipelines is included below.   It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```
pachctl list pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl stop pipeline {}
```

It's also a useful practice, for simple to moderately complex deployments, to keep a terminal window up showing the state of all running kubernetes pods.

`watch -n 5 kubectl get pods`

You may need to install the `watch` and `kubectl` commands on your system, and configure `kubectl` to point at the cluster that Pachyderm is running in.

#### Pausing data loading operations

**Input repositories** or **input repos** in pachyderm are repositories created with the `pachctl create repo` command.  They're designed to be the repos at the top of a directed acyclic graph of pipelines. Pipelines have their own output repos associated with them, and are not considered input repos. If there are any processes external to pachyderm that put data into input repos using any method (the Pachyderm APIs, `pachctl put file`, etc.), they need to be paused.  See [Loading data from other sources into pachyderm](#loading-data-from-other-sources-into-pachyderm) below for design considerations for those processes that will minimize downtime during a restore or migration.

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
$ kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kubectl apply -f -

# Modify your environment so that *you* can talk to pachd on this new port
$ pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30649

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pachctl version
COMPONENT           VERSION
pachctl             1.9.5
pachd               1.9.5
```

### 2. Extract a pachyderm backup

You can use `pachctl extract` alone or in combination with cloning/snapshotting services.

#### Using `pachctl extract`

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

For on-premises Kubernetes deployments,  check the vendor documentation for your PV implementation on backing up and restoring.

##### Cloning object stores
Below, we give an example using the Amazon Web Services CLI to clone one bucket to another, [taken from the documentation for that command](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html).  Similar commands are available for [Google Cloud](https://cloud.google.com/storage/docs/gsutil/commands/cp) and [Azure blob storage](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).

`aws s3 sync s3://mybucket s3://mybucket2`

For on-premises Kubernetes deployments,  check the vendor documentation for your on-premises object store for details on  backing up and restoring a bucket.

#### Combining cloning, snapshots and extract/restore

You can use `pachctl extract` command with the `--no-objects` flag to exclude the object store, and use an object store snapshot or clone command to back up the object store. You can run the two commands at the same time.  For example, on Amazon Web Services, the following commands can be run simultaneously.

`aws s3 sync s3://mybucket s3://mybucket2`

`pachctl extract --no-objects --url s3://anotherbucket`

#### Use case: minimizing downtime during a migration
The above cloning/snapshotting technique is recommended when doing a migration where minimizing downtime is desirable, as it allows the duplicated object store to be the basis of the upgraded, new cluster instead of requiring Pachyderm to extract the data from object store.

### 3. Restart all pipeline and data loading operations

Once the backup is complete, restart all paused pipelines and data loading operations.

From the directed acyclic graphs (DAG) that define your pachyderm cluster, start each pipeline.    You can either run a multiline shell command, shown below, or you must, for each pipeline, manually run the `pachctl start pipeline` command.

`pachctl start pipeline <pipeline-name>`

You can confirm each pipeline is started using the `list pipeline` command

`pachctl list pipeline`

A useful shell script for running `start pipeline` on all pipelines is included below.  It may be necessary to install the utilities used in the script, like `jq` and `xargs`, on your system.

```
pachctl list pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl start pipeline {}
```

If you used the port-changing technique, [above](#pausing-data-loading-operations), to stop all data loading into Pachyderm from outside processes, you should change the ports back.

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
$ kubectl get svc/pachd -o json >pach_service_backup_30649.json

# Modify the service to accept traffic on port 30650, again
$ kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kubectl apply -f -

# Modify your environment so that *you* can talk to pachd on the old port
$ pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pachctl version
COMPONENT           VERSION
pachctl             1.9.5
pachd               1.9.5
```
## General restore procedure
### Restore your backup to a pachyderm cluster, same version

Spin up a Pachyderm cluster and run `pachctl restore` with the backup you created earlier.

`pachctl restore < path/to/your/backup/file`

You can also use the `-u` or `--url` flag to get the backup directly from the object store you placed it in

`pachctl restore --url s3://...`


### Loading data from other sources into Pachyderm

When writing systems that place data into Pachyderm input repos (see [above](#pausing-data-loading-operations) for a definition of 'input repo'), 
it is important to provide ways of 'pausing' output while queueing any data output requests to be output when the systems are 'resumed'.
This allows all Pachyderm processing to be stopped while the extract takes place.

In addition, it is desirable for systems that load data into Pachyderm have a mechanism for replaying a queue from any checkpoint in time.
This is useful when doing migrations from one release to another, where you would like to minimize downtime of a production Pachyderm system. 
After an extract, 
the old system is kept running with the checkpoint established while a duplicate, upgraded pachyderm cluster is migrated with duplicated data. 
Transactions that occur while the migrated, 
upgraded cluster is being brought up are not lost, 
