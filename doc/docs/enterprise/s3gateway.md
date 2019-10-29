# S3Gateway

Pachyderm Enterprise includes s3gateway, which allows you to interact with PFS
storage through an HTTP API that imitates a subset of AWS S3's API. With this,
you can reuse a number of tools and libraries built to work with object stores
such as [minio](https://docs.minio.io/docs/minio-client-quickstart-guide.html)
or [Boto](https://github.com/boto/boto3) to interact with Pachyderm.

You can only interact with the HEAD commit of non-authorization-gated PFS
branches through the gateway. If you need richer access, you'll need to work
with PFS through its gRPC interface instead.

## Connecting to the s3gateway

The s3gateway runs in your cluster. You can access it by pointing your browser
to `http://<cluster ip>:30600`.

Alternatively, you can use port forwarding to connect to the cluster.
However, we do not recommend it, as Kubernetes' port forwarder incurs overhead,
and does not recover well from broken connections.

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
* Object copying. PFS supports this through gRPC. You can use the gRPC API directly.
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

Attempting any of these operations returns a standard `NotImplemented`
error.

Additionally, there are a few general differences:

* There is no support for authentication or ACLs.
* As per PFS rules, you cannot write to an output repo. If you try to write
to an output repository, Pachyderm returns a 500 error.

## Minio

If you have the option of what S3 client library to use for interfacing with
the s3gateway, Pachyderm recommends [minio](https://min.io/), as integration with
its Golang client SDK is thoroughly tested.

Pachyderm supports the following Minio operations:

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
