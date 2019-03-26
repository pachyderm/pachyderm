# S3Gateway

Pachyderm includes s3gateway, which allows you to interact with PFS storage
through an HTTP API that imitates a subset of AWS S3's API. With this, you
can reuse a number of tools and libraries built to work with object stores
(e.g. [minio](https://docs.minio.io/docs/minio-client-quickstart-guide.html),
[Boto](https://github.com/boto/boto3)) to interact with pachyderm.

## Connecting to the s3gateway

You can start the s3gateway locally by running `pachctl s3gateway`. Then point
your browser or favorite s3 tool at `http://localhost:30600`. The s3gateway
also runs in the background on your pachyderm cluster, which you can reach via
`http://<cluster ip>:30600`.

**Note:** A third way to access the s3gateway is using the explicit port
forwarder (`pachctl port-forward`). However, we don't recommend it, as
kubernetes' port forwarder incurs overhead, and does not recover well from
broken connections.

## Overview

The s3gateway gives you access to the HEAD commit of non-authorization-gated
PFS branches and repos.

Buckets are represented via `branch.repo`, e.g. the `master.images` bucket
corresponds to the `master` branch of the `images` repo.

## Bucket operations

### Creating buckets

Route: `PUT /branch.repo/`.

If the repo does not exist, it is created. If the branch does not exist, it
is likewise created. As per S3's behavior in some regions (but not all),
trying to create the same bucket twice will return a `BucketAlreadyOwnedByYou`
error.

### Deleting buckets

Route: `DELETE /branch.repo/`.

Unlike S3, you can delete non-empty branches.

### Listing buckets

Route: `GET /`.

### Bucket stats

Route: `HEAD /branch.repo/`

## Object operations

### Writing objects

Route: `PUT /branch.repo/filepath`.

Writes the PFS file at `filepath` in an atomic commit on the HEAD of `branch`.
Any existing file content is overwritten. Unlike S3, there is no limit to
upload size.

The s3gateway does not support multipart uploads, but you can use this
endpoint to upload very large files. We recommend setting the `Content-MD5`
request header - especially for larger files - to ensure data integrity.

Some S3 libraries and clients will detect that our s3gateway does not support
multipart uploads and automatically fallback to using this endpoint. Notably,
this includes minio.

### Removing objects

Route: `DELETE /branch.repo/filepath`.

Deletes the PFS file `filepath` in an atomic commit on the HEAD of `branch`.

### Listing objects

Route: `GET /branch.repo/`

Only v1 is supported. PFS directories are represented via `CommonPrefixes`.
This largely mirrors how S3 is used in practice, but leads to a couple of
differences:
* If you set a delimiter, it must be `/`.
* Empty directories are included in listed results.

With regard to listed results:
* Due to PFS peculiarities, the `LastModified` field references when the most
recent commit to the branch happened, which may or may not have modified the
specific object listed.
* The `ETag` field does not use md5, but should still be treated as a
cryptographically secure hash of the file contents.
* The `StorageClass` and `Owner` fields always have the same default value.

### Getting objects

Route: `GET /branch.repo/filepath`.

There is support for range queries and conditional requests, however error
response bodies for bad requests using these headers are not standard S3 XML.

With regard to HTTP response headers:
* Due to PFS peculiarities, `Last-Modified` references when the most recent
commit to the branch happened, which may or may not have modified this
specific object.
* `ETag` does not use md5, but should still be treated as a cryptographically
secure hash of the file contents.

### Object stats

Route: `HEAD /branch.repo/filepath`.

With regard to HTTP response headers:
* Due to PFS peculiarities, `Last-Modified` references when the most recent
commit to the branch happened, which may or may not have modified this
specific object.
* `ETag` does not use md5, but should still be treated as a cryptographically
secure hash of the file contents.

## Unsupported operations

These S3 operations are not supported by the gateway:

* Accelerate.
* Analytics.
* Object copying. PFS does support this through gRPC though, so if you need
this, you can use the gRPC API directly.
* CORS configuration.
* Encryption.
* HTML form uploads.
* Inventory.
* Legal holds.
* Lifecycles.
* Logging.
* Metrics.
* Multipart uploads. See writing object documentation above for a workaround.
* Notifications.
* Object locks.
* Payment requests.
* Policies.
* Public access blocks.
* Regions.
* Replication.
* Retention policies.
* Tagging.
* Torrents.
* Website configuration.

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
