# Deploying Pachyderm - Custom Object Stores

In other sections of this guide was have demonstrated how to deploy Pachyderm in a single cloud using that cloud's object store offering.  However, Pachyderm can be backed by any object store, and you are not restricted to the object store service provided by the cloud in which you are deploying.

As long as you are running an object store that has an S3 compatible API, you can easily deploy Pachyderm in a way that will allow you to back Pachyderm by that object store.  For example, we have seen Pachyderm be backed by [Minio](https://minio.io/), [GlusterFS](https://www.gluster.org/), [Ceph](http://ceph.com/), and more.  

To deploy Pachyderm with your choice of object store in Google, Azure, or AWS, see the below guides.  To deploy Pachyderm on premise with a custom object store, see the [on premise docs](http://pachyderm.readthedocs.io/en/stable/deployment/on_premises.html).

## Common Prerequisites

1. A working Kubernetes cluster and `kubectl`.
2. An account on or running instance of an object store with an S3 compatible API.  You should be able to get an ID, secret, bucket name, and endpoint that point to this object store.

## Google + Custom Object Store

Additional prerequisites:

- [Google Cloud SDK](https://cloud.google.com/sdk/) >= 124.0.0 - If this is the first time you use the SDK, make sure to follow the [quick start guide](https://cloud.google.com/sdk/docs/quickstarts).

First, we need to create a persistent disk for Pachyderm's metadata:

```sh
# Name this whatever you want, we chose pach-disk as a default
$ STORAGE_NAME=pach-disk

# For a demo you should only need 10 GB. This stores PFS metadata. For reference, 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the volume that you are going to create, in GBs. e.g. "10"]

# Create the disk.
gcloud compute disks create --size=${STORAGE_SIZE}GB ${STORAGE_NAME}
```

Then we can deploy Pachyderm:

```sh
pachctl deploy custom --persistent-disk google --object-store s3 ${STORAGE_NAME} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint>
```

## AWS + Custom Object Store

Additional prerequisites:

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.

First, we need to create a persistent disk for Pachyderm's metadata:

```sh
# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the EBS volume that you are going to create, in GBs. e.g. "10"]

$ AWS_REGION=[the AWS region of your Kubernetes cluster. e.g. "us-west-2" (not us-west-2a)]

$ AWS_AVAILABILITY_ZONE=[the AWS availability zone of your Kubernetes cluster. e.g. "us-west-2a"]

# Create the volume.
$ aws ec2 create-volume --size ${STORAGE_SIZE} --region ${AWS_REGION} --availability-zone ${AWS_AVAILABILITY_ZONE} --volume-type gp2

# Store the volume ID.
$ aws ec2 describe-volumes
$ STORAGE_NAME=[volume id]
```

The we can deploy Pachyderm:

```sh
pachctl deploy custom --persistent-disk aws --object-store s3 ${STORAGE_NAME} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint>
```

## Azure + Custom Object Store

Additional prerequisites:

- Install [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) >= 0.10.6
- Install [jq](https://stedolan.github.io/jq/download/)
- Clone github.com/pachyderm/pachyderm and work from the root of that project.

First, we need to create a persistent disk for Pachyderm's metadata. To do this, start by declaring some environmental variables:

```sh
# Needs to be globally unique across the entire Azure location.
$ AZURE_RESOURCE_GROUP=[The name of the resource group where the Azure resources will be organized]

$ AZURE_LOCATION=[The Azure region of your Kubernetes cluster. e.g. "West US2"]

# Needs to end in a ".vhd" extension
$ STORAGE_NAME=pach-disk.vhd

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the data disk volume that you are going to create, in GBs. e.g. "10"]
```

And then run:

```sh
$ azure group create --name ${AZURE_RESOURCE_GROUP} --location ${AZURE_LOCATION}
$ azure storage account create ${AZURE_STORAGE_NAME} --location ${AZURE_LOCATION} --resource-group ${AZURE_RESOURCE_GROUP} --sku-name LRS --kind Storage

# Retrieve the Azure Storage Account Key
$ AZURE_STORAGE_KEY=`azure storage account keys list ${AZURE_STORAGE_NAME} --resource-group ${AZURE_RESOURCE_GROUP} --json | jq .[0].value -r`

# Build the microsoft_vhd container.
$ make docker-build-microsoft-vhd

# Create an empty data disk in the "disks" container
$ STORAGE_VOLUME_URI=`docker run -it microsoft_vhd ${AZURE_STORAGE_NAME} ${AZURE_STORAGE_KEY} "disks" ${STORAGE_NAME} ${STORAGE_SIZE}G`
```

To check that everything has been setup correctly, try:

```sh
$ azure storage account list
# should see a number of storage accounts, including the one specified with ${AZURE_STORAGE_NAME}

$ azure storage blob list --account-name ${AZURE_STORAGE_NAME} --account-key ${_AZURE_STORAGE_KEY}
# should see a disk with the name ${STORAGE_NAME}
```

The we can deploy Pachyderm:

```sh
pachctl deploy custom --persistent-disk azure --object-store s3 ${STORAGE_VOLUME_URI} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint>
```
