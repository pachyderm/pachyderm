# Azure

!!! Important "Before your start your installation process." 
      - Refer to our generic ["Helm Install"](./helm_install.md) page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/). You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
      - If you are planning to install Pachyderm UI. Read our [Console deployment](../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up on an Ingress.

The following section walks you through deploying a Pachyderm cluster on Microsoft® Azure® Kubernetes
Service environment (AKS). 

In particular, you will:

1. [Install Prerequisites](#1-install-prerequisites)
1. [Deploy Kubernetes](#2-deploy-kubernetes)

Create an GCS bucket for your data and grant Pachyderm access.
Enable Persistent Volumes Creation
Create An GCP Managed PostgreSQL Instance
Deploy Pachyderm
Finally, you will need to install pachctl to interact with your cluster.
And check that your cluster is up and running

1. [Install Prerequisites](#install-prerequisites)
2. [Deploy Kubernetes](#deploy-kubernetes)
3. [Deploy Pachyderm](#deploy-pachyderm)
4. [Point your CLI `pachctl` to your cluster](#have-pachctl-and-your-cluster-communicate)

## 1. Install Prerequisites

Before your start creating your cluster, install the following
clients on your machine. If not explicitly specified, use the
latest available version of the components listed below.

* [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
* [jq](https://stedolan.github.io/jq/download/)
* [kubectl](https://docs.microsoft.com/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli)
* [pachctl](../../../../getting_started/local_installation#install-pachctl)
 
!!! Note
   This page assumes that you have an [Azure Subsciption](https://docs.microsoft.com/en-us/azure/guides/developer/azure-developer-guide#understanding-accounts-subscriptions-and-billing).

## 2. Deploy Kubernetes

You can deploy Kubernetes on Azure by following the official [Azure Container Service documentation](https://docs.microsoft.com/azure/aks/tutorial-kubernetes-deploy-cluster), [use the quickstart walkthrough](https://docs.microsoft.com/en-us/azure/aks/kubernetes-walkthrough), or
follow the steps in this section.

At a minimum, you will need to specify the parameters below:

|Variable|Description|
|--------|-----------|
|RESOURCE_GROUP|A unique name for the resource group where Pachyderm is deployed. For example, `pach-resource-group`.|
|LOCATION|An Azure availability zone where AKS is available. For example, `centralus`.|
|NODE_SIZE|The size of the Kubernetes virtual machine (VM) instances. To avoid performance issues, Pachyderm recommends that you set this value to at least `Standard_DS4_v2` which gives you 8 CPUs, 28 Gib of Memory, 56 Gib SSD.|
|CLUSTER_NAME|A unique name for the Pachyderm cluster. For example, `pach-aks-cluster`.|

You can choose to follow the guided steps in [Azure Service Portal's Kubernetes Services](https://portal.azure.com/) or use Azure CLI.


1. [Log in](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli) to Azure:

    ```shell
    az login
    ```

    This command opens a browser window. Log in with your Azure credentials.
    Resources can now be provisioned on the Azure subscription linked to your account.
    
    
1. Create an [Azure resource group](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/manage-resource-groups-portal#what-is-a-resource-group) or retrieve an existing group.

    ```shell
    az group create --name ${RESOURCE_GROUP} --location ${LOCATION}
    ```

    **Example:**

    ```shell
    az group create --name test-group --location centralus
    ```

    **System Response:**

    ```json
    {
      "id": "/subscriptions/6c9f2e1e-0eba-4421-b4cc-172f959ee110/resourceGroups/pach-resource-group",
      "location": "centralus",
      "managedBy": null,
      "name": "pach-resource-group",
      "properties": {
        "provisioningState": "Succeeded"
      },
      "tags": null,
      "type": null
    }
    ```

1. Create an AKS cluster in the same :

   For more configuration options: Find the list of [all available flags of the `az aks create` command](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_create).

    ```shell
    az aks create --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME} --node-vm-size ${NODE_SIZE} --node-count <node_pool_count> --location ${LOCATION}
    ```

    **Example:**

    ```shell
    az aks create --resource-group test-group --name test-cluster --generate-ssh-keys --node-vm-size Standard_DS4_v2 --location centralus
    ```

1. Confirm the version of the Kubernetes server by running  `kubectl version`.

!!! note "See Also:"
    - [Azure Virtual Machine sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/sizes-general)

Once your Kubernetes cluster is up, and your infrastructure configured, you are ready to prepare for the installation of Pachyderm. Some of the steps below will require you to keep updating the values.yaml started during the setup of the recommended infrastructure:


## 3. Create an Azure Blob Storage Bucket
Create a Storage Account
Create a Blob Container
Upload a Blob
### Create an ABS object store bucket for your data
Pachyderm needs a [ABS bucket](https://docs.microsoft.com/en-us/azure/databricks/data/data-sources/azure/azure-storage) (Object store) to store your data. 
You can create a blob container storage account using the [Azure portal](https://docs.microsoft.com/en-us/azure/container-instances/container-instances-quickstart-portal)You can create the bucket by running the following commands:

Warning

The GCS bucket name must be globally unique across the entire GCP region.

## 4. Persistent Volumes Creation

Pachyderm requires you to deploy an object store and two persistent
volumes in your cloud environment to function correctly. For best
results, you need to use faster disk drives, such as *Premium SSD
Managed Disks* that are available with the Azure Premium Storage offering.

You need to specify the following parameters when you create storage
resources:

|Variable|Description|
|--------|-----------|
|STORAGE_ACCOUNT|The name of the storage account where you store your data, unique in the Azure location.|
|CONTAINER_NAME|The name of the Azure blob container where you store your data.|
|STORAGE_SIZE|The size of the persistent volume to create in GBs. Allocate at least 10 GB.|

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your disk choice, you may need to oversize the volume significantly to ensure enough IOPS.

To create these resources, follow these steps:

1. Clone the [Pachyderm GitHub repo](https://github.com/pachyderm/pachyderm).
1. Change the directory to the root directory of the `pachyderm` repository.
1. Create an Azure storage account:

    ```shell
    az storage account create \
      --resource-group="${RESOURCE_GROUP}" \
      --location="${LOCATION}" \
      --sku=Premium_LRS \
      --name="${STORAGE_ACCOUNT}" \
      --kind=BlockBlobStorage
    ```
    **System response:**

    ```json
    {
      "accessTier": null,
      "creationTime": "2019-06-20T16:05:55.616832+00:00",
      "customDomain": null,
      "enableAzureFilesAadIntegration": null,
      "enableHttpsTrafficOnly": false,
      "encryption": {
        "keySource": "Microsoft.Storage",
        "keyVaultProperties": null,
        "services": {
          "blob": {
            "enabled": true,
      ...
    ```

    Make sure that you set Stock Keeping Unit (SKU) to `Premium_LRS`
    and the `kind` parameter is set to `BlockBlobStorage`. This
    configuration results in a storage that uses SSDs rather than
    standard Hard Disk Drives (HDD).
    If you set this parameter to an HDD-based storage option, your Pachyderm
    cluster will be too slow and might malfunction.

1. Verify that your storage account has been successfully created:

    ```shell
    az storage account list
    ```

1. Obtain the key for the storage account (`STORAGE_ACCOUNT`) and the resource group to be used to deploy Pachyderm:

    ```shell
    STORAGE_KEY="$(az storage account keys list \
                  --account-name="${STORAGE_ACCOUNT}" \
                  --resource-group="${RESOURCE_GROUP}" \
                  --output=json \
                  | jq '.[0].value' -r
                )"
    ```

1. Find the generated key in the **Storage accounts > Access keys**
   section in the Azure Portal or by running the following command:

    ```shell
    az storage account keys list --account-name=${STORAGE_ACCOUNT}
    ```

    **System Response:**

    ```json
    [
      {
        "keyName": "key1",
        "permissions": "Full",
        "value": ""
      }
    ]
    ```

1. Create a new storage container within your storage account:

    ```shell
    az storage container create --name ${CONTAINER_NAME} \
              --account-name ${STORAGE_ACCOUNT} \
              --account-key "${STORAGE_KEY}"
    ```

!!! note "See Also:"
    - [Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)

## 5. Create a Azure Managed PostgreSQL Server Database
## 6. Deploy Pachyderm

After you complete all the sections above, you can deploy Pachyderm
on Azure. If you have previously tried to run Pachyderm locally,
make sure that you are using the right Kubernetes context. Otherwise,
you might accidentally deploy your cluster on Minikube.

1. Verify cluster context:

    ```shell
    kubectl config current-context
    ```

    This command should return the name of your Kubernetes cluster that
    runs on Azure.

    If you have a different contents displayed, configure `kubectl`
    to use your Azure configuration:

    ```shell
    az aks get-credentials --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME}
    ```

    **System Response:**

    ```shell
    Merged "${CLUSTER_NAME}" as current context in /Users/test-user/.kube/config
    ```

1. Create your values.yaml   

    Update your values.yaml with your container name ([see example of values.yaml here](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/examples/microsoft-values.yaml)) or use our minimal example below.

    ```yaml
    deployTarget: MICROSOFT

    pachd:
      storage:
        microsoft:
          container: "foo"
          id: "bar"
          secret: "baz"
    ```

    **Load Balancer Setup**
    If you would like to expose your pachd instance to the internet via load balancer, add the following config under `pachd` to your `values.yaml`

    **NOTE:** It is strongly recommended to configure SSL when exposing Pachyderm publicly

    ```yaml
    pachd:
      service:
        type: LoadBalancer
    ```
    !!! Note
        Check the [list of all available helm values](../../../reference/helm_values/) at your disposal in our reference documentation.

1. Run the following command:

    ```shell
    $ helm repo add pach https://helm.pachyderm.com
    $ helm repo update
    $ helm install pachd -f my_values.yaml pach/pachyderm --version <version-of-the-chart>
    ```

    **System Response:**

    ```shell
    NAME: pachd
    LAST DEPLOYED: Mon Jul 12 18:28:59 2021
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    ```

    Because Pachyderm pulls containers from DockerHub, it might take some time
    before the `pachd` pods start. You can check the status of the
    deployment by periodically running `kubectl get all`.

1. When pachyderm is up and running, get the information about the pods:

    ```shell
    kubectl get pods
    ```

    Once the pods are up, you should see a pod for `pachd` running 
    (alongside etcd, pg-bouncer or postgres, console, depending on your installation). 
     
    **System Response:**

    ```shell
    NAME                      READY     STATUS    RESTARTS   AGE
    pachd-1971105989-mjn61    1/1       Running   0          54m
    ...
    ```

    **Note:** Sometimes Kubernetes tries to start `pachd` nodes before
    the `etcd` nodes are ready which might result in the `pachd` nodes
    restarting. You can safely ignore those restarts.

## 7. Have 'pachctl' and your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you are exposing your cluster publicly, retrieve the external IP address of your TCP load balancer or your domain name and:

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      $ echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
      ```

  1. Check that your are using the right context: 

      ```shell
      $ pachctl config get active-context`
      ```

      Your cluster context name should show up.

If you're not exposing `pachd` publicly, you can run:

```shell
# Background this process because it blocks.
$ pachctl port-forward
``` 

## 8. Check That Your Cluster Is Up And Running

```shell
$ pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```
