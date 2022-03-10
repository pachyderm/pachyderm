# Backing up and restoring Pachyderm

Pachyderm state is stored in three places: the database; the object
store; and in Kubernetes itself.  The last of these can be managed
using Helm, assuming that one has retained a copy of one’s initial
Helm values.  Accordingly, in order to back a Pachyderm cluster up
one will need to suspend any state-mutating operations, back up the
the database & the object store, then resume normal operations.

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
outage and an estimated time when downtime will cease.  Downtime
duration is a function of the size of the data be to backed up and the
networks involved; it may be wise to test this prior to going into
production, and monitor backup times on an ongoing basis in order to
make accurate predictions.

Also of note, it takes some time for scaling down to take effect;
`kubectl get pods` can be used to monitor the state of the `pachd` and
worker pods.

### Backing up database and object store

This is specific to one’s database and object store hosting.  If
running one’s own PostgreSQL, one can use `pg_dumpall` to dump one’s
entire PostgreSQL state, or targeted `pg_dump` commands to dump the
Pachyderm and Dex databases.  If using a cloud provider, one may
choose to use the provider’s method of making PostgreSQL backups.

To back up the object store, one could either download all objects or
use the object store provider’s backup method.  The latter is
preferable, since it will typically not incur egress costs.

### Resuming operations

To resume normal operations, scale `pachd` back up.  It will take care
of restoring the worker pods.

```sh
$ kubectl scale deployment pachd --replicas 1
```

## Restoring Pachyderm

The most straightforward way to restore a Pachyderm cluster is to
create a new empty cluster and a new PostgreSQL database, then restore
into them.  One may or may not choose to use the backed-up object
store, or a new one with the objects copied into it (but not that if
one uses the backed-up store it will no longer be a backup!).

One will need to restore PostgreSQL backups, using the appropriate
method (this is most straightforward when using a cloud provider).
One will also need to point Pachyderm at the new database and the
new object store (or the backed-up object store, if that is the course
one chooses).

Finally, one will need to make a copy of the original Helm values and
update them with the new information and then use Helm to install
pachyderm into the new cluster.

### Connect to a restored cluster

Get the static IP address of your new cluster, then:

  - `echo "{\"pachd_address\": \"grpc://${STATIC_IP_ADDR}:30650\"}" | pachctl config set context "${CLUSTER_NAME}" --overwrite`
  - `pachctl config set active-context ${CLUSTER_NAME}`
