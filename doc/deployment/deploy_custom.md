# Create a Custom Pachyderm Deployment

Pachyderm includes the `pachctl deploy custom` command for creating
customized deployments for cloud providers or on-premises use.

This section describes how to use `pachctl deploy custom` to create
a manifest for a custom, on-premises deployment.
Although deployment automation is out of scope of this section,
Pachyderm strongly encourages you to treat your
[infrastructure as code](./on-premises.html#infrastructure-as-code).

This document describes that customization in two primary parts:

* Invoking `pachctl deploy custom` to create a custom manifest.
* Examining that manifest in detail.

This section also goes through some examples and
provides the next steps for editing and deployment.

This section includes the following topics:

.. toctree::
   :maxdepth: 1

   deploy_custom_pachyderm_deployment_manifest.md
   deploy_manifest_example_custom_deployment.md
   deploy_custom_configuring_persistent_disk_parameters.md
   deploy_custom_configuring_object_store.md
   deploy_custom_complete_example_invocation.md
   deploy_custom_additional_flags.md

## Before You Begin

Before you start creating a custom deployment, verify that you have
completed the following steps:

1. Read and complete the steps described in the [Introduction](./on_premises.html#introduction)
in the *On Premises* section. This section explains differences between
static persistent volumes, StatefulSets, and StatefulSets with StorageClasses.
Also, it explains the meanings of the variables, such as  `PVC_STORAGE_SIZE`
and `OS_ENDPOINT` that are used in the examples below.
1. Install [kubectl](https://kubernetes.io/docs/user-guide/prereqs/).
1. Install [pachctl](http://docs.pachyderm.io/en/latest/pachctl/pachctl.html).

## Pachyderm Deployment Manifest

When you run the `pachctl deploy` command, Pachyderm generates a JSON-encoded
Kubernetes manifest named `manifest.custom` which consists of sections that
describe a Pachyderm deployment.

Pachyderm deploys the following sets of application components:

- `pachd`: The main Pachyderm pod.
- `etcd`: The administrative datastore for `pachd`.
- `dash`: The web-based UI for Pachyderm Enterprise Edition.

**Example:**

```
pachctl deploy custom --persistent-disk <persistent disk backend> --object-store <object store backend> \
    <persistent disk arg1>  <persistent disk arg 2> \
    <object store arg 1> <object store arg 2>  <object store arg 3>  <object store arg 4> \
    [[--dynamic-etcd-nodes n] | [--static-etcd-volume <volume name>]]
    [optional flags]
```

As you can see in the example command above, you can run the
`pachctl deploy custom` command with different flags that generate
an appropriate manifest for your infrastructure. The flags broadly fall
into the following categories:

| Category               | Description                                         |
| ---------------------- | --------------------------------------------------- |
| `--persistent-disk`    | Configures the storage resource for etcd. Pachyderm uses etcd <br> to manage administrative metadata. User data is stored in an <br> object store, not in etcd. |
| `--object-store`       | Configures the object store that Pachyderm uses for storing all <br> user data that you want to be versioned and managed. |
| Optional flags         | Optional flags, that are not required to deploy Pachyderm but <br> enable you to configure access, output format, logging verbosity, and other parameters. |

### Configuring Persistent Disk Parameters

To create a custom deployment,
you need to configure persistent storage
that Pachyderm uses to store metadata.
You can do so by using the `--persistent-disk` flag
that creates a PV backend on a supported provider.

Pachyderm has automated configuration for styles of backend for the
following major cloud providers:

- Amazon Web Services™ (AWS)
- Google Cloud Platform™ (GCP)
- Microsoft® Azure™

For each of those providers,
the final configuration highly depends on which of
of the following flags you define:

* `--dynamic-etcd-nodes`. The `--dynamic-etcd-nodes` flag is
used when your Kubernetes installation is configured to use
[StatefulSets](./on_premises.html#statefulsets).
Many Kubernetes deployments use StatefulSets as a reliable solution that
ensures persistence of pod storage. It is likely that your on-premises
Kubernetes installation is configured to use StatefulSets.
The `--dynamic-etcd-nodes` flag specifies the number of `etcd` nodes
that your deployment creates. Pachyderm recommends that you keep this
number at `1`. If you want to change it, consult with your Pachyderm
support team.
This flag creates a `VolumeClaimTemplate` in the `etcd` `StatefulSet`
that uses the standard `etcd-storage-class`.

**NOTE** Consult with your Kubernetes administrator about the StorageClass 
that you should use for `etcd` in your Kubernetes deployment.
If you need to use a different than the default setting,
you can use the `--etcd-storage-class` flag to specify the StorageClass.

* `--static-etc-volume`. The `--static-etc-volume` is used when
your Kubernetes installation has not been configured to use StatefulSets.
When you specify `--static-etc-volume` flag, Pachyderm creates a static
volume for `etcd`. Pachyderm creates a PV with a spec appropriate
for each of the cloud providers:

- `aws`: awsElasticBlockStore for Amazon Web Services
- `google`: gcePersistentDisk for Google Cloud Storage
- `azure`: azureDisk for Microsoft Azure

These choices are not applicable to  most on-premises deployments.
To determine the correct choices for your on-prem infrastructure,
consult with your Kubernetes administrators
and edit your manifest manually.

Regardless whether you choose to deploy with StatefulSets or static volumes,
the `--persistent-disk` flag takes two arguments
that you specify right after the single argument to the `--object-store` flag.
Although the first argument is required, Pachyderm ignores it.
Therefore, you can set it to any text value.
The second argument is the size,
in gigabytes (GB), that Pachyderm requests for the `etcd` disk.
A good value for most deployments is 10.

#### Example invocation with persistent disk parameters

This example on-premises cluster has StatefulSets
enabled, with the standard etcd storage class configured.
The deployment command uses the following flags:

```
pachctl deploy custom --persistent-disk aws --object-store <object store backend> \
    foobar 10 \
    <object store arg 1> <object store arg 2>  <object store arg 3>  <object store arg 4> \
    --dynamic-etcd-nodes 10
    [optional flags]
```

For more information, see [storage manifests](#persistentvolume).

### Configuring Object Store

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
Also, this flag disables SSL for connections to the object store
with the `minio` driver.
You can re-enable it with the `-s` or `--secure` flag.

Also, you can edit the `pachyderm-storage-secret` Kubernetes manifest.
The `--object-store` flag takes four required, additional
configuration arguments.
These arguments must be placed immediately after
[the persistent disk configuration parameters](#persistent-disk-parameters):

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
    foobar 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 10
    [optional flags]
```

In the example command above, some of the arguments may
contain characters that the shell could interpret in single-quotes.

### Complete example invocation

Because you are deploying to a code infrastructure that works
with YAML files, you need to add the flags for that.
Limit your configuration to local roles only by adding the
`--local-roles` flag.
Because you need to deploy your cluster to a code infrastructure
that works with YAML files, the output will be saved to a file
that the you can use with your code infrastructure for automated
deployment and management.
The final deployment command includes the following flags:

```
pachctl deploy custom --persistent-disk aws --object-store s3 \
    foobar 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 10
    --local-roles --output yaml  --dry-run > custom_deploy.yaml
```

For more information about the contents of the `custom_deploy.yaml` file,
see [Anatomy of a Pachyderm deployment manifest](#anatomy-of-a-pachyderm-deployment-manifest).

## Roles and permissions manifests

This section walks you through the roles and permissions manifests.

### ServiceAccount

Typically at the top of the file, a roles and permissions manifest has the `kind` key set to `ServiceAccount`. 
[ServiceAccounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) are a way Kubernetes can assign namespace-specific privileges to applications in a lightweight way.
The Pachyderm's service account is called  `pachyderm`.

### Role or ClusterRole

Depending on whether you used the `--local-roles` flag or not, the next manifest `kind` is `Role` or `ClusterRole`.

### RoleBinding or ClusterRoleBinding

This manifest binds the `Role` or `ClusterRole` to the `ServiceAccount` created above.

## Application-related

This section walks you through the application-related manifests.

### PersistentVolume

If you used `--static-etc-volume` to deploy Pachyderm,
the value that you specify for `--persistent-disk` causes `pachctl` to write a manifest for creating a [`PersistentVolume`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) that Pachyderm's `etcd` uses in its [`PersistentVolumeClaim`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims).

A common persistent volume that is used in enterprises is an NFS mount backed by a storage fabric.
In this case, a StorageClass for an NFS mount is made available for consumption.
Consult with your Kubernetes administrators to learn what resources are available for your deployment.

### PersistentVolumeClaim

If you deployed Pachyderm by using `--static-etc-volume`, the Pachyderm's `etcd` uses this `PersistentVolumeClaim`.
See this manifest's name in the `etcd` Deployment manifest below.

### StorageClass

If you used the `--dynamic-etcd-nodes` flag to deploy Pachyderm,
this manifest specifies the kind of storage and provisioner that is appropriate for what you have specified in the `--persistent-disk` flag. 

**Note:** You will not see this manifest if you specified `azure` as the argument to `--persistent-disk`.

### Service

In a typical Pachyderm deployment, 
you see three [`Service`](https://kubernetes.io/docs/concepts/services-networking/service/) manifests. 
Services are how Kubernetes exposes Pods to the network.
If you use StatefulSets to deploy Pachyderm,
that is, you use `--dynamic-etcd-nodes` flag,
Pachyderm deploys one `Service` for `etcd-headless`, one for `pachd`, and one for `dash`.
A static deployment has `Services` for `etcd`, `pachd`, and `dash`.

If you use the `--no-dashboard` flag, Pachyderm does not create the `dash` `Service` and `Deployment`.
Likewise, if `--dashboard-only` is specified,
Pachyderm generates the manifests for the Pachyderm enterprise UI only. 

The most common items that you can edit in `Service` manifests are the `NodePort` values for various services, 
and the `containerPort` values for `Deployment` manifests.
To make your `containerPort` values work properly, add environment variables to a `Deployment` or `StatefulSet` object.
You can verify this functionality in the [OpenShift](./openshift.html) example.

### The Pachyderm pods

#### Deployment 

A [`Deployment`](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) declares the desired state of application pods to Kubernetes.

If you configure a static deployment,
Pahyderm deploys `Deployment` manifests for `etcd`, `pachd`, and `dash`.
If you specify `--dynamic-etd-nodes`, Pachyderm deploys the `pachd` and `dash` as `Deployment`
and `etcd` as a`StatefulSet`.

If you run the deploy command with the `--no-dashboard` flag, Pachyderm omits the deployment of the `dash` `Service` and `Deployment`.


#### StatefulSet

For a `--dynamic-etcd-nodes` deployment, Pachyderm replaces the `etcd` `Deployment` manifest with a `StatefulSet`.

### Secret

The final manifest is a Kubernetes [`Secret`](https://kubernetes.io/docs/concepts/configuration/secret/).
Pachyderm uses the `Secret` to store the credentials that are necessary to access object storage.
The final manifest uses the command-line arguments that you submit to the `pachctl deploy` command to store such parameters
as region, secret, token, and endpoint,
that are used to access an object store. 
The exact values in the secret depend on the kind of object store you configure for your deployment.
You can update the values after the deployment either by using `kubectl` to deploy a new `Secret`
or the `pachctl deploy storage` command.


## More examples

### Configuring with a static persistent volume
The command you'll want to run is 
```sh
$ pachctl deploy custom --persistent-disk aws --object-store s3 
         ${PVC STORAGE_NAME} ${PVC STORAGE_SIZE} ${OS_BUCKET_NAME} ${OS_ACCESS_KEY_ID} ${OS_SECRET_KEY} ${OS_ENDPOINT} \
         --static-etcd-volume=${PVC_STORAGE_NAME}  \
         --dry-run > pachyderm-with-static-volume.json
```
### Configuring with StatefulSets
The command you'll want to run is 
```sh
$ pachctl deploy custom --object-store s3 any-string 
         ${PVC_STORAGE_SIZE} ${OS_BUCKET_NAME} ${OS_ACCESS_KEY_ID} ${OS_SECRET_KEY} ${OS_ENDPOINT} \
         --dynamic-etcd-nodes=1 \
         --dry-run > pachyderm-with-statefulset.json
```
Note: we use `any-string` as the first argument above because, 
while the `deploy custom` command expects 6 arguments, 
it will ignore the first argument when deploying with StatefulSets.
### Configuring with StatefulSets using StorageClasses
```sh
$ pachctl deploy custom --object-store s3 any-string 
         ${PVC_STORAGE_SIZE} ${OS_BUCKET_NAME} ${OS_ACCESS_KEY_ID} ${OS_SECRET_KEY} ${OS_ENDPOINT} \
         --dynamic-etcd-nodes=1  --etcd-storage-class $PVC_STORAGECLASS \
         --dry-run > pachyderm-with-statefulset-using-storageclasses.json
```
Note: we use `any-string` as the first argument above because, 
while the `deploy custom` command expects 6 arguments, 
it will ignore the first argument when deploying with StatefulSets.

## Next steps

You may either deploy manifests you created above or edit them to customize them further, prior to deploying.

### Editing your manifest to customize it further

This functionality requires an experienced Kubernetes administrator.
If you are attempting a highly customized deployment, 
use one of the Pachyderm support resources below.

### Deploying
The command you'll want to run depends on the command you ran, above.

#### Deploying with a static persistent volume
```sh
$ kubectl apply -f ./pachyderm-with-static-volume.json
```
#### Deploying with StatefulSets
```sh
$ kubectl apply -f ./pachyderm-with-statefulset.json
```
#### Deploying with StatefulSets using StorageClasses
```sh
$ kubectl apply -f ./pachyderm-with-statefulset-using-storageclasses.json
```


