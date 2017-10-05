# Deploy storage credentials for a different cloud

Occasionally, you might find yourself needing to ingress data from or egress data to a storage solution that runs in a different cloud.  For instance, you might be running an Azure cluster, but you need to ingress files from a S3 bucket.

Fortunately, Pachyderm can be configured to use any number of supported cloud storage solutions, which currently include S3, Azure, and GCS (google cloud storage).  In general, all you need to do is to provide Pachyderm with the credentials it needs to communicate with the cloud provider.

To provide Pachyderm with the credentials, you use the `pachctl deploy storage` command:

```bash
$ pachctl deploy storage <backend> ...
```

Here, `<backend>` can be one of `aws`, `google`, and `azure`.  Different backends take different parameters.  When in doubt, simply type in `pachctl deploy storage <backend>` to view detailed usage information.

For instance, here's how you would deploy credentials for a S3 bucket:

```bash
$ pachctl deploy storage aws <bucket-name> <access key id> <secret access key>
```

Credentials are stored in a [Kubernetes secret](https://kubernetes.io/docs/concepts/configuration/secret/) and therefore share the same security properties with it.
