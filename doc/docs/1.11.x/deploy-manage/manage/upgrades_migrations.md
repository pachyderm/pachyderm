# Upgrades and Migrations

As new versions of Pachyderm are released, you might need to update
your cluster to get access to bug fixes and new features.
These updates fall into the following categories:

* [Upgrades](./upgrades.md) — An **upgrade** moves between **minor or point releases**.
For example, between version 1.12.2 and 1.13.0. 
Upgrades are typically a simple process that requires little to no downtime.

* [Migrations](./migrations.md) — A **migration** must be performed when you are **moving between major releases**,
such as moving from 1.13.x to 2.0.

!!! Important
    Performing an *upgrade* between *major releases* might lead to corrupted
    data. You must perform a [migration](./migrations.md) when going between
    major versions!

Whether you upgrade or migrate your cluster, Pachyderm recommends that you
[perform a back up](./backup_restore.md). A backup guarantees that you can restore
your cluster to its previous, stable state.

