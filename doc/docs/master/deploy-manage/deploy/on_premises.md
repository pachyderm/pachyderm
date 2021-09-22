# On Premises

This document is broken down into the following sections, available at the links below

- [Introduction to on-premises deployments](#introduction) takes you through what you need to know about Kubernetes, persistent volumes, object stores and best practices.  That's this page.
- [Single-node Pachyderm deployment](./single-node.md) is the document you should read when deploying Pachyderm for personal, low-volume usage.
- [Registries](./docker_registries.md) takes you through on-premises, private Docker registry configuration.
- [Ingress](https://docs.pachyderm.com/latest/deploy-manage/deploy/ingress/) details the Kubernetes ingress configuration you'd need for using `pachctl` and the dashboard outside of the Kubernetes cluster
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
1. [Deploy Pachyderm using Helm](./helm_install.md) by running the `helm install` command with appropriate values configured to create a Kubernetes manifest for the Pachyderm deployment.

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
2. [pachctl](../../../getting_started/local_installation/#install-pachctl)

## Setting up to deploy on-premises

### Deploying Kubernetes

The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/getting-started-guides/#on-premises-vms).
We recommend following one of these guides to get Kubernetes running on premise.

#### StorageClasses 

Once you deploy Kubernetes, you'll also need to configure [storage classes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1) to consume persistent volumes for `etcd` and `postgresql`.

!!! Warning
    The database and metadata service (Persistent disks) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your storage provider, you may need to oversize the volume significantly to ensure enough IOPS.

Once you've determined the name of the storage class you're going to use and the sizes, you can add them to your helm values file, specifically:

```yaml

etcd:
  storageClass: MyStorageClass
  size: 10Gi

postgresql:
  persistence:
    storageClass: MyStorageClass
    size: 10Gi
```

   
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
We're prefixing each item with how we'll refer to it in the helm values file.

1. `endpoint`: The access endpoint.
   For example, MinIO's endpoints are usually something like `minio-server:9000`. 
   Don't begin it with the protocol; it's an endpoint, not an url. Also, check if your object store (e.g. MinIO) is using SSL/TLS.
   If not, disable it using `secure: false`.
2. `bucket`: The bucket name you're dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
3. `id`: The access key id for the object store.  This is like a user name for logging into the object store.
4. `secret`: The secret key for the object store.  This is like the above user's password.

```yaml
pachd:
  storage:
    backend: minio
    minio:
      bucket: ""
      endpoint: ""
      id: ""
      secret: ""
      secure: ""
```

### Next step: creating a custom helm values file for Pachyderm
Once you have Kubernetes deployed, your storage class setup, and your object store configured, it's time to [create the Pachyderm helm values for deploying to Kubernetes](./helm_install.md).

## See Also
### Kubernetes variants
- [OpenShift](./openshift.md)
### Object storage variants
- [EMC ECS](./non-cloud-object-stores.md#emc-ecs)
- [MinIO](./non-cloud-object-stores.md#minio)
- [SwiftStack](./non-cloud-object-stores.md#swiftstack)
