# Upgrades and Migrations

As new versions of Pachyderm are released, you may need to update your cluster to get access to bug fixes and new features. 
These updates fall into two categories, which are covered in detail at the links below:

* [Upgrades](./upgrades.html) - An upgrade is moving between point releases within the same major release (e.g. 1.7.2 --> 1.7.3).
  Upgrades are typically a simple process that require little to no downtime.

* [Migrations](./migrations.html) -- A migration what you must perform to move between major releases such as 1.8.7 --> 1.9.0.  

*Important*: Performing an _upgrade_ when going between _major releases_ may lead to corrupted data. 
*You must perform a [migration](./migrations.html) when going between major releases!*

Whether you're doing an upgrade or migration, it is recommended you [backup Pachyderm](./backups.html) prior. 
That will guarantee you can restore your cluster to its previous, good state. 
