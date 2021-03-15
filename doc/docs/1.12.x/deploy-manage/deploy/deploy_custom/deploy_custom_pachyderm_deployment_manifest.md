# Pachyderm Deployment Manifest

This section provides an overview of the Kubernetes manifest that
you use to deploy your Pachyderm cluster. This section is provided for
your reference and does not include configuration steps. If you are
familiar with Kubernetes or do not have immediate questions about
the configuration parameters, you can skip this section and proceed to
[Configuring Persistent Disk Parameters](deploy_custom_configuring_persistent_disk_parameters.md).

When you run the `pachctl deploy` command, Pachyderm generates a JSON-encoded
Kubernetes manifest which consists of sections that
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
| `--persistent-disk`    | Configures the storage resource for etcd. Pachyderm uses etcd <br> to manage administrative metadata. User data is stored in an   <br> object store, not in etcd. |
| `--object-store`       | Configures the object store that Pachyderm uses for storing all <br> user data that you want to be versioned and managed. |
| Optional flags         | Optional flags that are not required to deploy Pachyderm but <br> enable you to configure access, output format, logging         verbosity, <br>and other parameters. |

## Kubernetes Manifest Parameters

Your Kubernetes manifest includes sections that describe
the configuration of your Pachyderm cluster.

The manifest includes the following sections:

**Roles and permissions manifests**

| Manifest | Description |
| ------- | ----------- |
| `ServiceAccount` | Typically at the top of the manifest file Pachyderm produces, <br> a roles and permissions manifest has the `kind` key set to `ServiceAccount`. <br> Kubernetes uses `ServiceAccounts` to assign namespace-specific privileges<br> to applications in a lightweight way. Pachyderm's <br> service account is called  `pachyderm`. |
| `Role` or `ClusterRole` | Depending on whether you used the `--local-roles` flag or not, <br>the next manifest kind is Role or ClusterRole. |
| `RoleBinding` or <br> `ClusterRoleBinding` | This manifest binds the `Role` or `ClusterRole` to the <br>`ServiceAccount` created above. |

**Application-related manifests**

| Manifest | Description |
| ------- | ----------- |
| `PersistentVolume` | If you used `--static-etcd-volume` to deploy Pachyderm, the <br> value that you specify for `--persistent-disk` causes `pachctl` to write <br> a manifest for creating a `PersistentVolume` that Pachyderm’s etcd uses in its <br> `PersistentVolumeClaim`. A common persistent volume that is used in <br>enterprises is an NFS mount backed by a storage fabric. In this case, a <br> `StorageClass` for an NFS mount is made available for consumption. Consult <br> with your Kubernetes administrators to learn what resources are available <br> for your deployment. |
| `PersistentVolumeClaim` | If you deployed Pachyderm by using `--static-etcd-volume`, <br> the Pachyderm's `etcd` store uses this `PersistentVolumeClaim`.<br> See this manifest's name in the Deployment manifest <br> for the `etcd` pod, described below in Pachyderm pods manifests. |
| `StorageClass` | If you used the `--dynamic-etcd-nodes` flag to deploy Pachyderm, <br> this manifest specifies the kind of storage and provisioner that <br>is appropriate for what you have specified in the `--persistent-disk` flag. <br> **Note:** You will not see this manifest if you specified `azure` as the argument <br> to `--persistent-disk`, since Azure has their own provisioner. |
| `Service` | In a typical Pachyderm deployment, you see three `Service` manifests.<br> A `Service` is a Kubernetes abstraction that exposes `Pods` to the network. <br> If you use `StatefulSets` to deploy Pachyderm, that is, you used the <br>`--dynamic-etcd-nodes` flag, Pachyderm deploys one `Service` for `etcd-headless`, <br>one for `pachd`, and one for `dash`. A static deployment has `Services` <br> for `etcd`, `pachd`, and `dash`. If you use the `--no-dashboard` flag, <br>Pachyderm does not create a `Service` and `Deployment` for the dashboard.<br> Similarly, if `--dashboard-only` is specified, Pachyderm generates <br> the manifests for the Pachyderm enterprise UI only. The most common items <br>that you can edit in `Service` manifests are the `NodePort` values for<br> various services, and the `containerPort` values for `Deployment` manifests.<br> To make your `containerPort` values work properly, you will need to <br> add environment variables to a `Deployment` or `StatefulSet` object. <br> You can see how what environment variables to add in the [OpenShift](../openshift.md) example.

**Pachyderm pods manifests**

| Manifest | Description |
| -------- | ----------- |
| `Deployment` | Declares the desired state of application pods to Kubernetes. If you <br> configure a static deployment, Pachyderm deploys `Deployment` manifests for <br> `etcd`, `pachd`, and `dash`. If you specify `--dynamic-etcd-nodes`, <br>Pachyderm deploys the `pachd` and `dash` as `Deployment` and `etcd` as a <br> `StatefulSet`. If you run the deploy command with the `--no-dashboard`<br> flag, Pachyderm omits the deployment of the `dash` `Service` and `Deployment`.
| `StatefulSet` | For a `--dynamic-etcd-nodes` deployment, Pachyderm replaces the `etcd` `Deployment` <br> manifest with a `StatefulSet`. |

**Pachyderm Kubernetes secrets manifests**

| Manifest | Description |
| -------- | ----------- |
| `Secret`      | Pachyderm uses the Kubernetes `Secret` manifest to store the credentials that <br> are necessary to access object storage. The final manifest uses the <br> command-line arguments that you submit to the `pachctl deploy` <br> command to store such parameters as region, secret, token, and endpoint, that are <br>used to access an object store. The exact values in the secret <br> depend on the kind of object store you configure for your deployment. You <br>can update the values after the deployment either by using `kubectl` <br>to deploy a new `Secret` or the `pachctl deploy storage` command. Note: If you want to modify the contents of the secret directly, make sure you understand that on-premise deployments of non-cloud object stores (e.g. MinIO or ECS) also use the values prefixed with `amazon-` and not the ones prefixed with `minio-`.|
