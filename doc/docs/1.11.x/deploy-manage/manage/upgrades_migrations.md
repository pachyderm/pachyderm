# Upgrades and Migrations

As new versions of Pachyderm are released, you might need to update
your cluster to get access to bug fixes and new features.
These updates fall into the following categories:

* [Upgrades](./upgrades.md) — An upgrade is moving between point releases
  within the same major/minor release. For example, between version 1.12.0 and 1.12.2.
  Upgrades are typically a simple process that require little to no downtime.

* [Migrations](./migrations.md) — A migration that you must perform to move
  between major or minor releases, such as between version 1.11.7 and {{ config.pach_latest_version }}.

!!! Important
    Performing an *upgrade* between *major or minor releases* might lead to corrupted
    data. You must perform a [migration](./migrations.md) when going between
    major/minor releases!

Whether you upgrade or migrate your cluster, Pachyderm recommendeds that you
[back up Pachyderm](./backup_restore.md). A backup guarantees that you can restore
your cluster to its previous, stable state.
