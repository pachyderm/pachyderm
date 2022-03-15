# Backup Your Cluster

This page will walk you through the main steps required
to manually back up and restore the state of a Pachyderm cluster.

Details on how to perform those steps might vary depending
on your infrastructure and cloud provider / on-premises setup. 

Refer to your provider's documentation.

!!! Info    
    For additional questions about backup / restore, you can post them in the community #help channel on [Slack](https://www.pachyderm.com/slack/){target=_blank}, or reach out to your TAM if you are an Enterprise customer.

## Overview

Pachyderm state is stored in three places 
(See our high-level [instructure diagram](../../../deploy-manage/#overview)):

- an **object-store** holding Pachyderm's data.
- a PostgreSQL instance made up of **two databases**: `pachyderm` holding Pachyderm's metadata and `dex` holding authentication data. 
- and in Kubernetes itself (`etcd`). This last item is automatically managed using Helm, **assuming that you have retained a copy of the initialHelm values are used to deploy the cluster you are planning to backup**. 

Backing up a Pachyderm cluster involves snapshotting both
the object store and the PostgreSQL databases (see above),
in a consistent state, at a given point in time.

Restoring it involves populating the databases in a new PostgreSQL instance and a new object store using those backups, then recreating a Pachyderm cluster.

!!! Important
    Make sure that you have a bucket for backup use, separate from the object store used by your cluster.
## Manual Back Up Of A Pachyderm Cluster

Before any manual backup, you need to suspend any state-mutating operations.

!!! Note 

    - **Backups incur downtime** until operations are resumed.
    - Operational best practices include notifying Pachyderm users of the outage and providing an estimated time when downtime will cease.  
    - Downtime duration is a function of the size of the data be to backed up and the
    networks involved; Testing before going into production and monitoring backup times on an ongoing basis might help make accurate predictions.

### Suspend Operations

- Pause any external automated process ingressing data to Pachyderm input repos,
 or queue/divert those as they will fail to connect to the cluster while the backup occurs.

- Suspend all mutation of state by scaling `pachd` and the worker pods down:

      ```shell 
      kubectl scale deployment pachd --replicas 0 
      kubectl scale rc --replicas 0 -l suite=pachyderm,component=worker
      ```

      Note that it takes some time for scaling down to take effect;

      Run the `watch` command to monitor the state of `pachd` and worker pods terminating:

      ```shell
      watch -n 5 kubectl get pods
      ```
      And confirm that the pipelines are paused:

      ```shell
      pachctl list pipeline
      ```

### Back Up The Databases And The Object Store

This step is specific to your database and object store hosting.

- If your PostgreSQL instance is solely dedicated to Pachyderm, 
you can use `pg_dumpall` to dump your entire PostgreSQL state.  
    Otherwise, use targeted `pg_dump` commands to dump the
    `pachyderm` and `dex` databases.  

    Note that if you are using a cloud provider, you might
    choose to use the provider’s method of making PostgreSQL backups.

    Here are some pointers to the relevant documentation:

      - [PostgreSQL on AWS RDS backup](https://aws.amazon.com/backup/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc){target=_blank}
      - [GCP Cloud SQL backup](https://cloud.google.com/sql/docs/postgres/backup-recovery/backing-up){target=_blank}
      - [Azure Database for PostgreSQL backup](https://docs.microsoft.com/en-us/azure/backup/backup-azure-database-postgresql){target=_blank}

        For on-premises Kubernetes deployments, check the vendor documentation
        for your on-premises PostgreSQL for details on backing up and restoring your databases.

- To back up the object store, you can either download all objects or
use the object store provider’s backup method.  
    The latter is preferable since it will typically not incur egress costs.

    Here are some pointers to the relevant documentation:

      - [AWS backup for S3](https://aws.amazon.com/backup/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc){target=_blank}
      - [GCP Cloud storage bucket backup](https://cloud.google.com/storage-transfer/docs/overview){target=_blank}
      - [Azure blob backup](https://docs.microsoft.com/en-us/azure/backup/blob-backup-configure-manage){target=_blank}

        For on-premises Kubernetes deployments, check the vendor documentation
        for your on-premises object store for details on backing up and
        restoring a bucket.

### Resuming operations

Once your backup is completed, resume normal operations by scaling `pachd` back up.  It will take care
of restoring the worker pods.

```sh
 kubectl scale deployment pachd --replicas 1
```

## Restore Pachyderm

The most straightforward way to restore a Pachyderm cluster is to:
###  Create A New PostgreSQL Instance, A New Bucket, A New K8s Cluster

- Create a new PostgreSQL instance with pachyderm and dex databases
- Create a new bucket or use the backed-up object
store (note that, in that case, it will no longer be a backup)
- Create a new empty Kubernetes cluster and give it access to your databases and new bucket

!!! Info
    Find the detailed installations instructions of your PostgreSQL instance, bucket, Kubernetes cluster, and permissions setup for each Cloud Provider in the [Deploy section of our Documentation](../../../deploy-manage/deploy/){target=_blank}

### Restore The Databases And Objects

- Restore PostgreSQL backups into your new databases using the appropriate
method (this is most straightforward when using a cloud provider).
- Copy the objects from the backed-up object store to your new bucket.

### Deploy Pachyderm Into The New Cluster

Finally, make a copy of your original Helm values,
update them accordingly to point Pachyderm at the new databases and the
new object store, then use Helm to install
Pachyderm into the new cluster.

!!! Info
    The values needing an update and deployment instructions are detailed in the Chapter 6 of all our cloud  installation pages. For example, in the case of GCP, [check the `deploy Pachyderm` chapter](../../../deploy-manage/deploy/aws-deploy-pachyderm/#6-deploy-pachyderm){target=_blank}

### [Have 'pachctl' And Your Restored Cluster Communicate](../../../deploy-manage/deploy/aws-deploy-pachyderm/#7-have-pachctl-and-your-cluster-communicate){target=_blank}

...and [check that your cluster is up and running](../../../deploy-manage/deploy/aws-deploy-pachyderm/#8-check-that-your-cluster-is-up-and-running).
