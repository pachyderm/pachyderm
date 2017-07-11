# Azure

## Prerequisites

* Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) >= 2.0.1
* Install [jq](https://stedolan.github.io/jq/download/)

## Deploy Kubernetes

The easiest way to deploy a Kubernetes cluster is to use the [official Kubernetes guide](http://kubernetes.io/docs/getting-started-guides/azure/).

## Deploy Pachyderm

To deploy Pachyderm we will need to:

1. Add some storage resources on Azure, 
2. Install the Pachyderm CLI tool, `pachctl`, and
3. Deploy Pachyderm on top of the storage resources.

### Set up the Storage Resources

Pachyderm requires an object store ([Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)) and a [data disk](https://azure.microsoft.com/documentation/articles/virtual-machines-windows-about-disks-vhds/#data-disk) to function correctly.

Here are the parameters required to create these resources:

```sh
# Needs to be globally unique across the entire Azure location
$ RESOURCE_GROUP=[The name of the resource group where the Azure resources will be organized]

$ LOCATION=[The Azure region of your Kubernetes cluster. e.g. "West US2"]

# Needs to be globally unique across the entire Azure location
$ STORAGE_ACCOUNT=[The name of the storage account where your data will be stored]

$ CONTAINER_NAME=[The name of the Azure blob container where your data will be stored]

# Needs to end in a ".vhd" extension
$ STORAGE_NAME=pach-disk.vhd

# We recommend between 1 and 10 GB. This stores PFS metadata. For reference 1GB
# should work for 1000 commits on 1000 files.
$ STORAGE_SIZE=[the size of the data disk volume that you are going to create, in GBs. e.g. "10"]
```

And then run:

```sh
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

```sh
$ az storage account list | jq '.[].name'
$ az storage blob list \
  --container=${CONTAINER_NAME} \
  --account-name=${STORAGE_ACCOUNT} \
  --account-key=${STORAGE_KEY}
```

### Install `pachctl`

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.5

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.5.0/pachctl_1.5.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.6
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

### Deploy Pachyderm

Now we're ready to boot up Pachyderm:

```sh
$ pachctl deploy microsoft ${CONTAINER_NAME} ${STORAGE_ACCOUNT} ${STORAGE_KEY} ${STORAGE_SIZE} --static-etcd-volume=${VOLUME_URI}
```

It may take a few minutes for the pachd nodes to be running because it's pulling containers from Docker Hub. You can see the cluster status by using:

```sh
NAME                        READY     STATUS    RESTARTS   AGE
po/etcd-4197107720-br61m    1/1       Running   0          8m
po/pachd-3548222380-s086m   1/1       Running   2          8m

NAME             CLUSTER-IP     EXTERNAL-IP   PORT(S)                       AGE
svc/etcd         10.111.11.36   <nodes>       2379:32379/TCP                8m
svc/kubernetes   10.96.0.1      <none>        443/TCP                       10m
svc/pachd        10.97.116.5    <nodes>       650:30650/TCP,651:30651/TCP   8m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           8m
deploy/pachd   1         1         1            1           8m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-4197107720    1         1         1         8m
rs/pachd-3548222380   1         1         1         8m
```

Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before etcd was ready so it restarted them.

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.6
pachd               1.4.6
```

