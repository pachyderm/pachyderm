# Deploying Pachyderm - Azure

## Microsoft Azure

### Prerequisites

* Install [Azure CLI](https://azure.microsoft.com/documentation/articles/xplat-cli-install/) >= 0.10.6
* Install [jq](https://stedolan.github.io/jq/download/)

### Deploy Kubernetes

The easiest way to deploy a Kubernetes cluster is to use the [official Kubernetes guide](http://kubernetes.io/docs/getting-started-guides/azure/).

### Deploy Pachyderm

To deploy Pachyderm we will need to:

1. Add some storage resources on Azure, 
2. Install the Pachyderm CLI tool, `pachctl`, and
3. Deploy Pachyderm on top of the storage resources.

#### Set up the Storage Resources

Pachyderm requires an object store ([Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)) and a [data disk](https://azure.microsoft.com/documentation/articles/virtual-machines-windows-about-disks-vhds/#data-disk) to function correctly.

Here are the parameters required to create these resources:

```sh
# Needs to be globally unique across the entire Azure location
$ AZURE_RESOURCE_GROUP=[The name of the resource group where the Azure resources will be organized]

$ AZURE_LOCATION=[The Azure region of your Kubernetes cluster. e.g. "West US2"]

# Needs to be globally unique across the entire Azure location
$ AZURE_STORAGE_NAME=[The name of the storage account where your data will be stored]

$ CONTAINER_NAME=[The name of the Azure blob container where your data will be stored]

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

#### Install `pachctl`

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

```sh
# For OSX:
$ brew tap pachyderm/tap && brew install pachctl

# For Linux (64 bit):
$ curl -o /tmp/pachctl.deb -L https://pachyderm.io/pachctl.deb && dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly, but Pachyderm itself isn't deployed yet so you won't get a `pachd` version.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.0
pachd               (version unknown) : error connecting to pachd server at address (0.0.0.0:30650): context deadline exceeded.
```

#### Deploy Pachyderm

Now we're ready to boot up Pachyderm:

```sh
$ pachctl deploy microsoft ${CONTAINER_NAME} ${AZURE_STORAGE_NAME} ${AZURE_STORAGE_KEY} ${STORAGE_VOLUME_URI} ${STORAGE_SIZE}
```

It may take a few minutes for the pachd nodes to be running because it's pulling containers from Docker Hub. You can see the cluster status by using:

```sh
$ kubectl get all
NAME               READY     STATUS    RESTARTS   AGE
po/etcd-xzc0d      1/1       Running   0          55s
po/pachd-6m6wm     1/1       Running   0          55s
po/rethink-388b3   1/1       Running   0          55s

NAME         DESIRED   CURRENT   READY     AGE
rc/etcd      1         1         1         55s
rc/pachd     1         1         1         55s
rc/rethink   1         1         1         55s

NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)                                          AGE
svc/etcd         10.0.0.92    <none>        2379/TCP,2380/TCP                                55s
svc/kubernetes   10.0.0.1     <none>        443/TCP                                          9m
svc/pachd        10.0.0.61    <nodes>       650:30650/TCP,651:30651/TCP                      55s
svc/rethink      10.0.0.87    <nodes>       8080:32080/TCP,28015:32081/TCP,29015:32085/TCP   55s

NAME              DESIRED   SUCCESSFUL   AGE
jobs/pachd-init   1         1            55s

```
Note: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before Rethink was ready so it restarted them.

Finally, we need to set up forward a port so that pachctl can talk to the cluster.

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl portforward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.4.0
pachd               1.4.0
```

