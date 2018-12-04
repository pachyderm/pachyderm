# Azure

To deploy Pachyderm to Azure, you need to:

1. [Install Prerequisites](#prerequisites)
2. [Deploy Kubernetes](#deploy-kubernetes)
3. [Deploy Pachyderm on Kubernetes](#deploy-pachyderm)

## Prerequisites

Install the following prerequisites:

* [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) >= 2.0.1
* [jq](https://stedolan.github.io/jq/download/)
* [kubectl](https://docs.microsoft.com/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli)
* [pachctl](#install-pachctl)

## Deploy Kubernetes

The easiest way to deploy a Kubernetes cluster is through the [Azure Container Service (AKS)](https://docs.microsoft.com/azure/aks/tutorial-kubernetes-deploy-cluster). To create a new AKS Kubernetes cluster using the Azure CLI `az`, run:

```sh
$ RESOURCE_GROUP=<a unique name for the resource group where Pachyderm will be deployed, e.g. "pach-resource-group">

$ LOCATION=<a Azure availability zone where AKS is available, e.g, "Central US">

$ NODE_SIZE=<size for the k8s instances, we recommend at least "Standard_DS4_v2">

$ CLUSTER_NAME=<unique name for the cluster, e.g., "pach-aks-cluster">

# Create the Azure resource group.
$ az group create --name=${RESOURCE_GROUP} --location=${LOCATION}

# Create the AKS cluster.
$ az aks create --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME} --generate-ssh-keys --node-vm-size ${NODE_SIZE}
```

Once Kubernetes is up and running you should be able to confirm the version of the Kubernetes server via:

```sh
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"9", GitVersion:"v1.9.3", GitCommit:"d2835416544f298c919e2ead3be3d0864b52323b", GitTreeState:"clean", BuildDate:"2018-02-07T12:22:21Z", GoVersion:"go1.9.2", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.9", GitCommit:"19fe91923d584c30bd6db5c5a21e9f0d5f742de8", GitTreeState:"clean", BuildDate:"2017-10-19T16:55:06Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

**Note** - Azure AKS is still a relatively new managed service. As such, we have had some issues consistently deploying AKS clusters in certain availability zones. If you get timeouts or issues when provisioning an AKS cluster, we recommend trying in a fresh resource group and possibly trying a different zone.

## Deploy Pachyderm

To deploy Pachyderm we will need to:

1. Add some storage resources on Azure,
2. Install the Pachyderm CLI tool, `pachctl`, and
3. Deploy Pachyderm on top of the storage resources.

### Set up the Storage Resources

Pachyderm requires an object store and persistent volume ([Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)) to function correctly. To create these resources, you need to clone the [Pachyderm GitHub repo](https://github.com/pachyderm/pachyderm) and then run the following from the root of that repo:

```sh
$ STORAGE_ACCOUNT=<The name of the storage account where your data will be stored, unique in the Azure location>

$ CONTAINER_NAME=<The name of the Azure blob container where your data will be stored>

$ STORAGE_SIZE=<the size of the persistent volume that you are going to create in GBs, we recommend at least "10">

# Create an Azure storage account
az storage account create \
  --resource-group="${RESOURCE_GROUP}" \
  --location="${LOCATION}" \
  --sku=Standard_LRS \
  --name="${STORAGE_ACCOUNT}" \
  --kind=Storage

# Build a microsoft tool for creating Azure VMs from an image. Necessary to create the blank PV.
$ STORAGE_KEY="$(az storage account keys list \
                 --account-name="${STORAGE_ACCOUNT}" \
                 --resource-group="${RESOURCE_GROUP}" \
                 --output=json \
                 | jq '.[0].value' -r
              )"
```

### Install `pachctl`

`pachctl` is a command-line utility used for interacting with a Pachyderm cluster.

```shell
# For OSX:
$ brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@1.8

# For Linux (64 bit) or Window 10+ on WSL:
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.8.0/pachctl_1.8.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
```

You can try running `pachctl version` to check that this worked correctly:

```sh
$ pachctl version --client-only
COMPONENT           VERSION
pachctl             1.7.0
```

### Deploy Pachyderm

Now we're ready to deploy Pachyderm:

```sh
$ pachctl deploy microsoft ${CONTAINER_NAME} ${STORAGE_ACCOUNT} ${STORAGE_KEY} ${STORAGE_SIZE} --dynamic-etcd-nodes 1
```

It may take a few minutes for the pachd pods to be running because it's pulling containers from Docker Hub. When Pachyderm is up and running, you should see something similar to the following state:

```sh
$ kubectl get pods
NAME                      READY     STATUS    RESTARTS   AGE
dash-482120938-vdlg9      2/2       Running   0          54m
etcd-0                    1/1       Running   0          54m
pachd-1971105989-mjn61    1/1       Running   0          54m
```

**Note**: If you see a few restarts on the pachd nodes, that's totally ok. That simply means that Kubernetes tried to bring up those containers before etcd was ready so it restarted them.

Finally, assuming you want to connect to the cluster from your local machine (i.e., your laptop), we need to set up forward a port so that `pachctl` can talk to the cluster:

```sh
# Forward the ports. We background this process because it blocks.
$ pachctl port-forward &
```

And you're done! You can test to make sure the cluster is working by trying `pachctl version` or even by creating a new repo.

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.0
pachd               1.7.0
```
