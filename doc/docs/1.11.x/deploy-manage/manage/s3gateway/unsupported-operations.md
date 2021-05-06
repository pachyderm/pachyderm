## Unsupported operations

Some of the S3 operations are not yet supported by Pachyderm.
If you run any of these operations, Pachyderm returns a standard
S3 `NotImplemented` error.

The S3 Gateway does not support the following S3 operations:

* Accelerate
* Analytics
* Object copying. PFS supports this functionality through gRPC.
* CORS configuration
* Encryption
* HTML form uploads
* Inventory
* Legal holds
* Lifecycles
* Logging
* Metrics
* Notifications
* Object locks
* Payment requests
* Policies
* Public access blocks
* Regions
* Replication
* Retention policies
* Tagging
* Torrents
* Website configuration
