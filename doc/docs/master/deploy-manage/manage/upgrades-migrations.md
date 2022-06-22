# Upgrades and Migrations

!!! Info

    If you have questions about upgrades and migrations, you can post them in the community #help channel on [Slack](https://www.pachyderm.com/slack/){target=_blank}, or reach out to your TAM if you are an Enterprise customer.

As new versions of Pachyderm are released, you might need to update
your cluster to get access to bug fixes and new features.
These updates fall into the following categories:

* An [**upgrade**](../upgrades/) moves between **minor or patch releases**.
For example, between version 2.1.2 and 2.2.0. 
Upgrades are typically a simple process should require little to no downtime.

* Migrations — A **migration** must be performed when you are **moving between major releases**,
For questions on how to migrate, please contact your technical account manager or email support@pachyderm.com.

!!! Important 
    Performing an *upgrade* between *major releases* might lead to corrupted
    data. You must perform a migration when going between
    major versions.

Whether you upgrade or migrate your cluster, Pachyderm recommends that you
[perform a back up](../backup-restore/). A backup guarantees that you can restore
your cluster to its previous, stable state.
