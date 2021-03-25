# Ingress and Egress Data from an External Object Store

Occasionally, you might need to download data from or upload data
to an object store that runs in a different cloud platform. For example,
you might be running a Pachyderm cluster in Microsoft Azure, but
you need to ingress files from an S3 bucket that resides on Amazon AWS.

You can configure Pachyderm to work with an external object
store by using the following methods:

* Ingress data from an external object store by using the
  `pachtl put file` with a URL to the S3 bucket. Example:

  ```
  $ pachctl put file repo@branch -f <s3://my_bucket/file>
  ```

* Egress data to an external object store by configuring the
  `egress` files in the pipeline specification. Example:

  ```shell
  # pipeline.json
  "egress": {
    "URL": "s3://bucket/dir"
  ```

## Configure Credentials

You can configure Pachyderm to ingress and egress from and to any
number of supported cloud object stores, including Amazon S3,
Microsoft Azure Blob storage, and Google Cloud Storage. You need
to provide Pachyderm with the credentials to communicate with
the selected cloud provider.

The credentials are stored in a
[Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/)
and share the same security properties.

!!! note
    For each cloud provider, parameters and configuration steps
    might vary.

To provide Pachyderm with the object store credentials, complete
the following steps:

1. Deploy object storage:

   ```shell
   $ pachctl deploy storage <storage-provider> ...
   ```

1. In the command above, specify `amazon`, `google`, or `microsoft` as
   a storage provider.

1. Depending on the storage provider, configure the required
   parameters. Run `pachctl deploy storage <backend> --help` for more
   information.

   For example, if you select `amazon`, you need to specify the following
   parameters:

   ```shell
   $ pachctl deploy storage amazon <region> <access-key-id> <secret-access-key> [<session-token>]
   ```

!!! note "See Also:"
    - [Custom Object Store](../deploy-manage/deploy/custom_object_stores.md)
    - [Create a Custom Pachyderm Deployment](../deploy-manage/deploy/deploy_custom/index.md)
    - [Pipeline Specification](../reference/pipeline_spec.md)
