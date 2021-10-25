# Azure

For a quick installation of Pachyderm on Azure, jump to our [Quickstart page](../quickstart/).


!!! Important "Before your start your installation process." 
      - Refer to our generic ["Helm Install"](./helm_install.md) page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/). You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
      - If you are planning to install Pachyderm UI. Read our [Console deployment](../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up on an Ingress.

The following section walks you through deploying a Pachyderm cluster on Microsoft® Azure® Kubernetes
Service environment (AKS). 

In particular, you will:

1. [Install Prerequisites](#1-install-prerequisites)
1. [Deploy Kubernetes](#2-deploy-kubernetes)
1. [Create an Azure Storage Container For Your Data](#3-create-an-azure-storage-container-for-your-data)
1. [Persistent Volumes Creation](#4-persistent-volumes-creation)
1. [Create an Azure Managed PostgreSQL Server Database](#5-create-an-azure-managed-postgresql-server) Database
1. [Deploy Pachyderm](#6-deploy-pachyderm)
1. [Have 'pachctl' and your Cluster Communicate](#7-have-pachctl-and-your-cluster-communicate)
1. [Check That Your Cluster Is Up And Running](#8-check-that-your-cluster-is-up-and-running)

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

You can deploy Kubernetes on Azure by following the official [Azure Kubernetes Service documentation](https://docs.microsoft.com/azure/aks/tutorial-kubernetes-deploy-cluster), [use the quickstart walkthrough](https://docs.microsoft.com/en-us/azure/aks/kubernetes-walkthrough), or follow the steps in this section.

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

1. Create an AKS cluster in the resource group/location:

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


## 3. Create an Azure Storage Container For Your Data

Pachyderm needs an [Azure Storage Container](https://docs.microsoft.com/en-us/azure/databricks/data/data-sources/azure/azure-storage) (Object store) to store your data. 

To access your data, Pachyderm uses a [Storage Account](https://docs.microsoft.com/en-us/azure/storage/common/storage-account-overview) with permissioned access to your desired container. You can either use an existing account or create a new one in your default subscription, then use the JSON key associated with the account and pass it on to Pachyderm.

To create a new storage account, follow the steps below:

!!! Warning
      The storage account name must be unique in the Azure location.

* Set up the following variables:

      * STORAGE_ACCOUNT - The name of the storage account where you store your data.
      * CONTAINER_NAME - The name of the Azure blob container where you store your data.
      * STORAGE_SIZE - The size of the persistent volume to create in GBs. Allocate at least 10 GB.

* Create an Azure storage account:

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

* Verify that your storage account has been successfully created:

    ```shell
    az storage account list
    ```

* Obtain the key for the storage account (`STORAGE_ACCOUNT`) and the resource group to be used to deploy Pachyderm:

    ```shell
    STORAGE_KEY="$(az storage account keys list \
                  --account-name="${STORAGE_ACCOUNT}" \
                  --resource-group="${RESOURCE_GROUP}" \
                  --output=json \
                  | jq '.[0].value' -r
                )"
    ```

!!! Note
    Find the generated key in the **Storage accounts > Access keys**
    section in the [Azure Portal](https://portal.azure.com/) or by running the following command `az storage account keys list --account-name=${STORAGE_ACCOUNT}`.


* Create a new storage container within your storage account:

    ```shell
    az storage container create --name ${CONTAINER_NAME} \
              --account-name ${STORAGE_ACCOUNT} \
              --account-key "${STORAGE_KEY}"
    ```

!!! note "See Also:"
    - [Azure Storage](https://azure.microsoft.com/documentation/articles/storage-introduction/)
## 4. Persistent Volumes Creation

etcd and PostgreSQL (metadata storage) each claim the creation of a pv. 

If you plan to deploy Pachyderm with its default bundled PostgreSQL instance, read the warning below and jump to the [deployment section](#6-deploy-pachyderm):

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your disk choice, you may need to oversize the volume significantly to ensure enough IOPS.

If you plan to deploy a managed PostgreSQL instance (Recommended in production), read the following section.

## 5. Create an Azure Managed PostgreSQL Server Database

By default, Pachyderm runs with a bundled version of PostgreSQL. 
For production environments, it is strongly recommended that you disable the bundled version and use a PostgreSQL Server instance.

This section will provide guidance on the configuration settings you will need to:

- Create an environment to run your Azure PostgreSQL Server databases.
- Create two databases (pachyderm and dex).
- Update your values.yaml to turn off the installation of the bundled postgreSQL and provide your new instance information.

!!! Note
    It is assumed that you are already familiar with PostgreSQL Server, or will be working with an administrator who is.

### Create A PostgreSQL Server Instance¶

!!! Info 
    Find the details of the steps and available parameters to create a PostgreSQL Server instance with Azure Console in Azure Documentation ["Create an Azure Database for PostgreSQL server by using the Azure portal"](https://docs.microsoft.com/en-us/azure/postgresql/quickstart-create-server-database-portal). Alternatively, you can use the cli and run [`az postgres server create`](https://docs.microsoft.com/en-us/cli/azure/postgres/server?view=azure-cli-latest) with your relevant parameters.

In the Azure console, choose the **Azure Database for PostgreSQL servers** service. You will be asked to pick your server type: `Single Server` or `Hyperscale` (for multi-tenant applications), then configure your DB instance as follows.

| SETTING | Recommended value|
|:----------------|:--------------------------------------------------------|
| *subscription*  and *resource group*| Pick an existing [resource group](https://docs.microsoft.com/en-us/azure/cloud-adoption-framework/govern/resource-consistency/resource-access-management#what-is-an-azure-resource-group) or create a new one.|
|*server name*|Name your instance.|
|*location*|Create a database **in the region matching your Pachyderm cluster**.|
|*compute + storage*|The standard instance size (GP_Gen5_4 = Gen5 VMs with 4 cores) should work. Remember that Pachyderm's metadata services require **high IOPS (1500)**. Oversize the disk accordingly |
| *Master username* | Choose your Admin username. ("postgres")|
| *Master password* | Choose your Admin password.|

You are ready to create your instance. 
Once created, go back to your newly created database, and: 

- To open the access to your instance: In the **Connection Security** or your newly created server, *Allow access to Azure services* then set your range of IP addresses in the firewall rules. 

!!! Warning
    Keep the SSL setting `Disabled`.

- Also: In the **Essentials** page of your instance, you will find the full server name and admin username that will be required in your [values.yaml](#update-your-values-yaml).

### Create Your Databases
After the instance is created, those two commands create the databases that pachyderm uses.

```shell
az postgres db create -g <your_group> -s <server_name> -n pachyderm
az postgres db create -g <your_group> -s <server_name> -n dex
```
!!! Note
    Note that the second database must be named `dex`. Read more about [dex on PostgreSQL on Dex's documentation](https://dexidp.io/docs/storage/#postgres).

Pachyderm will use the same user "postgres" to connect to `pachyderm` as well as to `dex`. 
### Update your values.yaml 
Once your databases have been created, add the following fields to your Helm values:


```yaml
global:
  postgresql:
    postgresqlUsername: "username"
    postgresqlPassword: "password"
    # The server name of the instance
    postgresqlDatabase: "INSTANCE_NAME"
    # The postgresql database host to connect to. 
    postgresqlHost: "PostgreSQL Server CNAME"
    # The postgresql database port to connect to. Defaults to postgres server in subchart
    postgresqlPort: "5432"

postgresql:
  # turns off the install of the bundled postgres.
  # If not using the built in Postgres, you must specify a Postgresql
  # database server to connect to in global.postgresql
  enabled: false
```


## 6. Deploy Pachyderm
You have created your data container, given your cluster access to those data, and created a Managed PostgreSQL instance (or chosen to use the default bundled version).

You can now finalize your values.yaml and deploy Pachyderm.

!!! Note 
    - If you have created a Managed PostgreSQL Server instance, you will have to replace the Postgresql section below with the appropriate values defined above.
    - If you plan to deploy Pachyderm with Console, follow these [additional instructions](../console/#deploy-in-the-cloud) and update your values.yaml accordingly.
### Update Your Values.yaml   

After you complete all the sections above, you can deploy Pachyderm
on Azure. If you have previously tried to run Pachyderm locally,
make sure that you are using the right Kubernetes context. 

1. Verify cluster context:

    ```shell
    kubectl config current-context
    ```

    This command should return the name of your Kubernetes cluster that
    runs on Azure.

    If you have a different context displayed, configure `kubectl`
    to use your Azure configuration:

    ```shell
    az aks get-credentials --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME}
    ```

    **System Response:**

    ```shell
    Merged "${CLUSTER_NAME}" as current context in /Users/test-user/.kube/config
    ```

1. Update your values.yaml   

    Update your values.yaml with your container name ([see example of values.yaml here](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/examples/microsoft-values.yaml)) or use our minimal example below.

        
    ```yaml
    deployTarget: MICROSOFT

    pachd:
      storage:
        microsoft:
          # storage container name
          container: "container_name"

          # storage account name
          id: AKIAIOSFODNN7EXAMPLE

          # storage account key
          secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

    postgresql:
      enabled: true
    ```
    Check the [list of all available helm values](../../../../reference/helm_values/) at your disposal in our reference documentation or on [Github](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml).

### Deploy Pachyderm On The Kubernetes Cluster


- Now you can deploy a Pachyderm cluster by running this command:


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
    Refer to our generic [Helm documentation](../helm_install/#install-the-pachyderm-helm-chart) for more information on how to select your chart version.

    Pachyderm pulls containers from DockerHub. It might take some time
    before the `pachd` pods start. You can check the status of the
    deployment by periodically running `kubectl get all`.

    When pachyderm is up and running, get the information about the pods:

    ```shell
    kubectl get pods
    ```

    Once the pods are up, you should see a pod for `pachd` running 
    (alongside etcd, pg-bouncer, postgres, or console, depending on your installation). 
     
    **System Response:**

    ```shell
    NAME                      READY     STATUS    RESTARTS   AGE
    pachd-1971105989-mjn61    1/1       Running   0          54m
    ...
    ```

    **Note:** Sometimes Kubernetes tries to start `pachd` nodes before
    the `etcd` nodes are ready which might result in the `pachd` nodes
    restarting. You can safely ignore those restarts.

- Finally, make sure [`pachtl` talks with your cluster](#7-have-pachctl-and-your-cluster-communicate).
## 7. Have 'pachctl' And Your Cluster Communicate

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
