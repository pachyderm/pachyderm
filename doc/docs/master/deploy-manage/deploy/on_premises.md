# On Premises

This page walks you through the fundamentals of what you need to know about Kubernetes, persistent volumes, and object stores to deploy Pachyderm on-premises.

!!! Note "Check Also"
    - [Ingress](../ingress/) details the Kubernetes ingress configuration you would need for using the Console.
    - Troubleshooting a deployment? Check out [Troubleshooting Deployments](../../troubleshooting/deploy_troubleshooting.md).

## Introduction

Deploying Pachyderm successfully on-premises requires a few prerequisites.
Pachyderm is built on [Kubernetes](https://kubernetes.io/).
Before you can deploy Pachyderm, you will need to perform the following actions:

1. [Deploy Kubernetes](#deploying-kubernetes) on-premises.
1. [Deploy two Kubernetes persistent volumes](#storage-classes ) that Pachyderm will use to store its metadata.
1. [Deploy an on-premises object store](#deploying-an-object-store) using a storage provider like [MinIO](https://min.io), [EMC's ECS](https://www.dellemc.com/storage/ecs/index.htm), or [SwiftStack](https://www.swiftstack.com/) to provide S3-compatible access to your data storage.
1. [Deploy Pachyderm using Helm](./helm_install.md) by running the `helm install` command with appropriate values configured to create a Kubernetes manifest for the Pachyderm deployment.

## Prerequisites
Before you start, you will need the following clients installed: 

1. [kubectl](https://kubernetes.io/docs/user-guide/prereqs/)
2. [pachctl](../../../getting_started/local_installation/#install-pachctl)

## Setting Up To Deploy On-Premises

### Deploying Kubernetes
The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/getting-started-guides/#on-premises-vms).
We recommend following one of these guides to get Kubernetes running.

### Storage Classes 
Once you deploy Kubernetes, you will also need to configure [storage classes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1) to consume persistent volumes for `etcd` and `postgresql`. 

!!! Warning
    The database and metadata service (Persistent disks) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your storage provider, you may need to oversize the volume significantly to ensure enough IOPS.

Once you have determined the name of the storage classes you are going to use and the sizes, you can add them to your helm values file, specifically:

```yaml
etcd:
  storageClass: MyStorageClass
  size: 10Gi

postgresql:
  persistence:
    storageClass: MyStorageClass
    size: 10Gi
```
   
### Deploying An Object Store
An object store is used by Pachyderm's `pachd` for storing all your data. 
The object store you use must be accessible via a low-latency, high-bandwidth connection.

!!! Note
    For an on-premises deployment, 
    it is not advisable to use a cloud-based storage mechanism.
    Do not deploy an on-premises Pachyderm cluster against cloud-based object stores (such as S3, GCS, Azure Blob Storage). 

    You will, however, **access your Object Store using the S3 protocol**. 

Storage providers like [MinIO](https://min.io), [EMC's ECS](https://www.dellemc.com/storage/ecs/index.htm), or [SwiftStack](https://www.swiftstack.com/) provide S3-compatible access to enterprise storage for on-premises deployment. 

Refer to our [Custom Object Stores](../custom_object_stores/) page for more details on how to set up non-cloud object stores.

#### Sizing And Configuring The Object Store
Start with a large multiple of your current data set size.

You will need four items to configure the object store.
We are prefixing each item with how we will refer to it in the helm values file.

1. `endpoint`: The access endpoint.
   For example, MinIO's endpoints are usually something like `minio-server:9000`. 

    Do not begin it with the protocol; it is an endpoint, not an url. Also, check if your object store (e.g. MinIO) is using SSL/TLS.
    If not, disable it using `secure: false`.

2. `bucket`: The bucket name you are dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
3. `id`: The access key id for the object store.  
4. `secret`: The secret key for the object store.  

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

### Next Step: Create A Custom Helm Values File For Pachyderm
Once you have Kubernetes deployed, your storage classes setup, and your object store configured, it is time to [create the Pachyderm helm values for deploying to Kubernetes](./helm_install.md).
