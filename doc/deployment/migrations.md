# Migrations

Occationally, Pachyderm introduces changes that are backward-incompatible: repos/commits/files created on an old version of Pachyderm may be unusable on a new version of Pachyderm. When that happens, we try our best to write a migration script that "upgrades" your data so it’s usable by the new version of Pachyderm.

## Migrate to 1.4.x

To migrate to 1.4.x, look under the directory named `migration/X-Y`. For instance, to upgrade from 1.3.12 to 1.4.0, look under `migration/1.3.12-1.4.0`.

**Note** - If you are migrating from Pachyderm <= 1.3 to 1.4+, you should read [this guide](https://github.com/pachyderm/pachyderm/tree/master/migration/1.3.x-1.4.x). In this particular case, a migration script is NOT provided due to significant changes in our processing and metadata structures. 

## Migrate to 1.5.x

To migrate from 1.4.x to 1.5.x, use the `pachctl migrate` command.  See `pachctl migrate --help` for detailed instructions.

As an example, to migrate from 1.4.8 to 1.5.0, use the following command:

```
$ pachctl migrate --from 1.4.8 --to 1.5.0
```

Note that the `pachctl migrate` command can be run either before or after you've redeployed your cluster with the new version (e.g. via `pachctl deploy`).

Most importantly, you need to ensure that your cluster is "at rest" when you run `pachctl migrate`.  That is, there shouldn't be any ongoing activities that are changing the state of the cluster.  Examples would be running jobs or ongoing `put-file` requests.

*Note: For v1.4 pipelines that specify environment variables in their pipeline specs, you will unfortunately need to reprocess all data for those pipelines as part of the v1.5 migration. This will automatically happen as part of the first job that spawns after the migration. Sorry for inconvenience.* 

## Backup

It’s paramount that you backup your data before running a migration.  While we’ve tested the migration code extensively, it’s still possible that they contain bugs, or that you accidentally use them in a wrong way.

In general, there are two data storage systems that you might consider backing up: the metadata storage and the data storage. Not all migration scripts touch both systems, so you might only need to back up one of them. Look at the README for a particular migration script for details.

### Backup the metadata store

Assuming you’ve deployed Pachyderm on a public cloud, your metadata is probably stored on a persistent volume. See the respective [Deploying Pachyderm](http://pachyderm.readthedocs.io/en/stable/deployment/deploy_intro.html) guide for details.

Here are official guides on backing up persistent volumes for each cloud provider:

- [GCE Persistent Volume](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Elastic Block Store (EBS)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

### Backup the object store 

We don’t currently have migration scripts that affect the object store.

