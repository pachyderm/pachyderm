# Backing up and restoring Pachyderm

Pachyderm state is stored in three places: the database; the object
store; and in Kubernetes itself.  Accordingly, in order to back it up
one will need to suspend any state-mutating operations, back up the
the database, the object store & Kubernetes, then resume normal
operations.

## Backing up Pachyderm

### Suspending operations

To suspend mutation of state, one simply scales `pachd` and the worker
pods down:

```sh
$ kubectl scale deployment pachd --replicas 0
$ kubectl scale rc --replicas 0 -l suite=pachyderm,component=worker
```

Note that this incurs downtime until operations are resumed;
operational best practices include notifying Pachyderm users of the
outage and an estimated time when downtime will cease.

Also of note, it takes some time for scaling down to take effect;
`kubectl get po` can be used to monitor the state of the `pachd` and
worker pods.

### Backing up database and object store

This is specific to your database and object store hosting.  If
running one’s own PostgreSQL, one can use `pg_dumpall` to dump your
entire PostgreSQL state, or targeted `pg_dump` commands to dump the
Pachyderm and Dex databases.  If using a cloud provider, one may
choose to use the provider’s method of making PostgreSQL backups.

To back up the object store, one could either download all objects or
use the object store provider’s backup method.  The latter is
preferable, since it will typically not incur egress costs.

### Backing up Kubernetes

There are a number of products which enable one to back up Kubernetes
objects.  One product which is free software and appears to work well
is Velero.  There is some cloud-specific configuration when installing
Velero, but once it is installed and properly configured a backup is
as simple as:

```sh
$ velero backup create NAME --include-namespaces PACHD-NAMESPACE
```

Where `NAME` is a unique name for this backup (e.g. one including the
date or a UUID) and `PACHD-NAMESPACE` is the namespace Pachyderm is
installed in (typically `default`).

Of note, Velero backups are asynchronous; you will need to check on
the progress of a backup with `velero backup describe NAME` to see
when it has completed.

### Resuming operations

To resume normal operations, scale `pachd` back up.  It will take care
of restoring the worker pods.

```sh
$ kubectl scale deployment pachd --replicas 1
```

## Restoring Pachyderm

The most straightforward way to restore a Pachyderm cluster is to
create a new empty cluster with a new PostgreSQL database, then
restored into it.  You may or may not choose to use the backed-up
object store, or a new one with the objects copied into it (but not
that if you use the backed-up store it will no longer be a backup!).
The process of restoring into it involves multiple components:
PostgreSQL, object store, Kubernetes objects, to include editing some
Kubernetes objects for the new cluster.

One will need to restore PostgreSQL backups, using the appropriate
method (this is most straightforward when using a cloud provider).
One will also need to point Pachyderm at the new database and the
backed-up object store (or a new object store).  Finally, one may need
to edit service accounts, depending on the cloud provider.  To do
this, first restore the Kubernetes objects (e.g. with `velero restore
create --from-backup NAME`.  Then, to update the database, run
`kubectl edit deployment cloudsql-auth-proxy` and edit the line beging
with `- -instances=` to point to the new database.  To edit the
bucket, run `kubectl edit secret pachyderm-storage-secret` and update
the Base64-encoded bucket.  There may be provider-specific service
accounts to update or delete.  E.g, these are the edits to update GCP
service accounts:

 - If using a new database:
     - `kubectl edit serviceaccount k8s-cloudsql-auth-proxy` (update
         `iam.gke.io/gcp-service-account` and delete `secrets`)
     - `kubectl edit serviceaccount pachyderm` (update
       `iam.gke.io/gcp-service-account` and delete `secrets`)
 - `kubectl edit serviceaccount pachyderm` (update
   `iam.gke.io/gcp-service-account` and delete `secrets`)
 - `kubectl edit serviceaccount pachyderm-worker` (update
   `iam.gke.io/gcp-service-account` and delete `secrets`)

### Connect to a restored cluster

Get the static IP address of your new cluster, then:

  - `echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite`
  - `pachctl config set active-context ${CLUSTER_NAME}`
