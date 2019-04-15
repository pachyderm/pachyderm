# S3Gateway

Pachyderm Enterprise includes s3gateway, which allows you to interact with PFS
storage through an HTTP API that imitates a subset of AWS S3's API. With this,
you can reuse a number of tools and libraries built to work with object stores
(e.g. [minio](https://docs.minio.io/docs/minio-client-quickstart-guide.html),
[Boto](https://github.com/boto/boto3)) to interact with pachyderm.

You can only interact with the HEAD commit of non-authorization-gated PFS
branches through the gateway. If you need richer access, you'll need to work
with PFS through its gRPC interface instead.

## Connecting to the s3gateway

You can start the s3gateway locally by running `pachctl s3gateway`. Then point
your browser or favorite s3 tool at `http://localhost:30600`. The s3gateway
also runs in the background on your pachyderm cluster, which you can reach via
`http://<cluster ip>:30600`.

**Note:** A third way to access the s3gateway is using the explicit port
forwarder (`pachctl port-forward`). However, we don't recommend it, as
kubernetes' port forwarder incurs overhead, and does not recover well from
broken connections.

## Supported operations

These operations are supported by the gateway:

* Creating buckets: creates a repo and/or branch.
* Deleting buckets: Deletings a branch and/or repo.
* Listing buckets: Lists all branches on all repos as s3 buckets.
* Writing objects: Atomically overwrites a file on the HEAD of a branch.
* Removing objects: Atomically removes a file on the HEAD of a branch.
* Listing objects: Lists the files in the HEAD of a branch.
* Getting objects: Gets file contents on the HEAD of a branch.

For details on what's going on under the hood and current peculiarities, see the
[s3gateway API](../reference/s3gateway_api.md).

## Unsupported operations

These S3 operations are not supported by the gateway:

* Accelerate
* Analytics
* Object copying. PFS does support this through gRPC though, so if you need
this, you can use the gRPC API directly.
* CORS configuration
* Encryption
* HTML form uploads
* Inventory
* Legal holds
* Lifecycles
* Logging
* Metrics
* Multipart uploads. See writing object documentation above for a workaround.
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

Attempting any of these operations will return a standard `NotImplemented`
error.

Additionally, there are a few general differences:

* There's no support for authentication or ACLs.
* As per PFS rules, you can't write to an output repo. At the moment, a 500
error will be returned if you try to do so.

## Minio

If you have the option of what S3 client library to use for interfacing with
the s3gateway, we recommend [minio](https://min.io/), as integration with it's
go client SDK is thoroughly tested. These minio operations are supported:

Bucket operations
* `MakeBucket`
* `ListBuckets`
* `BucketExists`
* `RemoveBucket`
* `ListObjects`

Object operations
* `GetObject`
* `PutObject`
* `StatObject`
* `RemoveObject`
* `FPutObject`
* `FGetObject`
