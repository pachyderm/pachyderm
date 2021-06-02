# Ingress and Egress Data from an External Object Store

You might need to download data from or upload data
to an object store that runs in a different cloud platform. For example,
you might be running a Pachyderm cluster in Microsoft Azure, but
you need to ingress files from an S3 bucket that resides on Amazon AWS.

You can configure Pachyderm to work with an external object
store by providing Pachyderm with the credentials to communicate with
the selected cloud provider. 

This is required when running a ` pachctl put file repo@branch -f <s3://my_bucket/file>` or setting up an egress in your pipeline specifications `"egress": {"URL": "s3://bucket/dir"}` for example.

!!! note
    For each cloud provider, parameters and configuration steps
    might vary.

To provide Pachyderm with the object store credentials:

1. Run the following command:

    ```shell
    pachctl deploy storage <storage-provider> ...
    ```
    Where storage-provider can be: `amazon`, `google`, or `microsoft`.

1. Depending on the storage provider, configure the required
   parameters. Run `pachctl deploy storage <storage-provider> --help` for more
   information.

    For example, if you select `amazon`, you need to specify the following
    parameters:

    ```shell
    pachctl deploy storage amazon <region> <access-key-id> <secret-access-key> [<session-token>]
    ```

      Those credentials are ultimately stored in a
      [Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/).

!!! note "See Also:"
    - [Custom Object Store](../../deploy-manage/deploy/custom_object_stores.md)
    - [Create a Custom Pachyderm Deployment](../../deploy-manage/deploy/deploy_custom/index.md)

