# Pachyderm Migrations

This document describes the procedures for and best practices surrounding
upgrading a Pachyderm cluster to a newer version

# Types of Migrations
You'll need to follow one of two procedures when upgrading a Pachyderm cluster:

1. Simple migration path (for minor versions, e.g. v1.8.2 to v1.8.3)
1. Full migration (for major versions, e.g. v1.7.11 to v1.8.0)

The simple migration path is shorter and faster than the full migration path
because it can reuse Pachyderm's existing internal state.

### Difference in Detail
In general, Pachyderm stores all of its state in two places: etcd (which in
turn stores its state in one or more persistent volumes, which were created when
the Pachyderm cluster was deployed) and an object store bucket (in e.g. S3).

In the simple migration path, you only need to deploy new Pachyderm pods that
reference the existing data, and the existing data structures residing there
will be re-used. In the full migration path, however, those data structures need
to be read, transformed, and rewritten, so the process involves:

1. bringing up a new Pachyderm cluster adjacent to the old pachyderm cluster
1. exporting the old pachdyerm cluster's repos, pipelines, and input commits
1. importing the old cluster's repos, commits, and pipelines into the new
   cluster, and then re-running all of the old cluster's pipelines from scratch
   once.

Most Pachyderm 1.8 data structures are substantively different from their 1.7
counterparts, and in particular, Pachyderm v1.8+ stores more information in its
output commits than its predecessors, which enables useful performance
improvements. Old (v1.7) Pachyderm output commits don't have this extra
information, so Pachyderm 1.8 must re-create its output commits

# Backups
We recommend starting any migration effort by backing up pachyderm's
pre-migration state before you start. Specifically, this means backing up the
persistent volume that etcd uses and the object store bucket that holds
pachyderm's actual data. You should follow your cloud provider's recommendation
for backing up these resources. For example, here are official guides on backing
up persistent volumes on Google Cloud Platform and AWS, respectively:

- [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Creating snapshots of Elastic Block Store (EBS) volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

# Simple Migration

In the simple migration path, pachd and pipeline pods running the old version of
Pachyderm need to be replaced by pachd pods running the new version of
Pachyderm. The new pachd pods can talk to the same etcd instances and use the
same object storage bucket as the old version of Pachyderm. This makes for a
particularly fast and lightweight migration, as it can happen in-place.

All that's necessary is to replace the old pachd pods is:

```
# 1. make sure you have a new version of pachctl installed locally
$ pc version
COMPONENT           VERSION
pachctl             1.8.4
pachd               1.8.3

# 2. Update the pachd deployment in your cluster.
# TODO(msteffen): will this mess up dynamically provisioned etcd volumes?
# We could also update just the pachd deployment, but that wouldn't upgrade
# workers or sidecar containers
$ pachctl undeploy
$ pachctl deploy
```

# Full Migration

At a high level, this process is just: 1. call 'pachctl extract' on your
existing cluster, and then 2. call 'pachctl restore' on your new cluster. Below
are some of the best practices we've found for accomplishing this.

### Stopping your cluster

`pachctl extract` won't run if there are any active pipelines in your cluster.
Therefore, you'll have to turn down your existing cluster:

```
# Stop all running pachyderm pipelines:
$ pc list-pipeline --raw \
  | jq -r '.pipeline.name' \
  | xargs -P3 -n1 -I{} pachctl stop-pipeline {}

# At this point, all pipelines in your cluster should be suspended. To stop all
# ingress processes, we're going to modify the pachd Kubernetes service so that
# it only accepts traffic on port 30649 (instead of the usual 30650). This way,
# any background users and services that send requests to your Pachyderm cluster
# while 'extract' is running will not interfere with the process
#
# Backup the Pachyderm service spec, in case you need to restore it quickly
$ kubectl get svc/pach -o json >pach_service_backup.json

# Modify the service to accept traffic on port 30649
# Note that you'll likely also need to modify your cloud provider's firewall
# rules to allow traffic on this port
$ kc get svc/pachd -o json | sed 's/30650/30649/g' | kc apply -f -

# Modify your environment so that *you* can talk to pachd on this new port
$ export PACHD_ADDRESS="${PACHD_ADDRESS/30650/30649}"

# Make sure you can talk to pachd (if not, firewall rules are a common culprit)
$ pc version
COMPONENT           VERSION
pachctl             1.7.11
pachd               1.7.11
```

### Extracting Data

Once your cluster is effectively drained, you can run 'pachctl extract'. You
might prefer to send your cluster's metadata to an object store bucket, to speed
up the extraction process and preserve disk space on your workstation:
```
$ pachctl extract -u s3://my-bucket/migration-data

# Make sure you successfully extracted data from your cluster
$ aws s3 ls --human-readable s3://my-bucket
```

### Restoring data

You'll want to bring up a new pachyderm cluster (perhaps in a different
namespace) in case there was some kind of problem with the extracted data and
'pachctl extract' needs to be run again. Once your new cluster is up and you're
connected to it, run:
```
$ pachctl restore -u s3://my-bucket/migration-data
```

Wait for all piplines to run once, inspect your data to make sure it's as
expected, and then proceed to use your upgraded cluster

## Note: Migrating 1.6.x to 1.7.x+

1.7 is the first Pachyderm version to support `extract` and `restore` which are
necessary for migration. To bridge the gap to previous Pachyderm versions,
we've made a final 1.6 release, 1.6.10, which backports the `extract` and
`restore` functionality to the 1.6 series of releases. 1.6.10 requires no
migration from other 1.6.x versions. You can simply `pachctl undeploy` and then `pachctl
deploy` after upgrading `pachctl` to version 1.6.10. After 1.6.10 is deployed you
should make a backup using `pachctl extract` and then upgrade `pachctl` again,
to 1.7.0. Finally you can `pachctl deploy ... ` with `pachctl` 1.7.0 to trigger
the migration.
