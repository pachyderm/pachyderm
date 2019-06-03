# Upgrades and Migrations


As new versions of Pachyderm are released, you may need to update your cluster to get access to bug fixes and new features. These updates fall into two categories, [Upgrades](./upgrades.md) and [Migrations](./migrations).

* [Upgrades](./upgrades.md) - An Pachyderm deployment upgrade is moving between point releases within the same major release (e.g. 1.7.2 --> 1.7.3). Upgrades are typically a simple process that require little to no downtime.

* [Migrations](./migrations) -- A Pachyderm Migration is upgrading between major releases such as 1.8.7 --> 1.9.0. Updating to the a new major release without following migration procedures can lead to severely corrupted data. 

Whether doing an upgrade or migration, it is recommended to also follow our [backup procedures](backups.md). That way, if something goes wrong, you can restore your cluster to it's previous state. 