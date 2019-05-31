# On Premises

This document is broken down into the following sections, available at the links below

- [Introduction to on-premises deployments](#introduction) takes you through what you need to know about Kubernetes, persistent volumes, object stores and best practices.  That's this page.
- [Customizing your Pachyderm deployment for on-premises use](./deploy_custom.html) details the various options of the `pachctl deploy custom ...` command for an on-premises deployment.
- [Single-node Pachyderm deployment](./single-node.html) is the document you should read when deploying Pachyderm for personal, low-volume usage.
- [Registries](./docker_registries.html) takes you through on-premises, private Docker registry configuration.
- [Ingress](./configuring_k8s_ingress.html) details the Kubernetes ingress configuration you'd need for using `pachctl` and the dashboard outside of the Kubernetes cluster
- [Non-cloud object stores](./non-cloud-object-stores.html) discusses common configurations for on-premises object stores.

Need information on a particular flavor of Kubernetes or object store?  Check out the [see also](#see-also) section.

Troubleshooting a deployment? Check out [Troubleshooting Deployments](./deploy_troubleshooting.html).

## Introduction

Deploying Pachyderm successfully on-premises requires a few prerequisites and some planning.
Pachyderm is built on [Kubernetes](https://kubernetes.io/).
Before you can deploy Pachyderm, you or your Kubernetes administrator will need to perform the following actions:
1. [Deploy Kubernetes](#deploying-kubernetes) on-premises.
1. [Deploy a Kubernetes persistent volume](#deploying-a-persistent-volume) that Pachyderm will use to store administrative data.
1. [Deploy an on-premises object store](#deploying-an-object-store) using a storage provider like [MinIO](https://min.io/) EMC's ECS or Swift to provide S3-compatible access to your on-premises storage.
1. [Create a Pachyderm manifest](./deploy_custom.html#creating-a-pachyderm-manifest) by running the `pachctl deploy custom` command with appropriate arguments and the `--dry-run` flag to create a Kubernetes manifest for the Pachyderm deployment.
1. [Edit the Pachyderm manifest](#deploy_custom.html#editing-a-pachyderm-manifest) for your particular Kubernetes deployment

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
2. [pachctl](http://docs.pachyderm.io/en/latest/pachctl/pachctl.html)

### Deploying Kubernetes

The Kubernetes docs have instructions for [deploying Kubernetes in a variety of on-premise scenarios](https://kubernetes.io/docs/getting-started-guides/#on-premises-vms).
We recommend following one of these guides to get Kubernetes running on premise.

### Deploying a persistent volume

#### Persistent volumes: how do they work?

A Kubernetes [persistent volume](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) is used by Pachyderm's `etcd` for storage of system metatada. 
In Kubernetes, [persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) are a mechanism for providing storage for consumption by the users of the cluster.
They are provisioned by the cluster administrators.
In a typical enterprise Kubernetes deployment, the administrators have configured persistent volumes that your Pachyderm deployment will consume by means of a [persistent volume claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims) in the [Pachyderm manifest you generate](./deploy_custom.html#creating-a-pachyderm-manifest). 
If your administrators require specification of [selectors](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#selector) or [classes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1) to consume persistent volumes, 
you will need to find out the particulars of your cluster's PV configuration and [edit the Pachyderm manifest](./deploy_custom.html#editing-the-pachyderm-manifest) accordingly.

#### Sizing the pv

You'll need to use a pv with enough space for the metadata associated with the data you plan to store in Pachyderm. 
We're currently developing good rules of thumb for scaling this storage as your Pachyderm deployment grows,
but it looks like 10G of disk space is sufficient for most purposes.

#### Creating the pv

In the case of cloud-based deployments, the `pachctl deploy` command for AWS, GCP and Azure creates persistent volumes for you, when you follow the instructions for those infrastructures.
In the case of on-premises deployments, the kind of PV you provision will be dependent on what kind of storage your Kubernetes administrators have attached to your cluster and configured.
For example, many on-premises deployments use Network File System (NFS) to access to some kind of enterprise storage.
Persistent volumes are provisioned in Kubernetes like all things in Kubernetes: by means of a manifest.
You can learn about creating [volumes](https://kubernetes.io/docs/concepts/storage/volumes/)  and [persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) in the Kubernetes documentation.

#### What you'll need for Pachyderm configuration of pv storage

You'll need the name of the pv and the amount of space you can use, in gigabytes.  We'll refer to those as `PVC_STORAGE_NAME` and `PVC_STORAGE_SIZE` further on.

Keep this information handy.
   
### Deploying an object store

#### Object store: what's it for?
An object store is used by Pachyderm's `pachd` for storing all your data. 
The object store you use must be accessible via a low-latency, high-bandwidth connection like [Gigabit](https://en.wikipedia.org/wiki/Gigabit_Ethernet)  or [10G Ethernet](https://en.wikipedia.org/wiki/10_Gigabit_Ethernet).

For an on-premises deployment, 
it's not advisable to use a cloud-based storage mechanism.
Don't deploy an on-premises Pachyderm cluster against cloud-based object stores such as S3 from [AWS](https://pachyderm.readthedocs.io/en/latest/deployment/amazon_web_services.html), GCS from [Google Cloud Platform](https://pachyderm.readthedocs.io/en/latest/deployment/google_cloud_platform.html), Azure Blob Storage from  [Azure](https://pachyderm.readthedocs.io/en/latest/deployment/azure.html).

#### Object store prerequisites

Object stores are accessible using the S3 protocol, created by Amazon. 
Storage providers like MinIO, EMC's ECS or Swift provide S3-compatible access to enterprise storage for on-premises deployment. 
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
   Don't begin it with the protocol; it's an endpoint, not an url.
1. `OS_BUCKET_NAME`: The bucket name you're dedicating to Pachyderm. Pachyderm will need exclusive access to this bucket.
1. `OS_ACCESS_KEY_ID`: The access key id for the object store.  This is like a user name for logging into the object store.
1. `OS_SECRET_KEY`: The secret key for the object store.  This is like the above user's password.

Keep this information handy.

### Next step: creating a custom deploy manifest for Pachyderm
Once you have Kubernetes deployed, your persistent volume created, and your object store configured, it's time to [create the Pachyderm manifest for deploying to Kubernetes](./deploy_custom.html).

## See Also
### Kubernetes variants
- [OpenShift](./openshift.html)
### Object storage variants
- [ECS](./non-cloud-object-stores.html#ecs)
- [MinIO](./non-cloud-object-stores.html#minio)
- [SwiftStack](./non-cloud-object-stores.html#swiftstack)


