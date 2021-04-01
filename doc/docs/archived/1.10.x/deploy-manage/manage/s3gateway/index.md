# Overview

Pachyderm includes an S3 gateway that enables you to interact with PFS storage
through an HTTP application programming interface (API) that imitates the
Amazon S3 Storage API. Therefore, with Pachyderm S3 gateway, you can interact
with Pachyderm through tools and libraries designed to work with object stores.
For example, you can use these tools:

* [MinIO](https://docs.min.io/docs/minio-client-complete-guide)
* [boto3](https://boto3.amazonaws.com/v1/documentation/api/latest/index.html)

When you deploy `pachd`, the S3 gateway starts automatically.

The S3 gateway has some limitations that are outlined below. If you need richer
access, use the PFS gRPC interface instead, or one of the
[client drivers](https://github.com/pachyderm/python-pachyderm).

## Authentication

If auth is enabled on the Pachyderm cluster, credentials must be passed with
each S3 gateway endpoint using AWS' signature v2 or v4 methods. Object store
tools and libraries provide built-in support for these methods, but they do
not work in the browser. When you use authentication, set the access and
secret key to the same value. They are both the Pachyderm auth token used
to issue the relevant PFS calls.

If auth is not enabled on the Pachyderm cluster, access credentials are
ignored. You can not pass an authorization header, or if you do, any
values for the access and secret keys are ignored.

## Buckets

The S3 gateway presents each branch from every Pachyderm repository as
an S3 bucket.
For example, if you have a `master` branch in the `images` repository,
an S3 tool sees `images@master` as the `master.images` S3 bucket.

## Versioning

Most operations act on the `HEAD` of the given branch. However, if your object
store library or tool supports versioning, you can get objects in non-`HEAD`
commits by using the commit ID as the S3 object version ID.

## Port Forwarding

If you do not have direct access to the Kubernetes cluster, you can use port
forwarding instead. Simply run `pachctl port-forward`, which will allow you
to access the s3 gateway through `localhost:30600`.

However, the Kubernetes port forwarder incurs substantial overhead and
does not recover well from broken connections. Therefore, connecting to the
cluster directly is faster and more reliable.
