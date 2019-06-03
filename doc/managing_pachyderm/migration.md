# Migration 

As new versions of Pachyderm are released, you may need to update your cluster to get access to bug fixes and new features. These updates fall into two categories, [Upgrades](./upgrades.md) and [Migrations](./migrations).

A Pachyderm Migration is upgrading between major releases such as 1.8.7 --> 1.9.0. Updating to the a new major release without following migration procedures can lead to corrupted data. Before migrating to a new major release you should first [upgrade](./upgrading.md) to the latest point release in the previous version (e.g. if you're on 1.8.5, you should first upgrade to 1.8.7 and then migrate to 1.9.0).

Whether doing an upgrade or migration, it is recommended to also follow our [backup procedures](backups.md). That way, if something goes wrong, you can restore your cluster to it's previous state. 

* [Before Migrating](#before-migrating), we recommend you follow standard backup procedures to ensure no data is lost in the case of an error. 
* [Migration Procedures](#migration) -- Also see specific instructions for each release
* Common Issues

## Before Migrating

Refer to [Backing up your cluster](./backups.md)...


# Migration

TODO: pull over specific info from exisiting doc and organize.

[Copy over from current migration doc](./backup_restore_and_migrate)

## 1.8.7 --> 1.9.0

TODO: FLESH OUT FURTHER

Only the etcd metadata needs to be migrated so the procedure should be much easier than a full data migration. Just to be safe, please follow standard backup procedures from your cloud provider to snapshot the bucket and persistent volume. 

1. `pachctl extract --no-objects` > migration
2. Undeploy, update pachctl, redeploy new version
3. pachctl restore

## 1.7.14 --> 1.8.0

**Important Information For v1.8**

In v1.8 we rearchitected core parts of the platform to [improve speed and scalability](http://www.pachyderm.io/2018/11/15/performance-improvements.html). The 1.7.x 1.8.x migration is a fairly indepth process (see our [Migrations docs](./backup_retore_and_migrate.md)) Therefore, it may actually be easier if you don't have tons of data to create a new 1.8 deployment and reingress data. Please come chat with us our [Public Slack Channel](slack.pachyderm.io) if you have any questions. 


1. `pachctl extract` > migration ( this will extract all data from the object store too so it may take a long time and be a large file)
2. Undeploy, update pachctl, redeploy new version
3. pachctl restore







