# Backup Restore Your Cluster

This page will walk you through the main steps required
to manually back up and restore the state of a Pachyderm cluster in production.

Details on how to perform those steps might vary depending
on your infrastructure and cloud provider / on-premises setup. 

Refer to your provider's documentation.

## Overview

Pachyderm state is stored in two main places
(See our high-level [architecture diagram](../../../deploy-manage/#overview)):

- an **object-store** holding Pachyderm's data.
- a PostgreSQL instance made up of **two databases**: `pachyderm` holding Pachyderm's metadata and `dex` holding authentication data. 

Backing up a Pachyderm cluster involves snapshotting both
the object store and the PostgreSQL databases (see above),
in a consistent state, at a given point in time.

Restoring it involves re-populating the databases and the object store using those backups, then recreating a Pachyderm cluster.

!!! Note
    - Make sure that you have a bucket for backup use, 
    separate from the object store used by your cluster.
    - Depending on the reasons behind your cluster recovery, you might choose to use an existing vs. a new instance of PostgreSQL and/or the object store.

## Manual Back Up Of A Pachyderm Cluster

Before any manual backup:

- Make sure to retain a copy of the Helm values used to deploy your cluster.
- Then, suspend any state-mutating operations.

!!! Note 

    - **Backups incur downtime** until operations are resumed.
    - Operational best practices include notifying Pachyderm users of the outage and providing an estimated time when downtime will cease.  
    - Downtime duration is a function of the size of the data be to backed up and the
    networks involved; Testing before going into production and monitoring backup times on an ongoing basis might help make accurate predictions.


### Suspend Operations

- **Pause any external automated process ingressing data to Pachyderm input repos**,
 or queue/divert those as they will fail to connect to the cluster while the backup occurs.

- **Suspend all mutation of state by scaling `pachd` and the worker pods down**:

    Before starting, make sure that your context points to the server you want to pause by running `pachctl config get active-context`. Find more information on how to [set your context](../../deploy/quickstart/#4-have-pachctl-and-your-cluster-communicate){target=_blank} in our deployment section.

    To pause Pachyderm, **run the `pachctl pause` command**. 

    !!! Tip "Alternatively, you can use `kubectl`"

         Before starting, make sure that `kubectl` [points to the right cluster](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/){target=_blank}.
         Run `kubectl config get-contexts` to list all available clusters and contexts (the current context is marked with a `*`), then `kubectl config use-context <your-context-name>` to set the proper active context.

         ```shell 
         kubectl scale deployment pachd --replicas 0 
         kubectl scale rc --replicas 0 -l suite=pachyderm,component=worker
         ```

         Note that it takes some time for scaling down to take effect;

         Run the `watch` command to monitor the state of `pachd` and worker pods terminating:

         ```shell
         watch -n 5 kubectl get pods
         ```

### Back Up The Databases And The Object Store

This step is specific to your database and object store hosting. 

- If your PostgreSQL instance is solely dedicated to Pachyderm, 
you can use PostgreSQL's tools, like `pg_dumpall`, to dump your entire PostgreSQL state.  

    Alternatively, you can use targeted `pg_dump` commands to dump the
    `pachyderm` and `dex` databases, or use your Cloud Provider's backup product.  
    In any case, make sure to use TLS.
    Note that if you are using a cloud provider, you might
    choose to use the provider’s method of making PostgreSQL backups.
    
    !!! Warning "Reminder"
        A production setting of Pachyderm implies that you are running a managed PostgreSQL instance.

    !!! Info "Here are some pointers to the relevant documentation"

         - [PostgreSQL on AWS RDS backup](https://aws.amazon.com/backup/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc){target=_blank}
         - [GCP Cloud SQL backup](https://cloud.google.com/sql/docs/postgres/backup-recovery/backing-up){target=_blank}
         - [Azure Database for PostgreSQL backup](https://docs.microsoft.com/en-us/azure/backup/backup-azure-database-postgresql){target=_blank}

         For on-premises Kubernetes deployments, check the vendor documentation
         for your on-premises PostgreSQL for details on backing up and restoring your databases.

- To back up the object store, you can either download all objects or
use the object store provider’s backup method.  
    The latter is preferable since it will typically not incur egress costs.

    !!! Info "Here are some pointers to the relevant documentation"

         - [AWS backup for S3](https://aws.amazon.com/backup/?whats-new-cards.sort-by=item.additionalFields.postDateTime&whats-new-cards.sort-order=desc){target=_blank}
         - [GCP Cloud storage bucket backup](https://cloud.google.com/storage-transfer/docs/overview){target=_blank}
         - [Azure blob backup](https://docs.microsoft.com/en-us/azure/backup/blob-backup-configure-manage){target=_blank}

         For on-premises Kubernetes deployments, check the vendor documentation
         for your on-premises object store for details on backing up and
         restoring a bucket.

### Resuming operations

Once your backup is completed, **run `pachctl unpause` to resume your normal operations** by scaling `pachd` back up. It will take care of restoring the worker pods. 

!!! Tip "Alternatively, if you used `kubectl`"

    ```sh
    kubectl scale deployment pachd --replicas 1
    ```

## Restore Pachyderm

There are two primary use cases for restoring a cluster:

1. Your data have been corrupted, preventing your cluster from functioning correctly. You want the same version of Pachyderm re-installed on the latest uncorrupted data set. 
1. You have upgraded a cluster and are encountering problems. You decide to uninstall the current version and restore the latest backup of a previous version of Pachyderm.

Depending on your scenario, pick all or a subset of the following steps:

- Populate new `pachyderm` and `dex` databases on your PostgreSQL instance
- Populate a new bucket or use the backed-up object-store (note that, in that case, it will no longer be a backup)
- Create a new empty Kubernetes cluster and give it access to your databases and bucket
- Deploy Pachyderm into your new cluster

!!! Info
    Find the detailed installations instructions of your PostgreSQL instance, bucket, Kubernetes cluster, permissions setup, and Pachyderm deployment for each Cloud Provider in the [Deploy section of our Documentation](../../../deploy-manage/deploy/){target=_blank}

### Restore The Databases And Objects

- Restore PostgreSQL backups into your new databases using the appropriate
method (this is most straightforward when using a cloud provider).
- Copy the objects from the backed-up object store to your new bucket or re-use your backup.

### Deploy Pachyderm Into The New Cluster

Finally, update the copy of your original Helm values to point Pachyderm to the new databases and the new object store, then use Helm to install
Pachyderm into the new cluster.

!!! Info
    The values needing an update and deployment instructions are detailed in the Chapter 6 of all our cloud  installation pages. For example, in the case of GCP, [check the `deploy Pachyderm` chapter](../../../deploy-manage/deploy/aws-deploy-pachyderm/#6-deploy-pachyderm){target=_blank}

### [Connect 'pachctl' To Your Restored Cluster](../../../deploy-manage/deploy/aws-deploy-pachyderm/#7-have-pachctl-and-your-cluster-communicate){target=_blank}

...and [check that your cluster is up and running](../../../deploy-manage/deploy/aws-deploy-pachyderm/#8-check-that-your-cluster-is-up-and-running).

## Backup/Restore A Stand-Alone Enterprise Server
Backing up / restoring an Enterprise Server is similar to the back up / restore of a regular cluster (see above), with two slight variations:

1. The name of its Kubernetes deployment is `pach-enterprise` versus `pachd` in the case of a regular cluster.
1. The Enterprise Server does not use an Object Store.


### Backup A Standalone Enterprise Server

- Make sure that `pachctl/kubectl` are pointing to the right cluster. Check your [Enterprise Server](../../../enterprise/auth/enterprise-server/setup/){target=_blank} context: `pachctl config get active-enterprise-context`, or `pachctl config set active-enterprise-context <my-enterprise-context-name> --overwrite` to set it.

- [Pause the Enterprise Server](#suspend-operations) like you would pause a regular cluster by running `pachctl pause`. Make sure that [your active context points to the right cluster](#suspend-operations){target=_blank} first.

    !!! Tip "Alternatively, you can use `kubectl`"
         Note that there is a difference with the pause of a regular cluster. The deployment of the enterprise server is named `pach-enterprise`; therefore, the first command should be:

         ```shell
         kubectl scale deployment pach-enterprise --replicas 0 
         ``` 

         There is no need to pause all the Pachyderm clusters registered to the Enterprise Server to backup the enterprise server; however, pausing the Enterprise server will result in your clusters becoming unavailable.

- As a reminder, the Enterprise Server does not use any object-store. Therefore, the [backup of the Enterprise Server](#back-up-the-databases-and-the-object-store) only consists in backing up the databases.

- [Resume the operations on your Enterprise Server](#resuming-operations) by running `pachctl unpause` to scale the `pach-enterprise` deployment back up: 

    !!! Tip "Alternatively, if you used `kubectl`"
        ```shell
        kubectl scale deployment pach-enterprise --replicas 1
        ```

### Restore An Enterprise Server

- [Follow the steps above](#restore-pachyderm) while skipping all tasks related to creating and populating a new object-store.

- Once your cluster is up and running, check that all [your clusters are automatically registered with your new Enterprise Server](../../../enterprise/auth/enterprise-server/manage/#list-all-registered-clusters){target=_blank}.

## Additional Info
   
For additional questions about backup / restore, you can post them in the community #help channel on [Slack](https://www.pachyderm.com/slack/){target=_blank}, or reach out to your TAM if you are an Enterprise customer.
