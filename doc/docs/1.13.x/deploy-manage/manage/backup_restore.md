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

1. Pause each pipeline individually by repeatedly running the single
`pachctl` command or by running a script:

=== "Command"
   ```shell
   pachctl stop pipeline <pipeline-name>
   ```

=== "Script"
   ```shell
   pachctl list pipeline --raw \
   | jq -r '.pipeline.name' \
   | xargs -P3 -n1 -I{} pachctl stop pipeline {}
   ```

1. Optionally, run the `watch` command to monitor the pods
   terminating:

   ```shell
   watch -n 5 kubectl get pods
   ```

1. Confirm that pipelines are paused:

   ```shell
   pachctl list pipeline
   ```

### Pause External Data Loading Operations

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
running will not interfere with the process. Use this port switching
technique to minimize downtime during the migration.

To pause external data loading operations, complete the
following steps:

1. Verify that all Pachyderm pipelines are paused:

   ```shell
   pachctl list pipeline
   ```

1. For safery, save the Pachyderm service spec in a `json`:

   ```shell
   kubectl get svc/pachd -o json >pach_service_backup_30650.json
   ```

1. Modify the `pachd` service to accept traffic on port 30649:

   ```shell
   kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kubectl apply -f -
   ```

   Most likely, you will need to modify your cloud provider's firewall
   rules to allow traffic on this port.

   Depending on your deployment, you might need to switch
   additional ports:

   1. Back up the `etcd` and dashboard manifests:

   ```shell
   kubectl get svc/etcd -o json >etcd_svc_backup_32379.json
   kubectl get svc/dash -o json >dash_svc_backup_30080.json
   ```

   1. Switch the `etcd` and dashboard manifests:

      ```shell
      kubectl get svc/pachd -o json | sed 's/30651/30648/g' | kubectl apply -f -
      kubectl get svc/pachd -o json | sed 's/30652/30647/g' | kubectl apply -f -
      kubectl get svc/pachd -o json | sed 's/30654/30646/g' | kubectl apply -f -
      kubectl get svc/pachd -o json | sed 's/30655/30644/g' | kubectl apply -f -
      kubectl get svc/etcd -o json | sed 's/32379/32378/g' | kubectl apply -f -
      kubectl get svc/dash -o json | sed 's/30080/30079/g' | kubectl apply -f -
      kubectl get svc/dash -o json | sed 's/30081/30078/g' | kubectl apply -f -
      kubectl get svc/pachd -o json | sed 's/30600/30611/g' | kubectl apply -f -
      ```

1. Modify your environment so that you can access `pachd` on this new
port

   ```shell
   pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30649
   ```

1. Verify that you can talk to `pachd`: (if not, firewall rules are a common culprit)

   ```shell
   pachctl version
   ```

   **System Response:**

   ```
   COMPONENT           VERSION
   pachctl             {{ config.pach_latest_version }}
   pachd               {{ config.pach_latest_version }}
   ```

??? note "pause-pipelines.sh"
    Alternatively, you can run **Steps 1 - 3** by using the following script:

    ```shell
    #!/bin/bash
    # Stop all pipelines:
    pachctl list pipeline --raw \
    | jq -r '.pipeline.name' \
    | xargs -P3 -n1 -I{} pachctl stop pipeline {}

    # Backup the Pachyderm services specs, in case you need to restore them:
    kubectl get svc/pachd -o json >pach_service_backup_30650.json
    kubectl get svc/etcd -o json >etcd_svc_backup_32379.json
    kubectl get svc/dash -o json >dash_svc_backup_30080.json

    # Modify all ports of all the Pachyderm service to avoid collissions
    # with the migration cluster:
    # Modify the pachd API endpoint to run on 30649:
    kubectl get svc/pachd -o json | sed 's/30650/30649/g' | kubectl apply -f -
    # Modify the pachd trace port to run on 30648:
    kubectl get svc/pachd -o json | sed 's/30651/30648/g' | kubectl apply -f -
    # Modify the pachd api-over-http port to run on 30647:
    kubectl get svc/pachd -o json | sed 's/30652/30647/g' | kubectl apply -f -
    # Modify the pachd saml authentication port to run on 30646:
    kubectl get svc/pachd -o json | sed 's/30654/30646/g' | kubectl apply -f -
    # Modify the pachd git api callback port to run on 30644:
    kubectl get svc/pachd -o json | sed 's/30655/30644/g' | kubectl apply -f -
    # Modify the etcd client port to run on 32378:
    kubectl get svc/etcd -o json | sed 's/32379/32378/g' | kubectl apply -f -
    # Modify the dashboard ports to run on 30079 and 30078:
    kubectl get svc/dash -o json | sed 's/30080/30079/g' | kubectl apply -f -
    kubectl get svc/dash -o json | sed 's/30081/30078/g' | kubectl apply -f -
    # Modify the pachd s3 port to run on 30611:
    kubectl get svc/pachd -o json | sed 's/30600/30611/g' | kubectl apply -f -
    ```

## Back up Your Pachyderm Cluster

After you pause all pipelines and external data operations,
you can use the `pachctl extract` command to back up your data.
You can use `pachctl extract` alone or in combination with
cloning or snapshotting services offered by your cloud provider.

The backup includes the following:

* Your data that is typically stored in an object store
* Information about Pachyderm primitives, such as pipelines, repositories,
commits, provenance and so on. This information is stored in etcd.

You can back up everything to one local file or you can back up
Pachyderm primitives to a local file and use object store's
capabilities to clone the data stored in object store buckets.
The latter is preferred for large volumes of data and minimizing
the downtime during the upgrade. Use the
`--no-objects` flag to separate backups.

In addition, you can extract your partial or full backup into a
separate S3 bucket. The bucket must have the same permissions policy as
the one you have configured when you originally deployed Pachyderm.

To back up your Pachyderm cluster, run one of the following commands:

* To create a partial back up of metadata-only, run:

  ```shell
  pachctl extract --no-objects > path/to/your/backup/file
  ```

  * If you want to save this partial backup in an object store by using the
  `--url` flag, run:

    ```shell
    pachctl extract --no-objects --url s3://...
    ```

* To back up everything in one local file:

  ```shell
  pachctl extract > path/to/your/backup/file
  ```

  Similarly, this backup can be saved in an object store with the `--url`
  flag.

## Using your Cloud Provider's Clone and Snapshot Services

Follow your cloud provider's recommendation
for backing up persistent volumes and object stores. Here are some pointers to the relevant documentation:

* Creating a snapshot of persistent volumes:

  - [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
  - [Creating snapshots of Elastic Block Store (EBS) volumes](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)
  - [Creating snapshots of Azure Virtual Hard Disk volumes](https://docs.microsoft.com/en-us/azure/virtual-machines/snapshot-copy-managed-disk)

    For on-premises Kubernetes deployments, check the vendor documentation for
    your PV implementation on backing up and restoring.

* Cloning object stores:

  - [Using AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/s3/sync.html)
  - [Using gsutil](https://cloud.google.com/storage/docs/gsutil/commands/cp)
  - [Using azcopy](https://docs.microsoft.com/en-us/azure/storage/common/storage-use-azcopy-linux?toc=%2fazure%2fstorage%2ffiles%2ftoc.json).

    For on-premises Kubernetes deployments, check the vendor documentation
    for your on-premises object store for details on  backing up and
    restoring a bucket.

# Restore your Cluster from a Backup:

After you backup your cluster, you can restore it by using the
`pachctl restore` command. Typically, you would deploy a new Pachyderm cluster
either in another Kubernetes namespace or in a completely separate Kubernetes cluster.

To restore your Cluster from a Backup, run the following command:

* If you have backed up your cluster to a local file:, run:

  ```shell
  pachctl restore < path/to/your/backup/file
  ```

* If you have backed up your cluster to an object store, run:

  ```shell
  pachctl restore --url s3://<path-to-backup>>
  ```

!!! note "See Also:"
    - [Migrate Your Cluster](../migrations/)
