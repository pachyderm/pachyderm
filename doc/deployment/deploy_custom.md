
# Custom Deployments

Pachyderm includes the `pachctl deploy custom` command for creating customized deployments
for cloud providers or on-premises use.

This section describes how to use `pachctl deploy custom` to create a manifest for a custom, on-premises deployment.
Although deployment automation is out of scope of this section, Pachyderm strongly encourages you  to treat your [infrastructure as code](./on-premises.html#infrastructure-as-code).

This document describes that customization in two primary parts: 
invoking `pachctl deploy custom` to create a custom manifest 
and examining that manifest in detail.
We also go through some examples and
detail the next steps for editing and deployment.

- [Prerequisites](prerequisites)
  - [Software you will need](software-you-will-need)
  - [Preparing your environment](preparing-your-environment)
- [Creating a Pachyderm deployment manifest](#creating-a-pachyderm-deployment-manifest)
  - [Persistent disk configuration](#persistent-disk-configuration)
    - [Persistent disk parameters](#persistent-disk-parameters)
    - [Example invocation with persistent disk parameters](#example-invocation-with-persistent-disk-parameters)
  - [Object store configuration](#object-store-configuration)
    - [Object store parameters](#object-store-parameters)
    - [Example invocation with pv and object store](example-invocation-with-pv-and-object-store)
  - [Additional flags](#additional-flags)
    - [Local vs cluster roles](local-vs-cluster-roles)
    - [Resource requests and limits](resource-requests-and-limits)
    - [Enterprise Edition features](enterprise-edition-features)
    - [Output formats](output-formats)
    - [Logging](logging)
    - [Complete example invocation](complete-example-invocation)
- [Anatomy of a Pachyderm deployment manifest](anatomy-of-a-pachyderm-deployment-manifest)
  - [Roles and permissions manifests](roles-and-permissions-manifests)
    - [ServiceAccount](serviceaccount)
    - [Role or ClusterRole](role-or-clusterrole)
    - [RoleBinding or ClusterRoleBinding](rolebinding-or-clusterrolebinding)
  - [Application-related](application-related)
    - [PersistentVolume](persistentvolume)
    - [PersistentVolumeClaim](persistentvolumeclaim)
    - [StorageClass](storageclass)
    - [Service](service)
  - [The Pachyderm pods](the-pachyderm-pods)
    - [Deployment](deployment)
    - [StatefulSet](statefulset)
  - [Secret](secret)
- [More examples](more-examples)
  - [Configuring with a static persistent volume](configuring-with-a-static-persistent-volume)
  - [Configuring with StatefulSets](configuring-with-statefulsets)
  - [Configuring with StatefulSets using StorageClasses](configuring-with-statefulsets-using-storageclasses)
- [Next steps](next-steps)
  - [Editing your manifest to customize it further](editing-your-manifest-to-customize-it-further)
  - [Deploying with a static persistent volume](deploying-with-a-static-persistent-volume)
  - [Deploying with StatefulSets](deploying-with-statefulsets)
  - [Deploying with StatefulSets using StorageClasses](deploying-with-statefulsets-using-storageclasses)

## Prerequisites

### Software you will need 
    
1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
2. [pachctl](http://docs.pachyderm.io/en/latest/pachctl/pachctl.html)

### Preparing your environment

See the [introduction to on-premises deployment](./on_premises.html) for steps that you need to take 
before creating a custom Pachyderm deployment manifest. 
That document also provides an explanation of the differences among static persistent volumes, 
StatefulSets and StatefulSets with StorageClasses, 
as well as the meanings of the variables, 
like  `PVC_STORAGE_SIZE` and `OS_ENDPOINT`, 
used in examples below.

## Creating a Pachyderm deployment manifest

The command to create a custom manifest is `pachctl deploy custom`.

An invocation looks like this
```
pachctl deploy custom --persistent-disk <persistent disk backend> --object-store <object store backend> \
    <persistent disk arg1>  <persistent disk arg 2> \
    <object store arg 1> <object store arg 2>  <object store arg 3>  <object store arg 4> \
    [[--dynamic-etcd-nodes n] | [--static-etcd-volume <volume name>]]
    [optional flags]
```

Let's look at each set of flags in turn. 
As we go through each set of flags, 
we'll build up a sample `pachctl deploy custom`  deployment invocation for 
- a StatefulSet deployment to
- an on-premises vanilla Kubernetes cluster 
  - with the standard etcd StorageClass configured along with
  - access controls that limit us to namespace-local roles only and
- an on-premises Min.io object store
  - with SSL turned on,
  - S3v4 signatures (explained below; this is the default choice),
  - the endpoint `minio:9000`
  - the access key `OBSIJRBE0PP2NO4QOA27`,
  - the secret key `tfteSlswRu7BJ86wekitnifILbZam1KYY3TG`,
  - and a bucket named `pachyderm-bucket`.

We'll save the output of the the invocation to a file that will be used by our [code infrastructure](./on-premises.html#infrastructure-as-code) to configure our deployment. 

Our scripts in that infrastructure work with YAML manifests.
  
### Persistent disk configuration

The `--persistent-disk` flag takes on the style of pv backend.
Pachyderm currently only has automated configuration for styles of backend for the major cloud providers: 

- aws
- google
- azure

For each of those providers, 
different configurations will result depending on a third required deployment flag, one of 
`--dynamic-etcd-nodes` or 
`--static-etc-volume`.

`--dynamic-etcd-nodes` is used when your Kubernetes installation has been configured to use [StatefulSets](./on_premises.html#statefulsets). 
StatefulSets is a useful technology which has been in stable releases of Kubernetes since 2018.
It is likely that your on-premises Kubernetes installation is configured to use StatefulSets.

The `--dynamic-etcd-nodes` flag has a parameter which specifies the number of `etcd` nodes which your deployment will create.
Pachyderm recommends you keep this number at 1.
Consult with your Pachyderm support team if you want to change it.

This flag will create a `VolumeClaimTemplate` in the `etcd` `StatefulSet` that uses the standard `etcd-storage-class`.

**NOTE** Consult with your Kubernetes administrator on the StorageClass you should use for `etcd` in your Kubernetes deployment.
If it's not the default, you can use the flag `--etcd-storage-class` to specify the StorageClass.

`--static-etc-volume` is used when your Kubernetes installation has not been configured to use StatefulSets.
It will use a static volume with Pachyderm's `etcd`, 
creating a PV with a spec appropriate for each of cloud providers specified above:

- `aws`: awsElasticBlockStore for Amazon Web Services
- `azure`: azureDisk for Microsoft Azure
- `google`: gcePersistentDisk for Google Cloud Storage

Of course, 
these choices are not relevant for most on-premises deployments,
so you will need manually edit your manifest 
after consulting with your Kubernetes administrators
to determine the correct choices for your infrastructure.

The section on that [storage manifests](#persistentvolume) goes into a little more detail on this.

#### Persistent disk parameters

Regardless whether you choose to deploy with StatefulSets or static volumes,
the `--persistent-disk` flag takes two arguments
that you specify right after the single argument to the `--object-store` flag.

[The first argument is always ignored,
but must be present.](https://github.com/pachyderm/pachyderm/issues/3312)
You may set it to any text value you like.

The second argument is the size, 
in gigabytes,
that will be requested for `etcd`'s disk.
A good value for most deployments is 10.

#### Example invocation with persistent disk parameters
Our example on-premises cluster has StatefulSets enabled,
with the standard etcd storage class configured.
Our deployment flags for our sample cluster look like this, so far:
```
pachctl deploy custom --persistent-disk aws --object-store <object store backend> \
    foobar 10 \
    <object store arg 1> <object store arg 2>  <object store arg 3>  <object store arg 4> \
    --dynamic-etcd-nodes 10
    [optional flags]
```
Note that the first argument for the persistent disk backend is ignored,
so we just put the common `foobar` epithet in the invocation.
In this case, even the `aws` persistent disk type is irrelevant,
as all the cloud providers use the standard etcd storage class.
The VolumeClaimTemplate manifests will be identical regardless of whether `aws`, `google` or `azure` is specified,
when using StatefulSets.
That's not the case for static volumes, of course.

### Object store configuration

The flag `--object-store` is used to configure Pachyderm to use one of two object store drivers.
It takes one argument, [which must be the value `s3`](https://github.com/pachyderm/pachyderm/issues/3996).
This will use the Amazon S3 driver to access your on-premises object store, 
regardless of the vendor,
since the Amazon S3 API is the standard that every object store is designed to work with.

However, the S3 API has two different extant versions of "signature styles", 
which are how the object store validates client requests.
S3v4 is the most current version,
but there are many S3v2 object store servers in the field.
[Amazon itself has announced the end-of-life of S3v2-type signatures on its own service](https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingAWSSDK.html#UsingAWSSDK-sig2-deprecation),
and their own drivers don't support it any more.

If you need to access an object store that uses S3v2 signatures,
you can specify the flag `--isS3V2`. 

This will configure Pachyderm to use the Min.io driver,
allowing the use of the older signature.
Using this flag will also disable SSL for connections to the object store with the `minio` driver.
You can reenable it with the `-s` or `--secure` flag.
You can also edit the `pachyderm-storage-secret` Kubernetes manifest it produces manually,
which is detailed below.

#### Object store parameters

The `--object-store` flag takes four (4) required, additional configuration arguments.
These arguments must be placed immediately after [the persistent disk configuration parameters](#persistent-disk-parameters).

- _bucket-name_: the name of the bucket, without the `s3://` prefix or a trailing `/`.
- _access-key_: the user access id used to access the object store.
- _secret-key_: the associated password used with the user access id to access the object store.
- _endpoint_: the hostname and port used to access the object store, in <hostname>:<port> format.

#### Example invocation with pv and object store
Our example on-premises cluster will use an on-premises Min.io object store
  - with SSL turned on,
  - S3v4 signatures,
  - the endpoint `minio:9000`
  - the access key `OBSIJRBE0PP2NO4QOA27`,
  - the secret key `tfteSlswRu7BJ86wekitnifILbZam1KYY3TG`,
  - and a bucket named `pachyderm-bucket`.
Our deployment flags for our sample cluster look like this, so far:
```
pachctl deploy custom --persistent-disk aws --object-store s3 \
    foobar 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 10
    [optional flags]
```
Note that we enclosed the arguments that may contain characters that the shell could interpret in single-quotes.

### Additional flags

#### Local vs cluster roles

The `--local-roles` flag is used to change the kind of role the `pachyderm` service account will use from cluster-wide (`ClusterRole`) to namespace-specific (`Role`). 
Using `--local-roles` will inhibit your ability to use a Pachyderm feature called [coefficient parallelism](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#parallelism-spec-optional).
You'll see this message in the `pachd` pod's logs on Kubernetes.

```
ERROR unable to access kubernetes nodeslist, Pachyderm will continue to work but it will not be possible to use COEFFICIENT parallelism. error: nodes is forbidden: User "system:serviceaccount:pachyderm-test-1:pachyderm" cannot list nodes at the cluster scope
```

#### Resource requests and limits

Larger deployments may need to request more resources for `pachd` and `etcd`
or set higher limits for transient workloads.
They are used to set attributes which are passed on to Kubernetes directly via the produced manifest.
None of these flags should be modifed from the default values for production deployments without consulting Pachyderm support.
- `--etcd-cpu-request`: the number of cores Kubernetes should give `etcd`. Fractions are allowed. 
- `--etcd-memory-request`: the amount of memory Kubernetes should give `etcd`. Accepts the SI suffixes that Kubernetes accepts for such values.
- `--no-guaranteed`: turn of QoS for `etcd` and `pachd`. Not to be used in production environments.
- `--pachd-cpu-request`: the number of cores Kubernetes should give `pachd`. Fractions are allowed. 
- `--pachd-memory-request`: the amount of memory Kubernetes should give `pachd`. Accepts the SI suffixes that Kubernetes accepts for such values.
- `shards`: the maximum number of `pachd` nodes allowed in the cluster. Increasing this number from the default value of 16 may result in degraded performance.

#### Enterprise Edition features

- `--dash-image`: the docker image for the Pachyderm Enterprise Edition dashboard
- `--image-pull-secret`: the name of a Kubernetes secret Pachyderm will need to pull from a private docker registry
- `--no-dashboard`: don't create a manifest for deploying the Enterprise Edition dashboard
- `--registry`: the registry for docker images
- `--tls`:  string of the form `"<cert path>,<key path>"` of the signed TLS certificate used for encrypting pachd communications

#### Output formats

- `--dry-run`: don't actually deploy to Kubernetes, just send the manifest to standard output
- `-o` or `--output`: choose from json (the default) or yaml

#### Logging
- `log-level`: sets the verbosity level of `pachd`, from most verbose to least the settings are `debug`, `info`, and `error`.
- `-v` or `--verbose`: controls the chattiness of the `pachctl` invocation itself.

#### Complete example invocation
Since we're deploying to a code infrastructure that works with YAML files,
we'll add the flags for that.
We need to limit ourselves to local roles only, so we'll add the `--local-roles` flag.
We'll also save the output to a file that our code infrastructure will use.
Our final deployment looks like this:
```
pachctl deploy custom --persistent-disk aws --object-store s3 \
    foobar 10 \
    pachyderm-bucket  'OBSIJRBE0PP2NO4QOA27' 'tfteSlswRu7BJ86wekitnifILbZam1KYY3TG' 'minio:9000' \
    --dynamic-etcd-nodes 10
    --local-roles --output yaml  --dry-run > custom_deploy.yaml
```

What does a file like `custom_deploy.yaml` have inside of it?
That's in the next section: 
a general exploration of all the deployment manifests that `pachctl deploy custom` produces.


## Anatomy of a Pachyderm deployment manifest

When you run the `pachctl deploy ...` command with the `--dry-run` flag,
you are generating, by default, a JSON-encoded Kubernetes manifest in one stream to standard output. 
That manifest consists of a number of smaller manifests,
each of which corresponds to a particular aspect of a Pachyderm deployment.

Pachyderm deploys the following sets of application components:
- `pachd`, the main Pachyderm pod
- `etcd`, the administrative datastore for `pachd`
- `dash`, the web-based enterprise ui for Pachyderm Enterprise Edition

In general, there are two categories of manifests in the file,
roles-and-permissions-related and application-related. 


## Roles and permissions manifests

### ServiceAccount

Typically at the top of the file, a roles and permissions manifest has the `kind` key set to `ServiceAccount`. 
[ServiceAccounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/) are a way Kubernetes can assign namespace-specific privileges to applications in a lightweight way.
The Pachyderm's service account is called  `pachyderm`.

### Role or ClusterRole

Depending on whether you used the `--local-roles` flag or not, the next manifest will be of `kind` `Role` or `ClusterRole`.
depending on whether you  used the `--local-roles` flag `pachctl deploy` command.

### RoleBinding or ClusterRoleBinding

This manifest binds the `Rule` or `ClusterRole` to the `ServiceAccount` created above.

## Application-related

### PersistentVolume

If you used `--static-etc-volume` to deploy Pachyderm,
the value that you specify for `--persistent-disk` causes `pachctl` to write a manifest for creating a [`PersistentVolume`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) that Pachyderm's `etcd` uses in its [`PersistentVolumeClaim`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims).

A common persistent volume uses in enterprises is an NFS mount backed by a storage fabric of some sort.
In this case, a StorageClass for an NFS mount will be made available for consumption.
Consult with your Kubernetes administrators to learn what resources are available for your deployment

### PersistentVolumeClaim

Pachyderm's `etcd` uses this `PersistentVolumeClaim` if you deployed using `--static-etc-volume`.
See this manifest's name in the `etcd` Deployment manifest below.

### StorageClass

If you used the `--dynamic-etcd-nodes` flag to deploy Pachyderm,
this manifest specifies the kind of storage and provisioner that is appropriate for what you have specified in the `--persistent-disk` flag. 

**Note:** You will not see this manifest if you specified `azure` as the argument to `--persistent-disk`.

### Service

In a typical Pachyderm deployment, you see three [`Service`](https://kubernetes.io/docs/concepts/services-networking/service/) manifests. 
Services are how Kubernetes exposes Pods to the network.
If you  use StatefulSets to deploy Pachyderm,
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
The final manifest uses the command-line arguments that you submit to the `pachctl deploy` command to store the parameters, 
like region, secret, token, and endpoint that are used to access an object store. 
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


