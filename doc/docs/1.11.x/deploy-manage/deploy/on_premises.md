# On Premises

This document is broken down into the following sections, available at the links below

- [Introduction to on-premises deployments](#introduction) takes you through what you need to know about Kubernetes, persistent volumes, object stores and best practices.  That's this page.
- [Customizing your Pachyderm deployment for on-premises use](deploy_custom/index.md) details the various options of the `pachctl deploy custom ...` command for an on-premises deployment.
- [Single-node Pachyderm deployment](./single-node.md) is the document you should read when deploying Pachyderm for personal, low-volume usage.
- [Registries](./docker_registries.md) takes you through on-premises, private Docker registry configuration.
- [Ingress](./configuring_k8s_ingress.md) details the Kubernetes ingress configuration you'd need for using `pachctl` and the dashboard outside of the Kubernetes cluster
- [Non-cloud object stores](./non-cloud-object-stores.md) discusses common configurations for on-premises object stores.

Need information on a particular flavor of Kubernetes or object store?  Check out the [see also](#see-also) section.

Troubleshooting a deployment? Check out [Troubleshooting Deployments](../../troubleshooting/deploy_troubleshooting.md).

## Introduction

Deploying Pachyderm successfully on-premises requires a few prerequisites and some planning.
Pachyderm is built on [Kubernetes](https://kubernetes.io/).
Before you can deploy Pachyderm, you or your Kubernetes administrator will need to perform the following actions:

1. [Deploy Kubernetes](#deploying-kubernetes) on-premises.
1. [Deploy a Kubernetes persistent volume](#deploying-a-persistent-volume) that Pachyderm will use to store administrative data.
1. [Deploy an on-premises object store](#deploying-an-object-store) using a storage provider like [MinIO](https://min.io), [EMC's ECS](https://www.dellemc.com/storage/ecs/index.htm), or [SwiftStack](https://www.swiftstack.com/) to provide S3-compatible access to your on-premises storage.
1. [Create a Pachyderm manifest](deploy_custom/deploy_custom_pachyderm_deployment_manifest.md) by running the `pachctl deploy custom` command with appropriate arguments and the `--dry-run` flag to create a Kubernetes manifest for the Pachyderm deployment.
1. [Edit the Pachyderm manifest](deploy_custom/deploy_custom_pachyderm_deployment_manifest.md) for your particular Kubernetes deployment

In this series of documents, we'll take you through the steps unique to Pachyderm.
We assume you have some Kubernetes knowledge.
We will point you to external resources for the general Kubernetes steps to give you background.

## Best practices
### Infrastructure as code

We highly encourage you to apply the best practices used in developing software to managing the deployment process.

1. Create scripts that automate as much of your processes as possible and keep them under version control.
1. Keep copies of all artifacts, such as manifests, produced by those scripts and keep those under version control.
1. Document your practices in the code and outside it.

### Infrastructure in general

Be sure that you design your Kubernetes infrastructure in accordance with recommended guidelines.
Don't mix on-premises Kubernetes and cloud-based storage.
It's important that bandwidth to your storage deployment meet the guidelines of your storage provider.

## Prerequisites

### Software you will need 
    
1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
2. [pachctl](../../../../../getting_started/local_installation/#install-pachctl)

## Setting up to deploy on-premises

### Deploying Kubernetes

The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/getting-started-guides/#on-premises-vms).
We recommend following one of these guides to get Kubernetes running on premise.

### Deploying a persistent volume

#### Persistent volumes: how do they work?

A Kubernetes [persistent volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) is used by Pachyderm's `etcd` for storage of system metatada. 
In Kubernetes, [persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) are a mechanism for providing storage for consumption by the users of the cluster.
They are provisioned by the cluster administrators.
In a typical enterprise Kubernetes deployment, the administrators have configured persistent volumes that your Pachyderm deployment will consume by means of a [persistent volume claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) in the Pachyderm manifest you generate.

You can deploy PV's to Pachyderm using our command-line arguments in three ways: using a static PV, with StatefulSets, or with StatefulSets using a StorageClass.

If your administrators are using [selectors](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#selector), or you want to use StorageClasses in a different way, you'll need to [edit the Pachyderm manifest](../deploy_custom/deploy_custom_pachyderm_deployment_manifest) appropriately before applying it.

##### Static PV

In this case, `etcd` will be deployed in Pachyderm as a [ReplicationController](https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/) with one (1) pod that uses a static PV. This is a common deployment for testing. 

##### StatefulSets

[StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) are a mechanism provided in  Kubernetes 1.9 and newer to manage the deployment and scaling of applications.  It uses either [Persistent Volume Provisioning](https://github.com/kubernetes/examples/blob/master/staging/persistent-volume-provisioning/README.md) or pre-provisioned PV's. 

If you're using StatefulSets in your Kubernetes cluster, you will need to find out the particulars of your cluster's PV configuration and [use appropriate flags to `pachctl deploy custom`](#configuring-with-statefulsets)

##### StorageClasses 
If your administrators require specification of [classes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1) to consume persistent volumes, 
you will need to find out the particulars of your cluster's PV configuration and [use appropriate flags to `pachctl deploy custom`](#configuring-with-statefulsets-using-storageclasses).

#### Common tasks to all types of PV deployments
##### Sizing the PV

You will need to use a PV with enough space for the metadata associated with the data you plan to store in Pachyderm. 

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your cloud provider, and disk generation, you may need to oversize the volume significantly to ensure enough IOPS (i.e. 500G for EBS gp2, 50G for GCP ssd).

##### Creating the PV

In the case of cloud-based deployments, the `pachctl deploy` command for AWS, GCP and Azure creates persistent volumes for you, when you follow the instructions for those infrastructures.

In the case of on-premises deployments, the kind of PV you provision will be dependent on what kind of storage your Kubernetes administrators have attached to your cluster and configured, and whether you are expected to consume that storage as a static PV, with Persistent Volume Provisioning  or as a StorageClass.

For example, many on-premises deployments use Network File System (NFS) to access to some kind of enterprise storage.
Persistent volumes are provisioned in Kubernetes like all things in Kubernetes: by means of a manifest.
You can learn about creating [volumes](https://kubernetes.io/docs/concepts/storage/volumes/)  and [persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) in the Kubernetes documentation.

You or your Kubernetes administrators will be responsible for configuring the PVs you create to be consumable as static PV's, with Persistent Volume Provisioning or as a StorageClass.

#### What you'll need for Pachyderm configuration of PV storage

Keep the information below at hand for when you [run `pachctl deploy custom` further on](deploy_custom/index.md)

##### Configuring with static volumes

You'll need the name of the PV and the amount of space you can use, in gigabytes.
We'll refer to those, respectively, as `PVC_STORAGE_NAME` and `PVC_STORAGE_SIZE` further on.
With this kind of PV,
you'll use the flag `--static-etcd-volume` with `PVC_STORAGE_NAME` as its argument in your deployment.

Note: this will override any attempt to configure with StorageClasses, below.

##### Configuring with StatefulSets

If you're deploying using [StatefulSets](#statefulsets),
you'll just need the amount of space you can use, in gigabytes, 
which we'll refer to as `PVC_STORAGE_SIZE` further on..

Note: The `--etcd-storage-class` flag and argument will be ignored if you use the flag `--static-etcd-volume` along with it.

##### Configuring with StatefulSets using StorageClasses

If you're deploying using [StatefulSets](#statefulsets) with [StorageClasses](#storageclasses), 
you'll need the name of the storage class and the amount of space you can use, in gigabytes.
We'll refer to those, respectively, as `PVC_STORAGECLASS` and `PVC_STORAGE_SIZE` further on.
With this kind of PV,
you'll use the flag `--etcd-storage-class` with `PVC_STORAGECLASS` as its argument in your deployment. 

Note: The `--etcd-storage-class` flag and argument will be ignored if you use the flag `--static-etcd-volume` along with it.

   
### Deploying an object store

#### Object store: what's it for?
An object store is used by Pachyderm's `pachd` for storing all your data. 
The object store you use must be accessible via a low-latency, high-bandwidth connection like [Gigabit](https://en.wikipedia.org/wiki/Gigabit_Ethernet)  or [10G Ethernet](https://en.wikipedia.org/wiki/10_Gigabit_Ethernet).

For an on-premises deployment, 
it's not advisable to use a cloud-based storage mechanism.
Don't deploy an on-premises Pachyderm cluster against cloud-based object stores such as S3 from [AWS](amazon_web_services/index.md), GCS from [Google Cloud Platform](google_cloud_platform.md), Azure Blob Storage from [Azure](azure.md). Note that the command line parameters for the object store (`--object-store`) are specifying `s3` in reference to the S3 protocol (which is used by solutions such as MinIO and the like) and not the Amazon product with the same name.

#### Object store prerequisites

Object stores are accessible using the S3 protocol, created by Amazon. 
Storage providers like [MinIO](https://min.io), [EMC's ECS](https://www.dellemc.com/storage/ecs/index.htm), or [SwiftStack](https://www.swiftstack.com/) provide S3-compatible access to enterprise storage for on-premises deployment. 
You can find links to instructions for providers of particular object stores in the [See also](#see-also) section.

#### Sizing the object store

Size your object store generously.
Once you start using Pachyderm, you'll start versioning all your data.
We're currently developing good rules of thumb for scaling your object store as your Pachyderm deployment grows,
but it's a good idea to start with a large multiple of your current data set size.

#### What you'll need for Pachyderm configuration of the object store
You'll need four items to configure the object store. 
We're prefixing each item with how we'll refer to it further on.

1. `OS_ENDPOINT`: The access endpoint.
   For example, MinIO's endpoints are usually something like `minio-server:9000`. 
   Don't begin it with the protocol; it's an endpoint, not an url. Also, check if your object store (e.g. MinIO) is using SSL/TLS.
   If not, disable it using `--disable-ssl`.
1. `OS_BUCKET_NAME`: The bucket name you're dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
1. `OS_ACCESS_KEY_ID`: The access key id for the object store.  This is like a user name for logging into the object store.
1. `OS_SECRET_KEY`: The secret key for the object store.  This is like the above user's password.

Keep this information handy.

### Next step: creating a custom deploy manifest for Pachyderm
Once you have Kubernetes deployed, your persistent volume created, and your object store configured, it's time to [create the Pachyderm manifest for deploying to Kubernetes](./deploy_custom/index.md).

## See Also
### Kubernetes variants
- [OpenShift](./openshift.md)
### Object storage variants
- [EMC ECS](./non-cloud-object-stores.md#emc-ecs)
- [MinIO](./non-cloud-object-stores.md#minio)
- [SwiftStack](./non-cloud-object-stores.md#swiftstack)
