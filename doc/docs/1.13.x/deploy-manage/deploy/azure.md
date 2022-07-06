# Azure

You can deploy Pachyderm in a new or existing Microsoft® Azure® Kubernetes
Service environment and use Azure's resource to run your Pachyderm
workloads. 
To deploy Pachyderm to AKS, you need to:

1. [Install Prerequisites](#install-prerequisites)
2. [Deploy Kubernetes](#deploy-kubernetes)
3. [Deploy Pachyderm](#deploy-pachyderm)
4. [Point your CLI `pachctl` to your cluster](#have-pachctl-and-your-cluster-communicate)

## Install Prerequisites

Before you can deploy Pachyderm on Azure, you need to configure a few
prerequisites on your client machine. If not explicitly specified, use the
latest available version of the components listed below.
Install the following:

* [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
* [jq](https://stedolan.github.io/jq/download/)
* [kubectl](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli)
* [pachctl](#install-pachctl)

### Install `pachctl`

 `pachctl` is a primary command-line utility for interacting with Pachyderm clusters.
 You can run the tool on Linux®, macOS®, and Microsoft® Windows® 10 or later operating
 systems and install it by using your favorite command line package manager.
 This section describes how you can install `pachctl` by using
 `brew` and `curl`.

 If you are installing `pachctl` on Windows, you need to first install
 Windows Subsystem (WSL) for Linux.

 To install `pachctl`, complete the following steps:

 - To install on macOS by using `brew`, run the following command:

    ```shell
    brew tap pachyderm/tap && brew install pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}
    ```
 - To install on Linux 64-bit or Windows 10 or later, run the following command:

    ```shell
    curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v{{ config.pach_latest_version }}/pachctl_{{ config.pach_latest_version }}_amd64.deb &&  sudo dpkg -i /tmp/pachctl.deb
    ```

 - Verify your installation by running `pachctl version`:

    ```shell
    pachctl version --client-only
    ```

    **System Response:**

    ```shell
    COMPONENT           VERSION
    pachctl             {{ config.pach_latest_version }}
    ```

## Deploy Kubernetes

You can deploy Kubernetes on Azure by following the official [Azure Container Service documentation](https://docs.microsoft.com/en-us/azure/aks/tutorial-kubernetes-deploy-cluster?tabs=azure-cli) or by
following the steps in this section. When you deploy Kubernetes on Azure,
you need to specify the following parameters:

|Variable|Description|
|--------|-----------|
|RESOURCE_GROUP|A unique name for the resource group where Pachyderm is deployed. For example, `pach-resource-group`.|
|LOCATION|An Azure availability zone where AKS is available. For example, `centralus`.|
|NODE_SIZE|The size of the Kubernetes virtual machine (VM) instances. To avoid performance issues, Pachyderm recommends that you set this value to at least `Standard_DS4_v2` which gives you 8 CPUs, 28 Gib of Memory, 56 Gib SSD.|
|CLUSTER_NAME|A unique name for the Pachyderm cluster. For example, `pach-aks-cluster`.|


To deploy Kubernetes on Azure, complete the following steps:

1. Log in to Azure:

    ```shell
    az login
    ```

    **System Response:**

    ```shell
    Note, we have launched a browser for you to login. For old experience with
    device code, use "az login --use-device-code"
    ```

    If you have not already logged in this command opens a browser window. Log in with your Azure credentials.
    After you log in, the following message appears in the command prompt:

    ```shell
    You have logged in. Now let us find all the subscriptions to which you have access...
    ```
    ```json
    [
      {
        "cloudName": "AzureCloud",
        "id": "your_id",
        "isDefault": true,
        "name": "Microsoft Azure Sponsorship",
        "state": "Enabled",
        "tenantId": "your_tenant_id",
        "user": {
          "name": "your_contact_id",
          "type": "user"
        }
      }
    ]
    ```

1. Create an Azure resource group.

    ```shell
    az group create --name=${RESOURCE_GROUP} --location=${LOCATION}
    ```

    **Example:**

    ```shell
    az group create --name="test-group" --location=centralus
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

1. Create an AKS cluster:

    ```shell
    az aks create --resource-group ${RESOURCE_GROUP} --name ${CLUSTER_NAME} --generate-ssh-keys --node-vm-size ${NODE_SIZE}
    ```

    **Example:**

    ```shell
    az aks create --resource-group test-group --name test-cluster --generate-ssh-keys --node-vm-size Standard_DS4_v2
    ```

    **System Response:**

    ```json
    {
      "aadProfile": null,
      "addonProfiles": null,
      "agentPoolProfiles": [
        {
          "availabilityZones": null,
          "count": 3,
          "enableAutoScaling": null,
          "maxCount": null,
          "maxPods": 110,
          "minCount": null,
          "name": "nodepool1",
          "orchestratorVersion": "1.12.8",
          "osDiskSizeGb": 100,
          "osType": "Linux",
          "provisioningState": "Succeeded",
          "type": "AvailabilitySet",
          "vmSize": "Standard_DS4_v2",
          "vnetSubnetId": null
        }
      ],
    ...
    ```

1. Confirm the version of the Kubernetes server:

    ```shell
    kubectl version
    ```

    **System Response:**

    ```shell
    Client Version: version.Info{Major:"1", Minor:"13", GitVersion:"v1.13.4", GitCommit:"c27b913fddd1a6c480c229191a087698aa92f0b1", GitTreeState:"clean", BuildDate:"2019-03-01T23:36:43Z", GoVersion:"go1.12", Compiler:"gc", Platform:"darwin/amd64"}
    Server Version: version.Info{Major:"1", Minor:"13", GitVersion:"v1.13.4", GitCommit:"c27b913fddd1a6c480c229191a087698aa92f0b1", GitTreeState:"clean", BuildDate:"2019-02-28T13:30:26Z", GoVersion:"go1.11.5", Compiler:"gc", Platform:"linux/amd64"}
    ```

!!! note "See Also:"
    - [Azure Virtual Machine sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes-general)


## Add storage resources

Pachyderm requires you to deploy an object store and a persistent
volume in your cloud environment to function correctly. For best
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

## Deploy Pachyderm

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

1. Run the following command:

    ```shell
    pachctl deploy microsoft ${CONTAINER_NAME} ${STORAGE_ACCOUNT} ${STORAGE_KEY} ${STORAGE_SIZE} --dynamic-etcd-nodes 1
    ```
    **Example:**

    ```shell
    pachctl deploy microsoft test-container teststorage <key> 10 --dynamic-etcd-nodes 1
    serviceaccount/pachyderm configured
    clusterrole.rbac.authorization.k8s.io/pachyderm configured
    clusterrolebinding.rbac.authorization.k8s.io/pachyderm configured
    service/etcd-headless created
    statefulset.apps/etcd created
    service/etcd configured
    service/pachd configured
    deployment.apps/pachd configured
    service/dash configured
    deployment.apps/dash configured
    secret/pachyderm-storage-secret configured

    Pachyderm is launching. Check its status with "kubectl get all"
    Once launched, access the dashboard by running "pachctl port-forward"
    ```

    Because Pachyderm pulls containers from DockerHub, it might take some time
    before the `pachd` pods start. You can check the status of the
    deployment by periodically running `kubectl get all`.

1. When pachyderm is up and running, get the information about the pods:

    ```shell
    kubectl get pods
    ```

    **System Response:**

    ```shell
    NAME                      READY     STATUS    RESTARTS   AGE
    dash-482120938-vdlg9      2/2       Running   0          54m
    etcd-0                    1/1       Running   0          54m
    pachd-1971105989-mjn61    1/1       Running   0          54m
    ```

    **Note:** Sometimes Kubernetes tries to start `pachd` nodes before
    the `etcd` nodes are ready which might result in the `pachd` nodes
    restarting. You can safely ignore those restarts.

## Have 'pachctl' and your Cluster Communicate

Finally, assuming your `pachd` is running as shown above, 
make sure that `pachctl` can talk to the cluster by:

- Running a port-forward:

```shell
# Background this process because it blocks.
pachctl port-forward   
```

- Exposing your cluster to the internet by setting up a LoadBalancer as follow:

!!! Warning 
    The following setup of a LoadBalancer only applies to pachd.

1. To get an external IP address for a Cluster, edit its k8s service, 
```shell
kubectl edit service pachd
```
and change its `spec.type` value from `NodePort` to `LoadBalancer`. 

1. Retrieve the external IP address of the edited service.
When listing your services again, you should see an external IP address allocated to the service you just edited. 
```shell
kubectl get service
```
1. Update the context of your cluster with their direct url, using the external IP address above:
```shell
echo '{"pachd_address": "grpc://<external-IP-address>:650"}' | pachctl config set context "your-cluster-context-name" --overwrite
```
1. Check that your are using the right context: 
```shell
pachctl config get active-context`
```
Your cluster context name set above should show up. 
    

You are done! You can test to make sure the cluster is working
by running `pachctl version` or even creating a new repo.

```shell
pachctl version
```

**System Response:**

```shell
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

