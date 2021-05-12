# S3 Gateway API

This section outlines the operations exposed by Pachyderm's HTTP API [S3 Gateway](../../deploy-manage/manage/s3gateway/). 


## `ListBuckets`

Route: `GET /`.

Lists all of the branches across all of the repos as S3 buckets.

## `DeleteBucket`

Route: `DELETE /<branch>.<repo>/`.

Deletes the branch. If it is the last branch in the repo, the repo is also
deleted. Unlike S3, you can delete non-empty branches.

## `ListObjects`

Route: `GET /<branch>.<repo>/`

Only S3's list objects v1 is supported.

PFS directories are represented via `CommonPrefixes`. This largely mirrors how
S3 is used in practice, but leads to a couple of differences:

* If you set the delimiter parameter, it must be `/`.
* Empty directories are included in listed results.

With regard to listed results:

* Due to PFS peculiarities, the `LastModified` field references when the most
recent commit to the branch happened, which may or may not have modified the
specific object listed.
* The HTTP `ETag` field does not use MD5, but is a cryptographically secure
hash of the file contents.
* The S3 `StorageClass` and `Owner` fields always have the same filler value.

## `GetBucketLocation`

Route: `GET /<branch>.<repo>/?location`

This will always serve the same location for every bucket, but the endpoint
is implemented to provide better compatibility with S3 clients.

## `GetBucketVersioning`

Route: `GET /<branch>.<repo>/?versioning`

This will get whether versioning is enabled, which is always true.

## `ListMultipartUploads`

Route: `GET /<branch>.<repo>/?uploads`

Lists the in-progress multipart uploads in the given branch. The `delimiter` query parameter is not supported.

## `CreateBucket`

Route: `PUT /<branch>.<repo>/`.

If the repo does not exist, it is created. If the branch does not exist, it
is likewise created. As per S3's behavior in some regions (but not all),
trying to create the same bucket twice will return a `BucketAlreadyOwnedByYou`
error.

## `DeleteObjects`

Route: `POST /<branch>.<repo>/?delete`.

Deletes multiple files specified in the request payload.

## `DeleteObject`

Route: `DELETE /<branch>.<repo>/<filepath>`.

Deletes the PFS file `filepath` in an atomic commit on the HEAD of `branch`.

## `GetObject`

Route: `GET /<branch>.<repo>/<filepath>`.

By default, this request gets the `HEAD` version of the file. You can use s3's
versioning API to get the object at a non-HEAD commit by specifying either a
specific commit ID, or by using the caret syntax -- for example, `HEAD^`.

There is support for range queries and conditional requests, however error
response bodies for bad requests using these headers are not standard S3 XML.

With regard to HTTP response headers:

* Due to PFS peculiarities, the HTTP `Last-Modified` header references when
the most recent commit to the branch happened, which may or may not have
modified this specific object.
* The HTTP `ETag` does not use MD5, but is a cryptographically secure hash of
the file contents.

## `PutObject`

Route: `PUT /<branch>.<repo>/<filepath>`.

Writes the PFS file at `filepath` in an atomic commit on the HEAD of `branch`.

Any existing file content is overwritten. Unlike S3, there is no limit to
upload size.

Unlike s3, a 64mb max size is not enforced on this endpoint. Especially,
as the file upload size gets larger, we recommend setting the `Content-MD5`
request header to ensure data integrity.

## `AbortMultipartUpload`

Route: `DELETE /<branch>.<repo>?uploadId=<uploadId>`

Aborts an in-progress multipart upload.

## `CompleteMultipartUpload`

Route: `POST /<branch>.<repo>?uploadId=<uploadId>`

Completes a multipart upload. If ETags are included in the request
payload, they must be of the same format as returned by the S3
gateway when the multipart chunks are included. If they are `md5`
hashes or any other hash algorithm, they are ignored.

## `CreateMultipartUpload`

Route: `POST /<branch>.<repo>?uploads`

Initiates a multipart upload.

## `ListParts`

Route: `GET /<branch>.<repo>?uploadId=<uploadId>`

Lists the parts of an in-progress multipart upload.

## `UploadPart`

Route: `PUT /<branch>.<repo>?uploadId=<uploadId>&partNumber=<partNumber>`

Uploads a chunk of a multipart upload.
