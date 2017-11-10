# Ingressing From a Separate Object Store

Occasionally, you might find yourself needing to ingress data from or egress data (with the `put-file` command or `egress` field in the pipeline spec) to/from an object store that runs in a different cloud. For instance, you might be running a Pachyderm cluster in Azure, but you need to ingress files from a S3 bucket.

Fortunately, Pachyderm can be configured to ingress/egress from any number of supported cloud object stores, which currently include S3, Azure, and GCS.  In general, all you need to do is to provide Pachyderm with the credentials it needs to communicate with the cloud provider.

To provide Pachyderm with the credentials, you use the `pachctl deploy storage` command:

```bash
$ pachctl deploy storage <backend> ...
```

Here, `<backend>` can be one of `aws`, `google`, and `azure`, and the different backends take different parameters. Execute `pachctl deploy storage <backend>` to view detailed usage information.

For example, here's how you would deploy credentials for a S3 bucket:

```bash
$ pachctl deploy storage aws <bucket-name> <access key id> <secret access key>
```

Credentials are stored in a [Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/) and therefore share the same security properties.
