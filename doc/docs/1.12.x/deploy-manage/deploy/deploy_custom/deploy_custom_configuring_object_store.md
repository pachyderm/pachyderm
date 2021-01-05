# Configuring Object Store

Before reading this section, complete the steps in
[Configuring Persistent Disk Parameters](./deploy_custom_configuring_persistent_disk_parameters.md).

You can use the `--object-store` flag to configure Pachyderm to use an
s3 storage protocol to access the configured object store.
This configuration uses the Amazon S3 driver to access your on-premises
object store, regardless of the vendor,
since the Amazon S3 API is the standard with which every object store
is designed to work.

The S3 API has two different extant versions of *signature styles*,
which are how the object store validates client requests.
S3v4 is the most current version, but many S3v2 object-store servers
are still in use. Because support for S3v2 is scheduled
to be deprecated, Pachyderm recommends that you use S3v4 in all
deployments.

If you need to access an object store that uses S3v2
signatures, you can specify the `--isS3V2` flag.
This parameter configures Pachyderm to use the MinIO driver,
which allows the use of the older signature.
This `--isS3V2` flag disables SSL for connections to the object store
with the `minio` driver.
You can re-enable it with the `-s` or `--secure` flag.
You may also manually edit the `pachyderm-storage-secret` Kubernetes manifest.

The `--object-store` flag takes four required  configuration arguments.
Place these arguments immediately after
[the persistent disk parameters](deploy_custom_configuring_persistent_disk_parameters.md):

- `bucket-name`: The name of the bucket, without the `s3://`
prefix or a trailing forward slash (`/`).
- `access-key`: The user access ID that is used to access the
object store.
- `secret-key`: The associated password that is used with the user
access ID to access the object store.
- `endpoint`: The hostname and port that are used to access the object
store, in `<hostname>:<port>` format.

#### Example Invocation with a PV and Object Store

This example on-premises cluster uses an on-premises
MinIO object store with the following configuration parameters:

* An on-premises MinIO object store with the following parameters:
  - SSL is enabled
  - S3v4 signatures
  - The endpoint is `minio:9000`
  - The access key is `OBSIJRBE0PP2NO4QOA27`
  - The secret key is `tfteSlswRu7BJ86wekitnifILbZam1KYY3TG`
  - A bucket named `pachyderm-bucket`

The deployment command uses the following flags:

```
pachctl deploy custom --persistent-disk aws --object-store s3 \
    any-string 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 1
    [optional flags]
```
In the example command above, some of the arguments might
contain characters that the shell could interpret.
Those are enclosed in single-quotes.

!!! note
    Because the `deploy custom` command ignores the first
    configuration argument for the `--persistent-disk` flag,
    you can specify any string. For more information,
    see [Configuring Persistent Disk Parameters](./deploy_custom_configuring_persistent_disk_parameters.md)

After completing the steps described in this section, proceed to
[Create a Complete Configuration](./deploy_custom_complete_example_invocation.md).
