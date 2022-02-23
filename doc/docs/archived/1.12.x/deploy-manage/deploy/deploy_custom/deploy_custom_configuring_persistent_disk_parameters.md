# Configuring Persistent Disk Parameters

Before reading this section, complete the steps in [Before You
Begin](./deploy_custom_before_you_begin.md).

To create a custom deployment, you need to configure
persistent storage that Pachyderm will use to store metadata.
You can do so by using the `--persistent-disk` flag
that creates a PersistentVolume (PV) backend on a supported provider.

Pachyderm has automated configuration for styles of backend for the
following major cloud providers:

- Amazon Web Services™ (AWS)
- Google Cloud Platform™ (GCP)
- Microsoft® Azure™

Choosing one of these providers creates a configuration close to what
you need.
After carefully reading the section below,
consult with your Kubernetes administrators on which provider to choose.
You might need to then edit your manifest manually,
based on configuration information they provide to you.

For each of the providers above,
the final configuration depends on which of
the following flags you define:

* `--dynamic-etcd-nodes`. The `--dynamic-etcd-nodes` flag is
used when your Kubernetes installation is configured to use
[StatefulSets](../../on_premises/#statefulsets).
Many Kubernetes deployments use StatefulSets as a reliable solution that
ensures the persistence of pod storage. Your on-premises
Kubernetes installation might also be configured to use StatefulSets.
The `--dynamic-etcd-nodes` flag specifies the number of `etcd` nodes
that your deployment creates. Pachyderm recommends that you keep this
number at `1`. If you want to change it, consult with your Pachyderm
support team.
This flag creates a `VolumeClaimTemplate` in the `etcd` `StatefulSet`
that uses the standard `etcd-storage-class`.

!!! note
    Consult with your Kubernetes administrator about the StorageClass
    that you should use for `etcd` in your Kubernetes deployment.
    If you need to use a different than the default setting,
    you can use the `--etcd-storage-class` flag to specify the StorageClass.

* `--static-etcd-volume`. The `--static-etcd-volume` flag is used when
your Kubernetes installation has not been configured to use StatefulSets.
When you specify `--static-etcd-volume` flag, Pachyderm creates a static
volume for `etcd`. Pachyderm creates a PV with a spec appropriate
for each of the cloud providers:

- `aws`: awsElasticBlockStore for Amazon Web Services
- `google`: gcePersistentDisk for Google Cloud Storage
- `azure`: azureDisk for Microsoft Azure

As stated above, the specifics of one of these choices might
not match precisely what your on-premises deployment requires.
To determine the closest correct choices for your on-prem infrastructure,
consult with your Kubernetes administrators.
You might have to then edit your manifest manually,
based on configuration information they provide to you.

## Example invocation with persistent disk parameters

This example on-premises cluster has StatefulSets
enabled, with the standard etcd storage class configured.
The deployment command uses the following flags:

```
pachctl deploy custom --persistent-disk aws --object-store <object store backend> \
    any-string 10 \
    <object store arg 1> <object store arg 2>  <object store arg 3>  <object store arg 4> \
    --dynamic-etcd-nodes 1
    [optional flags]
```

The `--persistent-disk` flag takes two arguments
that you specify right after the single argument to the `--object-store` flag.
Although the first argument is required, Pachyderm ignores it.
Therefore, you can set it to any text value, such as `any-string` in the
example above.
The second argument is the size,
in gigabytes (GB), that Pachyderm requests for the `etcd` disk.
A good value for most deployments is 10.

After completing the steps described in this section, proceed to
[Configuring Object Store](./deploy_custom_configuring_object_store.md).
