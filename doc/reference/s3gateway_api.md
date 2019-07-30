# S3Gateway API

This outlines the HTTP API exposed by the s3gateway and any peculiarities
relative to S3. The operations largely mirror those documented in S3's
[official docs](https://docs.aws.amazon.com/AmazonS3/latest/API/Welcome.html).

Generally, you would not call these endpoints directly, but rather use a
tool or library designed to work with S3-like APIs. Because of that, some
working knowledge of S3 and HTTP is assumed.

### Operations on the Service

#### GET Service

Route: `GET /`.

Lists all of the branches across all of the repos as S3 buckets.

### Operations on Buckets

#### DELETE Bucket

Route: `DELETE /<repo>/`.

Deletes the repo.

#### GET Bucket (List Objects) Version 1

Route: `GET /<repo>/?branch=<branch>`

Only S3's list objects v1 is supported. The `branch` parameter is a
non-standard extension that allows you to specify which branch to on the given
repo to list objects from; if unspecified, it defaults to `master`.

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

#### GET Bucket location

Route: `GET /<repo>/?location`

This will always serve the same location for every bucket, but the endpoint is
implemented to provide better compatibility with S3 clients.

#### List Multipart Uploads

Route: `GET /<branch>.<repo>/?uploads&branch=<branch>`

Lists the in-progress multipart uploads in the given branch. The `delimiter`
query parameter is not supported. The `branch` parameter is a non-standard
extension that allows you to specify which branch on the given repo to
list objects from; if unspecified, it defaults to `master`.

#### PUT Bucket

Route: `PUT /<repo>/`.

If the repo does not exist, it is created. As per S3's behavior in some
regions (but not all), trying to create the same bucket twice will return a
`BucketAlreadyOwnedByYou` error.

### Operations on Objects

#### DELETE Object

Route: `DELETE /<repo>/<filepath>`.

Deletes the PFS file `filepath` in an atomic commit. Note that this can only
be done on the HEAD of a branch.

#### GET Object

Route: `GET /<repo>/<filepath>`.

There is support for range queries and conditional requests, however error
response bodies for bad requests using these headers are not standard S3 XML.

With regard to HTTP response headers:
* Due to PFS peculiarities, the HTTP `Last-Modified` header references when
the most recent commit to the branch happened, which may or may not have
modified this specific object.
* The HTTP `ETag` does not use MD5, but is a cryptographically secure hash of
the file contents.

#### PUT Object

Route: `PUT /<repo>/<filepath>`.

Writes the PFS file at `filepath` in an atomic commit. Note that this can only
be done on the HEAD of a branch.

#### Abort Multipart Upload

Route: `DELETE /<repo>?uploadId=<uploadId>&branch=<branch>`

Aborts an in-progress multipart upload. The `branch` parameter is a
non-standard extension that allows you to specify which branch to on the given
repo to list objects from; if unspecified, it defaults to `master`.

#### Complete Multipart Upload

Route: `POST /<repo>?uploadId=<uploadId>&branch=<branch>`

Completes a multipart upload. Note that if ETags are included in the request
payload, they must be of the same format as returned by s3gateway when the
multipart chunks are included. If they are md5 hashes (or any other hash
algorithm), they will be ignored. The `branch` parameter is a non-standard
extension that allows you to specify which branch on the given repo to list
objects from; if unspecified, it defaults to `master`.

#### Initiate Multipart Upload

Route: `POST /<repo>?uploads&branch=<branch>`

Initiates a multipart upload. The `branch` parameter is a non-standard
extension that allows you to specify which branch on the given repo to list
objects from; if unspecified, it defaults to `master`.

#### List Parts

Route: `GET /<repo>?uploadId=<uploadId>&branch=<branch>`

Lists the parts of an in-progress multipart upload. The `branch` parameter is
a non-standard extension that allows you to specify which branch on the
given repo to list objects from; if unspecified, it defaults to `master`.

#### Upload Part

Route: `PUT /<repo>?uploadId=<uploadId>&partNumber=<partNumber>&branch=<branch>`

Uploads a chunk of a multipart upload. The `branch` parameter is a
non-standard extension that allows you to specify which branch on the given
repo to list objects from; if unspecified, it defaults to `master`.
