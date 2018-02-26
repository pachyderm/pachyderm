# Pachyderm Migrations

New versions of Pachyderm often require a migration for some or all of the on
disk objects. This document describes how Pachyderm migration works and
the best practices surrounding it.

As of 1.7, Pachyderm's migration works by extracting objects into a stream of
API requests, and replaying those requests onto the newer version of pachd.
This process happens automatically using Kubernetes' "rolling update"
functionality. All you need to do is deploy Pachyderm (with `pachctl deploy) as
you would if you were deploying for the first time. While migration is
running you will see 2 pachd pods running, the one that was already
running, and the new one. During this time the previous pod will still
respond to requests, however write operations will race with the migration
and may not make it to the new cluster. Thus you should make sure that all
processes that write data to repos (i.e., call put-file) or create new
pipelines are turned down before migration begins. You don't need to worry
about pipelines running during the migration process.

## Backups

It is highly recommended that you backup your cluster before you perform
a migration. This is accomplished with the `pachctl extract` command,
which will generate a stream of API requests, the same stream that migrate
uses, which will reconstruct your cluster. See the docs for [`pachctl
extract`](http://docs.pachyderm.io/en/latest/pachctl/pachctl_extract.html)
and [`pachctl
restore`](http://docs.pachyderm.io/en/latest/pachctl/pachctl_restore.html)
for further usage.


## Before You Migrate.

1.7 is the first Pachyderm version to support `extract` and `restore`
which are necessary for migration. To solve this we've made a final 1.6
release, 1.6.9 which backports the `extract` and `restore` functionality
to the 1.6 series of releases. 1.6.9 requires no migration from 1.6.8 you
can simply `pachctl undeploy and then `pachctl deploy` after upgrading
`pachctl` to version 1.6.9. After 1.6.9 is deployed you should make
a backup using `pachctl extract` then upgrade `pachctl` again, to 1.7.0
and do `pachctl deploy` to trigger the migration.
