# Custom Object Stores

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

```shell
# Name this whatever you want, we chose pach-disk as a default
$ STORAGE_NAME=pach-disk

# For a demo you should only need 10 GB. This stores PFS metadata. For reference, 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the volume that you are going to create, in GBs. e.g. "10"]

# Create the disk.
gcloud compute disks create --size=${STORAGE_SIZE}GB ${STORAGE_NAME}
```

Then we can deploy Pachyderm:

```shell
pachctl deploy custom --persistent-disk google --object-store s3 ${STORAGE_NAME} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint> --static-etcd-volume=${STORAGE_NAME}
```

## AWS + Custom Object Store

Additional prerequisites:

- [AWS CLI](https://aws.amazon.com/cli/) - have it installed and have your [AWS credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html) configured.

First, we need to create a persistent disk for Pachyderm's metadata:

```shell
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

```shell
pachctl deploy custom --persistent-disk aws --object-store s3 ${STORAGE_NAME} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint> --static-etcd-volume=${STORAGE_NAME}
```

## Azure + Custom Object Store

Additional prerequisites:

- Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) >= 2.0.1
- Install [jq](https://stedolan.github.io/jq/download/)
- Clone github.com/pachyderm/pachyderm and work from the root of that project.

First, we need to create a persistent disk for Pachyderm's metadata. To do this, start by declaring some environmental variables:

```shell
# Needs to be globally unique across the entire Azure location
$ RESOURCE_GROUP=[The name of the resource group where the Azure resources will be organized]

$ LOCATION=[The Azure region of your Kubernetes cluster. e.g. "West US2"]

# Needs to be globally unique across the entire Azure location
$ STORAGE_ACCOUNT=[The name of the storage account where your data will be stored]

# Needs to end in a ".vhd" extension
$ STORAGE_NAME=pach-disk.vhd

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the data disk volume that you are going to create, in GBs. e.g. "10"]
```

And then run:

```shell
# Create a resource group
$ az group create --name=${RESOURCE_GROUP} --location=${LOCATION}

# Create azure storage account
az storage account create \
  --resource-group="${RESOURCE_GROUP}" \
  --location="${LOCATION}" \
  --sku=Standard_LRS \
  --name="${STORAGE_ACCOUNT}" \
  --kind=Storage

# Build microsoft tool for creating Azure VMs from an image
$ STORAGE_KEY="$(az storage account keys list \
                 --account-name="${STORAGE_ACCOUNT}" \
                 --resource-group="${RESOURCE_GROUP}" \
                 --output=json \
                 | jq .[0].value -r
              )"
$ make docker-build-microsoft-vhd 
$ VOLUME_URI="$(docker run -it microsoft_vhd \
                "${STORAGE_ACCOUNT}" \
                "${STORAGE_KEY}" \
                "${CONTAINER_NAME}" \
                "${STORAGE_NAME}" \
                "${STORAGE_SIZE}G"
             )"
```

To check that everything has been setup correctly, try:

```shell
$ az storage account list | jq '.[].name'
```

The we can deploy Pachyderm:

```shell
pachctl deploy custom --persistent-disk azure --object-store s3 ${VOLUME_URI} ${STORAGE_SIZE} <object store bucket> <object store id> <object store secret> <object store endpoint> --static-etcd-volume=${VOLUME_URI}
```
