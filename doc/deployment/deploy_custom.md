# Custom Deployments

If you are deploying Pachyderm to a cloud infrastructure, 
such as [Amazon Web Services (AWS)](https://pachyderm.readthedocs.io/en/latest/deployment/amazon_web_services.html),
[Google Cloud Platform (GCP)](https://pachyderm.readthedocs.io/en/latest/deployment/google_cloud_platform.html), or 
[Microsoft Azure](https://pachyderm.readthedocs.io/en/latest/deployment/azure.html), 
use a related `pachctl deploy` subcommand, such as `amazon`, `google`, or `microsoft`, respectively.
Also, you can customize cloud provider deployments extensively through flags available for each provider.

Pachyderm includes `pachctl deploy custom` for creating customized deployments for cloud providers or on-premises use.
Typically, you customize a deployment by running the command with the `--dry-run` flag.
The command's standard output is directed into a series of customization scripts or a file for editing.

This section describes how to use `pachctl deploy custom ... --dry-run` to create a manifest for a custom, on-premises deployment.
Although deployment automation is out of scope of this section, Pachyderm strongly encourages you to configure your [infrastructure as code](./on-premises.html#infrastructure-as-code).
but we do encourage you to treat your [infrastructure as code](./on-premises.html#infrastructure-as-code).


## Creating a Pachyderm deployment manifest

The command to create a custom manifest is `pachctl deploy custom`,
which takes two sets of required, primary flags, 
one required flag from either of two possible flags,
and one set of optional flags.
The first two sets of required flags configure the primary components for a Pachyderm deployment, 
the persistent volume and the object store.
They take a parameter to indicate the style of pv and object store backend.
Those parameters will drive the kind and number of arguments that follow the two required flags.
The last require flag configures the type of etcd deployment: static volume or StatefulSet.
An additional set of optional flags configures other deployment parameters.

A `pachctl deploy custom` invocation looks like this
```
pachctl deploy custom --persistent-disk <persistent disk backend> --object-store <object store backend> <persistent disk args> <object store args> [configuration flags]
```

Let's look at each set of flags in turn

### Persistent volume configuration

The `--persistent-disk` flag takes on the style of pv backend.
Pachyderm currently only has automated configuration for styles of backend for the major cloud providers: 

[//]: # (todo:fill this out better)

-amazon
-google
-azure

For each of those providers, 
different configurations will result depending on that third required deployment flag.
That third one is either of these flags, 
`--dynamic-etcd-nodes` or 
`--static-etc-volume`.

[//]: # (todo: provide links to statefulsets)

`--dynamic-etcd-nodes` is used when your Kubernetes installation has been configured to use StatefulSets. 
It takes as an parameter the number of `etcd` nodes which your deployment will create.
Pachyderm recommends you keep this number at 1.
Consult with your Pachyderm support team if you want to change it.
This flag will create a `VolumeClaimTemplate` in the `etcd` `StatefulSet` that uses the standard `etcd-storage-class`.
Consult with your Kubernetes administrator on the availability of this storage class in your Kubernetes deployment.

`--static-etc-volume` is used when your Kubernetes installation has been configured to use static volumes.
It'll create a PV with a spec appropriate for one of cloud providers 
(gcePersistentDisk for Google Cloud Storage, 
awsElasticBlockStore for Amazon Web Services, 
azureDisk for Microsoft Azure).
Of course, 
these choices are not relevant for most on-premises deployments,
so you will need to consult with your Kubernetes administrators to figure out the correct choices for your infrastructure.

[//]: # (todo: provide links to storage manifest sections)

In the section on that storage manifests,
we'll give you pointers to some common ones.


### Object store configuration
- [object store configuration](#object-store-configuration): `--object-store` immediately after the flag to 

- [other flags]: various flags to configure the deployment.
Each of these flagged arguments takes one parameter which indicates the style of configuration.
Different styles of configuration will take further downstream arguments to further configure them.


## Anatomy of a Pachyderm deployment manifest

When you run the `pachctl deploy ...` command with the `--dry-run` flag,
you are generating a JSON-encoded Kubernetes manifest in one stream to standard output. 
That manifest consists of a number of smaller manifests,
that correspond to a particular aspect of a Pachyderm deployment.

Pachyderm deploys the following sets of application components:
- `pachd`, the main Pachyderm pod
- `etcd`, the administrative datastore for `pachd`
- `dash`, the web-based enterprise ui for Pachyderm

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

If you did not use [StatefulSets](./on_premises.html#statefulsets) to deploy Pachyderm,
that is, you do not specify `--dynamic-etcd-nodes` flag, 
the value that you specify for `--persistent-disk` causes `pachctl` to write a manifest for creating a [`PersistentVolume`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) that Pachyderm's `etcd` uses in its [`PersistentVolumeClaim`](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims).

### PersistentVolumeClaim

Pachyderm's `etcd` uses this `PersistentVolumeClaim` unless you deploy using [StatefulSets](./on_premises.html#statefulsets).
See this manifest's name in the `etcd` Deployment manifest, below.

### StorageClass

If you *do* use [StatefulSets](./on_premises.html#statefulsets) to deploy Pachyderm
that is, you use `--dynamic-etcd-nodes` flag, 
this manifest specifies the kind of storage and provisioner that is appropriate for what you have specified in the `--persistent-disk` flag. 
You won't see this manifest if you specified `azure` as the argument to `--persistent-disk`.

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

## Prerequisites

### Software you will need 
    
1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
2. [pachctl](http://docs.pachyderm.io/en/latest/pachctl/pachctl.html)

### Preparing your environment

See the [introduction to on-premises deployment](./on_premises.html) for steps that you need to take before to creating a custom Pachyderm deployment manifest.

### Customizing `pachctl` flags

`pachctl` includes flags for customizing aspects of your deployment,
from memory and cpu requests for `etcd` and `pachd` to specifying a Kubernetes namespace.

You can learn what flags are available in your version of Pachyderm by running `pachctl deploy custom --help`.
Some of the flags are marked as to be used with caution.
When you are unsure of the effect of a flag, consult with your Kubernetes administrator and your Pachyderm support team.
Some flags, such as `--image-pull-secret`, require the creation and loading of Kubernetes manifests outside of `pachctl`.

## Creating a Pachyderm manifest

Please see the [introduction to on-premises deployment](./on_premises.html) for an explanation of the differences among static persistent volumes, StatefulSets and StatefulSets with StorageClasses, as well as the meanings of the variables, like  `PVC_STORAGE_SIZE` and `OS_ENDPOINT`, used below.

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

### Deploying
The command you'll want to run depends on the command you ran, above.

#### Deploying with a static persistent volume
```sh
$ kubectl apply -f ./pachyderm-with-static-volume.json
```
#### Deploying  with StatefulSets
```sh
$ kubectl apply -f ./pachyderm-with-statefulset.json
```
#### Deploying  with StatefulSets using StorageClasses
```sh
$ kubectl apply -f ./pachyderm-with-statefulset-using-storageclasses.json
```


