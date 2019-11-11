# S3Gateway API

This outlines the HTTP API exposed by the s3gateway and any peculiarities
relative to S3. The operations largely mirror those documented in S3's
[official docs](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html).

Generally, you would not call these endpoints directly, but rather use a
tool or library designed to work with S3-like APIs. Because of that, some
working knowledge of S3 and HTTP is assumed.

### Operations on buckets

Buckets are represented via `branch.repo`, e.g. the `master.images` bucket
corresponds to the `master` branch of the `images` repo.

#### Creating buckets

Route: `PUT /branch.repo/`.

If the repo does not exist, it is created. If the branch does not exist, it
is likewise created. As per S3's behavior in some regions (but not all),
trying to create the same bucket twice will return a `BucketAlreadyOwnedByYou`
error.

#### Deleting buckets

Route: `DELETE /branch.repo/`.

Deletes the branch. If it is the last branch in the repo, the repo is also
deleted. Unlike S3, you can delete non-empty branches.

#### Listing buckets

Route: `GET /`.

Lists all of the branches across all of the repos as S3 buckets.

### Object operations

Object operations act upon the HEAD commit of branches. Authorization-gated
PFS branches are not supported.

#### Writing objects

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

#### Removing objects

Route: `DELETE /branch.repo/filepath`.

Deletes the PFS file `filepath` in an atomic commit on the HEAD of `branch`.

#### Listing objects

Route: `GET /branch.repo/`

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

#### Getting objects

Route: `GET /branch.repo/filepath`.

There is support for range queries and conditional requests, however error
response bodies for bad requests using these headers are not standard S3 XML.

With regard to HTTP response headers:

* Due to PFS peculiarities, the HTTP `Last-Modified` header references when
  the most recent commit to the branch happened, which may or may not have
  modified this specific object.
* The HTTP `ETag` does not use MD5, but is a cryptographically secure hash of
  the file contents.
