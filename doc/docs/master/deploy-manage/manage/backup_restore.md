# Backup Your Cluster

Pachyderm provides the `pachctl extract` and `pachctl restore` commands to
back up and restore the state of a Pachyderm cluster.

The `pachctl extract` command requires that all pipeline and data loading
activity into Pachyderm stop before the extract occurs. This enables
Pachyderm to create a consistent, point-in-time backup.

Extract and restore commands are used to migrate between minor
and major releases of Pachyderm. In addition, there are a few design
points and operational techniques that data engineers should take
into consideration when creating complex pachyderm deployments to
minimize disruptions to production pipelines.

Backing up Pachyderm involves the persistent volume (PV) that
`etcd` uses for administrative data and the object store bucket that
holds Pachyderm's actual data.
Restoring involves populating that PV and object store with data to
recreate a Pachyderm cluster.

## Before You Begin

Before you begin, you need to pause all the pipelines and data operations
that run in your cluster. You can do so either by running a multi-line
shell script or by running the `pachctl stop pipeline` command for each
pipeline individually.

If you decide to use a shell script below, you need to have `jq` and
`xargs` installed on your system. Also, you might need to install
the `watch` and `kubectl` commands on your system, and configure
`kubectl` to point at the cluster that Pachyderm is running in.

To stop a running pipeline, complete the following steps:

1. Stop each pipeline individually by repeatedly running the following
command:

   ```bash
   pachctl stop pipeline <pipeline-name>
   ```

   * Alternatively, use this shell script that pauses all pipelines at
   once:

   ```bash
    pachctl list pipeline --raw \
   | jq -r '.pipeline.name' \
   | xargs -P3 -n1 -I{} pachctl stop pipeline {}
   ```

   * Optionally, run the `watch` command to monitor the pods
   terminating:

   ```bash
   watch -n 5 kubectl get pods
   ```

1. Confirm that pipelines are paused:

   ```bash
   pachctl list pipeline
   ```

## Pause External Data Loading Operations

**Input repositories** or **input repos** in pachyderm are
repositories created with the `pachctl create repo` command.
They are designed to be the repos at the top of a directed
acyclic graph of pipelines. Pipelines have their own output
repos associated with them. These repos are different from
input repos.

If you have any processes external to Pachyderm
that put data into input repos using any supported method,
such as the Pachyderm APIs, `pachctl put file`, or other,
you need to pause those processes.

When an external system writes data into Pachyderm
input repos, you need to provide ways of *pausing*
output while queueing any data output
requests to be output when the systems are *resumed*.
This allows all Pachyderm processing to be stopped while
the extract takes place.

In addition, it is desirable for systems that load data
into Pachyderm have a mechanism for replaying a queue
from any checkpoint in time.
This is useful when doing migrations from one release
to another, where you want to minimize downtime
of a production Pachyderm system. After an extract,
the old system is kept running with the checkpoint
established while a duplicate, upgraded Pachyderm
cluster is being migrated with duplicated data.
Transactions that occur while the migrated,
upgraded cluster is being brought up are not lost.

If you are not using any external way of pausing input
from internal systems, you can use the following commands to stop
all data loading into Pachyderm from outside processes.
To stop all data loading processes, you need to modify
the `pachd` Kubernetes service so that it only accepts
traffic on port 30649 instead of the usual 30650. This way,
any background users and services that send requests to
your Pachyderm cluster while `pachctl extract` is
running will not interfere with the process.

To pause external data loading operations, complete the
following steps:

1. Verify that all Pachyderm pipelines are paused:

   ```bash
   pachctl list pipeline
   ```

1. For safery, save the Pachyderm service spec in a `json`:

   ```bash
   kubectl get svc/pachd -o json >pach_service_backup_30650.json
   ```

1. Modify the `pachd` service to accept traffic on port 30649:

   ```bash
   kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kubectl apply -f -
   ```

   Most likely, you will need to modify your cloud provider's firewall
   rules to allow traffic on this port.

1. Modify your environment so that you can access `pachd` on this new
port

   ```bash
   pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30649
   ```

1. Verify that you can talk to `pachd`: (if not, firewall rules are a common culprit)

   ```bash
   pachctl version
   COMPONENT           VERSION
   pachctl             1.10.0
   pachd               1.10.0
   ```

## Back up Your Pachyderm Cluster

After you pause all your pipelines and external data operations,
you can use the `pachctl extract` command to back up your data.
You can use `pachctl extract` alone or in combination with
cloning or snapshotting services.

The backup includes the following:

* Your data that is typically stored in an object store
* Information about Pachyderm primitives, such as pipelines, repositories,
commits, provenance and so on. This information is stored in etcd.

You can back up everything in one local file or you can back up
Pachyderm primitives in the local file and use object store's
capabilities to clone the data itself. The latter is preferred for
large volumes of data. Use the `--no-objects` flag to separate
backups.

In addition, you can extract your partial or full backup into a
separate S3 bucket. The bucket must have the same permissions policy as
the one you have configured when you originally deployed Pachyderm.

To back up your Pachyderm cluster, run one of the following commands:

* To create a partial back up of metada-only, run:

  ```bash
  pachctl extract --no-objects > path/to/your/backup/file
  ```

  * If you want to save this partial backup in an object store by using the
  `--url` flag, run:

    ```bash
    pachctl extract --url s3://...
    ```

* To back up everything in one local file:

  ```bash
  pachctl extract > path/to/your/backup/file
  ```

  Similarly, this backup can be saved in an object store with the `--url`
  flag.

## Using your cloud provider's clone and snapshot services

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
# pachctl list pipeline --raw \
#   | jq -r '.pipeline.name' \
#   | xargs -P3 -n1 -I{} pachctl start pipeline {}

# all pipelines in your cluster should be restarted. To restart all data loading 
# processes, we're going to change the pachd Kubernetes service so that
# it only accepts traffic on port 30650 again (from 30649). 
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
kubectl get svc/pachd -o json >pach_service_backup_30649.json

# Modify the service to accept traffic on port 30650, again
kubectl get svc/pachd -o json | sed 's/30649/30650/g' | kubectl apply -f -

# Modify your environment so that *you* can talk to pachd on the old port
pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
pachctl version
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

