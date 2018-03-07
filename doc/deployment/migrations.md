# Pachyderm Migrations

New versions of Pachyderm often require a migration for some or all of the
on disk objects which persist Pachyderm's metadata for commits, jobs, etc.
This document describes how Pachyderm migration works and the best
practices surrounding it.

## How To Migrate

As of 1.7, Pachyderm's migration works by extracting objects into a stream of
API requests, and replaying those requests onto the newer version of pachd.
This process happens automatically using Kubernetes' "rolling update"
functionality. All you need to do is upgrade Pachyderm (with `pachctl
deploy`) as further described [here](upgrading.html).
Generally, you will need to:

1. Have version 1.6.9 or later of Pachyderm up and running in Kubernetes.
2. (Optional, but recommended) Create a backup of your cluster state with
   `pachctl extract` (see [below](#backups)).
3. Upgrade `pachctl` (see [here](upgrading.html) for more details).
4. Run `pachctl deploy ...` with whatever arguments you used to deploy Pachyderm
   previously.

While the migration is running, you will see 2 `pachd` pods running, the one that was
already running and the new one. The original `pachd` pod (deployed with the previous version of Pachyderm) will
still respond to requests. However, write operations will race with the
migration and may not make it to the new cluster. Thus, **you should make sure
that all external processes that write data to repos (i.e., calls to `put-file`) or create new
pipelines are turned down before migration begins**. You don't need to worry
about pipelines running during the migration process.

## Backups

It is highly recommended that you backup your cluster before you perform
a migration. This is accomplished with the `pachctl extract` command. Running
this command will generate a stream of API requests, similar to the stream used
by migration above. This stream can then be used to reconstruct your cluster by
running `pachctl restore`. See the docs for [`pachctl
extract`](http://docs.pachyderm.io/en/latest/pachctl/pachctl_extract.html) and
[`pachctl
restore`](http://docs.pachyderm.io/en/latest/pachctl/pachctl_restore.html) for
further usage.


## Before You Migrate 1.6.x to 1.7.x+

1.7 is the first Pachyderm version to support `extract` and `restore` which are
necessary for migration. To bridge the gap to previous Pachyderm versions,
we've made a final 1.6 release, 1.6.9 which backports the `extract` and
`restore` functionality to the 1.6 series of releases. 1.6.9 requires no
migration from 1.6.8. You can simply `pachctl undeploy` and then `pachctl
deploy` after upgrading `pachctl` to version 1.6.9. After 1.6.9 is deployed you
should make a backup using `pachctl extract` and then upgrade `pachctl` again,
to 1.7.0. Finally you can `pachctl deploy ... ` with `pachctl` 1.7.0 to trigger
the migration.
