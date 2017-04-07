# Migrations

Occationally, Pachyderm introduces changes that are backward-incompatible: repos/commits/files created on an old version of Pachyderm may be unusable on a new version of Pachyderm. When that happens, we try our best to write a migration script that "upgrades" your data so it’s usable by the new version of Pachyderm.

To upgrade from version X to version Y, look under the directory named `migration/X-Y`. For instance, to upgrade from 1.3.12 to 1.4.0, look under `migration/1.3.12-1.4.0`.

**Note** - If you are migrating from Pachyderm <= 1.3 to 1.4+, you should read [this guide](https://github.com/pachyderm/pachyderm/tree/master/migration/1.3.x-1.4.x). In this particular case, a migration script is NOT provided due to significant changes in our processing and metadata structures. 

## Backup

It’s paramount that you backup your data before running a migration script. While we’ve tested the scripts extensively, it’s still possible that they contain bugs, or that you accidentally use them in a wrong way.

In general, there are two data storage systems that you might consider backing up: the metadata storage and the data storage. Not all migration scripts touch both systems, so you might only need to back up one of them. Look at the README for a particular migration script for details.

### Backup the metadata store

Assuming you’ve deployed Pachyderm on a public cloud, your metadata is probably stored on a persistent volume. See the respective [Deploying Pachyderm](http://pachyderm.readthedocs.io/en/stable/deployment/deploy_intro.html) guide for details.

Here are official guides on backing up persistent volumes for each cloud provider:

- [GCE Persistent Volume](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Elastic Block Store (EBS)](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

### Backup the object store 

We don’t currently have migration scripts that touch the data storage system.

