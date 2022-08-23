# On Premises

This page walks you through the fundamentals of what you need to know about Kubernetes, persistent volumes, and object stores to deploy Pachyderm on-premises.

!!! Note "Check Also"
    - Read our [infrastructure recommendations](../ingress/). You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
    - If you are planning to install Pachyderm UI. Read our [Console deployment](../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up of an Ingress.
    - Troubleshooting a deployment? Check out [Troubleshooting Deployments](../../troubleshooting/deploy-troubleshooting.md).

!!! Attention 
    We are now shipping Pachyderm with an **optional embedded proxy** 
    allowing your cluster to expose one single port externally. This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a Proxy, check out our new recommended architecture and [deployment instructions](../deploy-w-proxy/). 
## Introduction

Deploying Pachyderm successfully on-premises requires a few prerequisites.
Pachyderm is built on [Kubernetes](https://kubernetes.io/){target=_blank}.
Before you can deploy Pachyderm, you will need to perform the following actions:

1. [Deploy Kubernetes](#deploying-kubernetes) on-premises.
1. [Deploy two Kubernetes persistent volumes](#storage-classes ) that Pachyderm will use to store its metadata.
1. [Deploy an on-premises object store](#deploying-an-object-store) using a storage provider like [MinIO](https://min.io){target=_blank}, [EMC's ECS](https://www.delltechnologies.com/en-us/storage/ecs/index.htm){target=_blank}, or [SwiftStack](https://www.swiftstack.com/){target=_blank} to provide S3-compatible access to your data storage.
1. Finally, [Deploy Pachyderm using Helm](./helm-install.md) by running the `helm install` command with the appropriate values configured in your values.yaml. We recommend reading these generic deployment steps if you are unfamiliar with Helm.

## Prerequisites
Before you start, you will need the following clients installed: 

1. [kubectl](https://kubernetes.io/docs/tasks/tools/){target=_blank}
2. [pachctl](../../../getting-started/local-installation/#install-pachctl)

## Setting Up To Deploy On-Premises

### Deploying Kubernetes
The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/setup/){target=_blank}.
We recommend following one of these guides to get Kubernetes running.

!!! Attention
    Pachyderm recommends running your cluster on Kubernetes 1.19.0 and above.
### Storage Classes 
Once you deploy Kubernetes, you will also need to configure [storage classes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1){target=_blank} to consume persistent volumes for `etcd` and `postgresql`. 

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

Storage providers like [MinIO](https://min.io){target=_blank} (the most common and officially supported option), [EMC's ECS](https://www.delltechnologies.com/en-us/storage/ecs/index.htm){target=_blank}, [Ceph](https://ceph.io/en/){target=_blank}, or [SwiftStack](https://www.swiftstack.com/){target=_blank} provide S3-compatible access to enterprise storage for on-premises deployment. 

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

### Next Step: Proceed to your Helm installation
Once you have Kubernetes deployed, your storage classes setup, and your object store configured, follow those steps [to Helm install Pachyderm on your cluster](./helm-install.md).
