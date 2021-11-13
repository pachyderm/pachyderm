# Ingress and Egress Data from an External Object Store

You can configure Pachyderm to work with an external object
store by providing Pachyderm with its credentials. 

Running a ` pachctl put file repo@branch -f <s3://my_bucket/file>` or egressing your data `"egress": {"URL": "s3://bucket/dir"}` in an s3 bucket, for example, requires Pachyderm to be able to access it.

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

