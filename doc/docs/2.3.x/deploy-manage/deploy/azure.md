# Azure

For a quick test installation of Pachyderm on Azure (suitable for development), jump to our [Quickstart page](../quickstart/).


!!! Important "Before your start your installation process." 
      - Refer to our generic ["Helm Install"](./helm-install.md) page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/). You will find instructions on how to set up an ingress controller, a load balancer, or connect an Identity Provider for access control. 
      - Pachyderm comes with a [web UI (Console)](../console) for visualizing running pipelines and exploring your data. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires, at a minimum, the set up of an Ingress.
    

!!! Attention 
    We are now shipping Pachyderm with an **embedded proxy** 
    allowing your cluster to expose one single port externally. This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a Proxy, check out our new recommended architecture and [deployment instructions](../deploy-w-proxy/) as they alter the instructions below.

The following section walks you through deploying a Pachyderm cluster on Microsoft® Azure® Kubernetes
Service environment (AKS). 

In particular, you will:

- [Azure](#azure)
  - [1. Install Prerequisites](#1-install-prerequisites)
  - [2. Deploy Kubernetes](#2-deploy-kubernetes)
  - [3. Create an Azure Storage Container For Your Data](#3-create-an-azure-storage-container-for-your-data)
  - [4. Persistent Volumes Creation](#4-persistent-volumes-creation)
  - [5. Create an Azure Managed PostgreSQL Server Database](#5-create-an-azure-managed-postgresql-server-database)
    - [Create A PostgreSQL Server Instance¶](#create-a-postgresql-server-instance)
    - [Create Your Databases](#create-your-databases)
    - [Update your yaml values](#update-your-yaml-values)
  - [6. Deploy Pachyderm](#6-deploy-pachyderm)
    - [Update Your Values.yaml](#update-your-valuesyaml)
    - [Deploy Pachyderm On The Kubernetes Cluster](#deploy-pachyderm-on-the-kubernetes-cluster)
  - [7. Have 'pachctl' And Your Cluster Communicate](#7-have-pachctl-and-your-cluster-communicate)
  - [8. Check That Your Cluster Is Up And Running](#8-check-that-your-cluster-is-up-and-running)
  - [9. NOTEBOOKS USERS: Install Pachyderm JupyterLab Mount Extension](#9-notebooks-users-install-pachyderm-jupyterlab-mount-extension)

## 1. Install Prerequisites

Before your start creating your cluster, install the following
clients on your machine. If not explicitly specified, use the
latest available version of the components listed below.

* [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli){target=_blank}
* [jq](https://stedolan.github.io/jq/download/){target=_blank}
* [kubectl](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli){target=_blank}
* [pachctl](../../../getting-started/local-installation#install-pachctl)
 
!!! Note
    This page assumes that you have an [Azure Subsciption](https://docs.microsoft.com/en-us/azure/guides/developer/azure-developer-guide#understanding-accounts-subscriptions-and-billing){target=_blank}.

## 2. Deploy Kubernetes

You can deploy Kubernetes on Azure by following the official [Azure Kubernetes Service documentation](https://docs.microsoft.com/en-us/azure/aks/tutorial-kubernetes-deploy-cluster?tabs=azure-cli){target=_blank}, [use the quickstart walkthrough](https://docs.microsoft.com/en-us/azure/aks/kubernetes-walkthrough){target=_blank}, or follow the steps in this section.

!!! Attention 
      Pachyderm recommends running your cluster on Kubernetes 1.19.0 and above.

At a minimum, you will need to specify the parameters below:

|Variable|Description|
|--------|-----------|
|RESOURCE_GROUP|A unique name for the resource group where Pachyderm is deployed. For example, `pach-resource-group`.|
|LOCATION|An Azure availability zone where AKS is available. For example, `centralus`.|
|NODE_SIZE|The size of the Kubernetes virtual machine (VM) instances. To avoid performance issues, Pachyderm recommends that you set this value to at least `Standard_DS4_v2` which gives you 8 CPUs, 28 Gib of Memory, 56 Gib SSD.<br> <br>In any case, use VMs that support **premium storage**. See [Azure VM sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes){target=_blank}  for details around which sizes support Premium storage.|
|CLUSTER_NAME|A unique name for the Pachyderm cluster. For example, `pach-aks-cluster`.|

You can choose to follow the guided steps in [Azure Service Portal's Kubernetes Services](https://portal.azure.com/){target=_blank} or use Azure CLI.


1. [Log in](https://docs.microsoft.com/en-us/cli/azure/authenticate-azure-cli){target=_blank} to Azure:

    ```shell
    az login
    ```

    This command opens a browser window. Log in with your Azure credentials.
    Resources can now be provisioned on the Azure subscription linked to your account.
    
    
1. Create an [Azure resource group](https://docs.microsoft.com/en-us/azure/azure-resource-manager/management/manage-resource-groups-portal#what-is-a-resource-group){target=_blank} or retrieve an existing group.

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

    For more configuration options: Find the list of [all available flags of the `az aks create` command](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_create){target=_blank}.

      ```shell
      az aks create --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME} --node-vm-size ${NODE_SIZE} --node-count <node_pool_count> --location ${LOCATION}
      ```

      **Example:**

      ```shell
      az aks create --resource-group test-group --name test-cluster --generate-ssh-keys --node-vm-size Standard_DS4_v2 --location centralus
      ```

1. Confirm the version of the Kubernetes server by running  `kubectl version`.

!!! note "See Also:"
    - [Azure Virtual Machine sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes-general){target=_blank}

Once your Kubernetes cluster is up, and your infrastructure configured, you are ready to prepare for the installation of Pachyderm. Some of the steps below will require you to keep updating the values.yaml started during the setup of the recommended infrastructure:


## 3. Create an Azure Storage Container For Your Data

Pachyderm needs an [Azure Storage Container](https://docs.microsoft.com/en-us/azure/databricks/data/data-sources/azure/azure-storage){target=_blank} (Object store) to store your data. 

To access your data, Pachyderm uses a [Storage Account](https://docs.microsoft.com/en-us/azure/storage/common/storage-account-overview){target=_blank} with permissioned access to your desired container. You can either use an existing account or create a new one in your default subscription, then use the JSON key associated with the account and pass it on to Pachyderm.

To create a new storage account, follow the steps below:

!!! Warning
      The storage account name must be unique in the Azure location.

* Set up the following variables:

      * STORAGE_ACCOUNT - The name of the storage account where you store your data.
      * CONTAINER_NAME - The name of the Azure blob container where you store your data.

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
    section in the [Azure Portal](https://portal.azure.com/){target=_blank} or by running the following command `az storage account keys list --account-name=${STORAGE_ACCOUNT}`.


* Create a new storage container within your storage account:

    ```shell
    az storage container create --name ${CONTAINER_NAME} \
              --account-name ${STORAGE_ACCOUNT} \
              --account-key "${STORAGE_KEY}"
    ```
## 4. Persistent Volumes Creation

etcd and PostgreSQL (metadata storage) each claim the creation of a pv. 

If you plan to deploy Pachyderm with its default bundled PostgreSQL instance, read the warning below and jump to the [deployment section](#6-deploy-pachyderm):

!!! Warning
    The metadata service (Persistent disk) generally requires a small persistent volume size (i.e. 10GB) but **high IOPS (1500)**, therefore, depending on your disk choice, you may need to oversize the volume significantly to ensure enough IOPS.

If you plan to deploy a managed PostgreSQL instance (Recommended in production), read the following section.

## 5. Create an Azure Managed PostgreSQL Server Database

By default, Pachyderm runs with a bundled version of PostgreSQL. 
For production environments, we strongly recommend that you disable the bundled version and use a PostgreSQL Server instance.

This section will provide guidance on the configuration settings you will need to:

- Create an environment to run your Azure PostgreSQL Server databases.
- Create one or **two databases** (`pachyderm` and, depending on whether your cluster is standalone or managed by an enterprise server, a second database, `dex`).
- Update your values.yaml to turn off the installation of the bundled postgreSQL and provide your new instance information.

!!! Note
    It is assumed that you are already familiar with PostgreSQL Server, or will be working with an administrator who is.

### Create A PostgreSQL Server Instance¶

!!! Info 
    Find the details of the steps and available parameters to create a PostgreSQL Server instance with Azure Console in Azure Documentation ["Create an Azure Database for PostgreSQL server by using the Azure portal"](https://docs.microsoft.com/en-us/azure/postgresql/quickstart-create-server-database-portal){target=_blank}. 
    
    Alternatively, you can use the cli and run [`az postgres server create`](https://docs.microsoft.com/en-us/cli/azure/postgres/server?view=azure-cli-latest){target=_blank} with your relevant parameters.

In the Azure console, choose the **Azure Database for PostgreSQL servers** service. You will be asked to pick your server type: `Single Server` or `Hyperscale` (for multi-tenant applications), then configure your DB instance as follows.

| SETTING | Recommended value|
|:----------------|:--------------------------------------------------------|
| *subscription*  and *resource group*| Pick your existing [resource group](https://docs.microsoft.com/en-us/azure/cloud-adoption-framework/govern/resource-consistency/resource-access-management#what-is-an-azure-resource-group){target=_blank}.<br><br> **Important** Your Cluster and your Database must be deployed in the **same ressource group**.|
|*server name*|Name your instance.|
|*location*|Create a database **in the region matching your Pachyderm cluster**.|
|*compute + storage*|The standard instance size (GP_Gen5_4 = Gen5 VMs with 4 cores) should work. Remember that Pachyderm's metadata services require **high IOPS (1500)**. Oversize the disk accordingly |
| *Master username* | Choose your Admin username. ("postgres")|
| *Master password* | Choose your Admin password.|

You are ready to create your instance. 

!!! Example
    ```shell
    az postgres server create \
        --resource-group <your_resource_group> \
        --name <your_server_name>  \
        --location westus \
        --sku-name GP_Gen5_2 \
        --admin-user <server_admin_username> \
        --admin-password <server_admin_password> \
        --ssl-enforcement Disabled \
        --version 11
    ```

!!! Warning
    - Make sure that your PostgreSQL version is `>= 11`
    - Keep the SSL setting `Disabled`.


Once created, go back to your newly created database, and: 

- Open the access to your instance:

!!! Note
    Azure provides two options for pods running on an AKS worker nodes to access a PostgreSQL DB instance, pick what fit you best:

      - Create a [firewall rule](https://docs.microsoft.com/en-us/azure/mysql/concepts-firewall-rules#connecting-from-azure){target=_blank} on the Azure DB Server with a range of IP addresses that encompasses all IPs of the AKS Cluster nodes (this can be a very large range if using node auto-scaling).
      - Create a [VNet Rule](https://docs.microsoft.com/en-us/azure/mysql/concepts-data-access-and-security-vnet){target=_blank} on the Azure DB Server that allows access from the subnet the AKS nodes are in. This is used in conjunction with the Microsoft.Sql VNet Service Endpoint enabled on the cluster subnet.

    You can also choose the more secure option to [deny public access to your PostgreSQL instance](https://docs.microsoft.com/en-us/azure/postgresql/howto-deny-public-network-access){target=_blank} then [Create a private endpoint in the K8s vnet](https://docs.microsoft.com/en-us/azure/postgresql/howto-configure-privatelink-cli){target=_blank}. Read more about how to [configure a private link using CLI on Azure's documentation](https://docs.microsoft.com/en-us/azure/postgresql/concepts-data-access-and-security-private-link){target=_blank}

   
   Alternativelly, in the **Connection Security** of your newly created server, *Allow access to Azure services* (This is equivalent to running `az postgres server firewall-rule create --server-name <your_server_name> --resource-group <your_resource_group> --name AllowAllAzureIps --start-ip-address 0.0.0.0 --end-ip-address 0.0.0.0`). 

- In the **Essentials** page of your instance, find the full **server name** and **admin username** that will be required in your [values.yaml](#update-your-yaml-values).

![Instance overview page](../images/azure_postgresql_overview.png)

### Create Your Databases
After your instance is created, you will need to create Pachyderm's database(s).
      
If you plan to deploy a standalone cluster (i.e., if you do not plan to register your cluster with a separate [enterprise server](../../enterprise/auth/enterprise-server/setup.md), you will need to create a second database named "dex" in your PostgreSQL Server instance for Pachyderm's authentication service. Note that the database **must be named `dex`**. This second database is not needed when your cluster is managed by an enterprise server.

!!! Note
    Read more about [dex on PostgreSQL in Dex's documentation](https://dexidp.io/docs/storage/#postgres){target=_blank}.

Pachyderm will use the same user to connect to `pachyderm` as well as to `dex`. 

### Update your yaml values
Once your databases have been created, add the following fields to your Helm values:


```yaml
global:
  postgresql:
    postgresqlUsername: "see admin username above"
    postgresqlPassword: "password"
    # The server name of the instance
    postgresqlDatabase: "pachyderm"
    # The postgresql database host to connect to. 
    postgresqlHost: "see server name above"
    # The postgresql database port to connect to. Defaults to postgres server in subchart
    postgresqlPort: "5432"

postgresql:
  # turns off the install of the bundled postgres.
  # If not using the built in Postgres, you must specify a Postgresql
  # database server to connect to in global.postgresql
  enabled: false
```

## 6. Deploy Pachyderm
You have set up your infrastructure, created your data container and a Managed PostgreSQL instance, and granted your cluster access to both: you can now finalize your values.yaml and deploy Pachyderm.

### Update Your Values.yaml  

!!! Note 
     If you have not created a Managed PostgreSQL Server instance, **replace the Postgresql section below** with `postgresql:enabled: true` in your values.yaml. This setup is **not recommended in production environments**.

If you have previously tried to run Pachyderm locally,
make sure that you are using the right Kubernetes context first. 

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

    Update your values.yaml with your container name ([see example of values.yaml here](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/microsoft-values.yaml){target=_blank}) or use our minimal example below.
       
    ```yaml
    deployTarget: "MICROSOFT"
    pachd:
      storage:
        microsoft:
          # storage container name
          container: "container_name"
          # storage account name
          id: "AKIAIOSFODNN7EXAMPLE"
          # storage account key
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      externalService:
        enabled: true
    global:
      postgresql:
        postgresqlUsername: "see admin username above"
        postgresqlPassword: "password"
        # The server name of the instance
        postgresqlDatabase: "pachyderm"
        # The postgresql database host to connect to. 
        postgresqlHost: "see server name above"
        # The postgresql database port to connect to. Defaults to postgres server in subchart
        postgresqlPort: "5432"
    postgresql:
      # turns off the install of the bundled postgres.
      # If not using the built in Postgres, you must specify a Postgresql
      # database server to connect to in global.postgresql
      enabled: false
    ```

    Check the [list of all available helm values](../../../reference/helm-values/) at your disposal in our reference documentation or on [Github](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/pachyderm/values.yaml){target=_blank}.

### Deploy Pachyderm On The Kubernetes Cluster

- Now you can deploy a Pachyderm cluster by running this command:


    ```shell
    helm repo add pach https://helm.pachyderm.com
    helm repo update
    helm install pachd -f values.yaml pach/pachyderm --version <version-of-the-chart>
    ```

    **System Response:**

    ```shell
    NAME: pachd
    LAST DEPLOYED: Mon Jul 12 18:28:59 2021
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    ```
    Refer to our generic [Helm documentation](../helm-install/#install-pachyderms-helm-chart) for more information on how to select your chart version. 

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

- Finally, make sure that [`pachctl` talks with your cluster](#7-have-pachctl-and-your-cluster-communicate).
## 7. Have 'pachctl' And Your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you are exposing your cluster publicly:
  1. Retrieve the external IP address of your TCP load balancer or your domain name:
  
     ```shell
     kubectl get services | grep pachd-lb | awk '{print $4}'
     ```

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
      ```
      ```shell
      pachctl config set active-context "<your-cluster-context-name>"
      ```

  1. Check that your are using the right context: 

      ```shell
      pachctl config get active-context
      ```

      Your cluster context name should show up.

If you're not exposing `pachd` publicly, you can run:

```shell
# Background this process because it blocks.
pachctl port-forward
``` 

## 8. Check That Your Cluster Is Up And Running

!!! Attention
    If Authentication is activated (When you deploy with an enterprise key already set, for example), you need to run `pachct auth login`, then authenticate to Pachyderm with your User, before you use `pachctl`. 

```shell
pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

## 9. NOTEBOOKS USERS: Install Pachyderm JupyterLab Mount Extension

Once your cluster is up and running, you can helm install JupyterHub on your Pachyderm cluster and experiment with your data in Pachyderm from your Notebook cells. 

Check out our [JupyterHub and Pachyderm Mount Extension](../../how-tos/jupyterlab-extension/index.md){target=_blank} page for installation instructions. 

Use Pachyderm's default image and values.yaml [`jupyterhub-ext-values.yaml`](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/jupyterhub-ext-values.yaml){target=_blank} or follow the instructions to update your own.

!!! Note
       Make sure to check our [data science notebook examples](https://github.com/pachyderm/examples){target=_blank} running on Pachyderm, from a market sentiment NLP implementation using a FinBERT model to pipelines training a regression model on the Boston Housing Dataset.