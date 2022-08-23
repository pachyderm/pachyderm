# Quickstart

On this page, you will find simplified deployment instructions and Helm values to get you started with the latest release of Pachyderm on the Kubernetes Engine of your choice (AWS (EKS), Google (GKS), and Azure (AKS)).

For each cloud provider, we will give you the option to "quick deploy" Pachyderm with or without an enterprise key. A quick deployment allows you to experiment with Pachyderm without having to go through any infrastructure setup. In particular, you do not need to set up any object store or PostgreSQL instance.

!!! Important 
    The deployment steps highlighted in this document are **not intended for production**. For production settings, please read our [infrastructure recommendations](../ingress/). In particular, we recommend:

     - the use of a **managed PostgreSQL server** (RDS, CloudSQL, or PostgreSQL Server) rather than Pachyderm's default bundled PostgreSQL.
     - the setup of a **TCP Load Balancer** in front of your pachd service.
     - the setup of an **Ingress Controller** in front of Console. 

    Then find your targeted Cloud provider in the [Deploy and Manage](../) section of this documentation.

!!! Attention 
    We are now shipping Pachyderm with an **optional embedded proxy** 
    allowing your cluster to expose one single port externally. This deployment setup is optional.
    
    If you choose to deploy Pachyderm with a Proxy, check out our new recommended architecture and [deployment instructions](../deploy-w-proxy/). 

    Deploying with a proxy presents a couple of advantages:

    - You only need to set up one TCP Load Balancer (No more Ingress in front of Console).
    - You will need one DNS only.
    - It simplifies the deployment of Console.
    - No more port-forward.

## 1. Prerequisites

Pachyderm is deployed on a Kubernetes Cluster.

Install the following clients on your machine before you start creating your cluster. 
Use the latest available version of the components listed below.

* [kubectl](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli){target=_blank}: the cli to interact with your cluster.
* [pachctl](../../../getting-started/local-installation#install-pachctl): the cli to interact with Pachyderm.
* Install [`Helm`](https://helm.sh/docs/intro/install/){target=_blank} for your deployment. 

!!! Warning "Get a Pachyderm Enterprise key"
    To get a free-trial token, fill in [this form](https://www.pachyderm.com/trial/){target=_blank}, get in touch with us at [sales@pachyderm.io](mailto:sales@pachyderm.io), or on our [Slack](https://www.pachyderm.com/slack/){target=_blank}. 

Select your favorite cloud provider.

!!! Important "Definition"
    Note that we often use the acronym `CE` for Community Edition.
## 2. Create Your Values.yaml

!!! Note
    Pachyderm comes with a [Web UI (Console)](../console/#deploy-in-the-cloud) per default.

### AWS

1. Additional client installation:
Install [AWS CLI](https://aws.amazon.com/cli/){target=_blank}

1. [Create an EKS cluster](../aws-deploy-pachyderm/#2-deploy-kubernetes-by-using-eksctl) 

1. [Create an S3 bucket](../aws-deploy-pachyderm/#3-create-an-s3-bucket) for your data

1. Create a values.yaml


=== "Deploy Pachyderm CE (includes Console CE)"

    ```yaml
    deployTarget: "AMAZON"
    pachd:
      storage:
        amazon:
          bucket: "bucket_name"      
          # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
          id: "AKIAIOSFODNN7EXAMPLE"                
          # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
          region: "us-east-2"
      externalService:
        enabled: true
    console:
      enabled: true
    ```
=== "Deploy Pachyderm Enterprise with Console"
    Note that when deploying Pachyderm Enterprise with Console, **we create a default mock user (username:`admin`, password: `password`)** to authenticate yourself to Console so you don't have to connect an Identity Provider to make things work. The mock user is a [Cluster Admin](../../enterprise/auth/authorization/index.md#cluster-roles){target=_blank} per default.

    ```yaml
    deployTarget: "AMAZON"
    pachd:
      storage:
        amazon:
          bucket: "bucket_name"                
          # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
          id: "AKIAIOSFODNN7EXAMPLE"                
          # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
          region: "us-east-2"
      # pachyderm enterprise key 
      enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
    console:
      enabled: true
    ```
     

Jump to [Helm install](#3-helm-install)

### Google
1. Additional client installation:
Install [Google Cloud SDK](https://cloud.google.com/sdk/){target=_blank}

1. [Create a GKE cluster](../google-cloud-platform/#2-deploy-kubernetes)
Note: 
Add `--scopes storage-rw` to your `gcloud container clusters create` command. 

1. [Create a GCS Bucket](../google-cloud-platform/#3-create-a-gcs-bucket) for your data

1. Create a values.yaml

=== "Deploy Pachyderm CE (includes Console CE)"

    ```yaml
    deployTarget: "GOOGLE"
    pachd:
      storage:
        google:
          bucket: "bucket_name"
          cred: |
            INSERT JSON CONTENT HERE
      externalService:
        enabled: true
    console:
      enabled: true
    ```
=== "Deploy Pachyderm Enterprise with Console"
    Note that when deploying Pachyderm Enterprise with Console, **we create a default mock user (username:`admin`, password: `password`)** to authenticate yourself to Console so you don't have to connect an Identity Provider to make things work. The mock user is a [Cluster Admin](../../enterprise/auth/authorization/index.md#cluster-roles){target=_blank} per default.

    ```yaml
    deployTarget: "GOOGLE"
    pachd:
      storage:
        google:
          bucket: "bucket_name"
          cred: |
            INSERT JSON CONTENT HERE
      # pachyderm enterprise key
      enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
    console:
      enabled: true
    ```

Jump to [Helm install](#3-helm-install)

### Azure

!!! Note
    - This section assumes that you have an [Azure Subscription](https://docs.microsoft.com/en-us/azure/guides/developer/azure-developer-guide#understanding-accounts-subscriptions-and-billing){target=_blank}.

1. Additional client installation:
Install [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli){target=_blank}.

1. [Create an AKS cluster](../azure/#2-deploy-kubernetes) 

1. [Create a Storage Container](../azure/#3-create-an-azure-storage-container-for-your-data) for your data

1. Create a values.yaml

=== "Deploy Pachyderm CE (includes Console CE)"

    ```yaml
    deployTarget: "MICROSOFT"
    pachd:
      storage:
        microsoft:
          # storage container name
          container: "blah"
          # storage account name
          id: "AKIAIOSFODNN7EXAMPLE"
          # storage account key
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      externalService:
        enabled: true
    console:
      enabled: true
    ```
=== "Deploy Pachyderm Enterprise with Console"
    Note that when deploying Pachyderm Enterprise with Console, **we create a default mock user (username:`admin`, password: `password`)** to authenticate yourself to Console so you don't have to connect an Identity Provider to make things work. The mock user is a [Cluster Admin](../../enterprise/auth/authorization/permissions.md#clusteradminrole){target=_blank} per default.

    ```yaml    
    deployTarget: "MICROSOFT"
    pachd:
      storage:
        microsoft:
          # storage container name
          container: "blah"
          # storage account name
          id: "AKIAIOSFODNN7EXAMPLE"
          # storage account key
          secret: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
      # pachyderm enterprise key
      enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
    console:
      enabled: true
    ```

Jump to [Helm install](#3-helm-install)

## 3. [Helm Install](../helm-install/#install-pachyderms-helm-chart)
- You will be deploying the [latest GA release](../../../reference/supported-releases/#generally-available-ga) of Pachyderm:

    ```shell
    helm repo add pach https://helm.pachyderm.com
    helm repo update
    helm install pachd pach/pachyderm -f my_pachyderm_values.yaml 
    ```

- Check your deployment:

    ```shell
    kubectl get pods
    ```
    The deployment takes some time. You can run `kubectl get pods` periodically
    to check the status of your deployment. 

    Once all the pods are up, you should see a pod for `pachd` running 
    (alongside etcd, pg-bouncer or postgres, console, depending on your installation). 
    If you are curious about the architecture of Pachyderm, take a look at our [high-level architecture diagram](../../).
    
    **System Response:**

    ```
    NAME                           READY   STATUS    RESTARTS   AGE
    console-7b69ddf66d-bxmg5       1/1     Running   0          18h
    etcd-0                         1/1     Running   0          18h
    pachd-5db79fb9dd-b2gdq         1/1     Running   2          18h
    pg-bouncer-55d9c86768-g8lx7    1/1     Running   0          18h
    postgres-0                     1/1     Running   0          18h
    ```

## 4. Have 'pachctl' And Your Cluster Communicate

=== "You have deployed Pachyderm without Console"

    - Retrieve the external IP address of pachd service:
        ```shell
        kubectl get services | grep pachd-lb | awk '{print $4}'
        ```
    - Then **update your context for pachctl to point at your cluster**:

        ```shell
        echo '{"pachd_address": "grpc://<external-IP-address>:30650"}' | pachctl config set context "<choose-a-cluster-context-name>" --overwrite
        ```

        ```shell
        pachctl config set active-context "<your-cluster-context-name>"
        ```
    - If Authentication is activated (When you deploy with an enterprise key already set, for example), you need to run `pachct auth login`, then authenticate to Pachyderm with your mock User (username:`admin`, password: `password`), before you use `pachctl`. 

=== "You have deployed Pachyderm with Console"
    - To connect to your new Pachyderm instance, run:

        ```shell
        pachctl config import-kube local --overwrite
        ```
        ```shell
        pachctl config set active-context local
        ```
    - Then run `pachctl port-forward` (Background this process in a new tab of your terminal).


- Finally, check that your cluster is up and running

    ```shell
    pachctl version
    ```

    **System Response:**

    ```shell
    COMPONENT           VERSION
    pachctl             {{ config.pach_latest_version }}
    pachd               {{ config.pach_latest_version }}
    ```

## 5. Connect to Console
To connect to your Console (Pachyderm UI):

- Point your browser to `http://localhost:4000` 
- If Authentication is activated (When you deploy with an enterprise key already set, for example), you you will be prompted to authenticate: Use your mock User (username:`admin`, password: `password`).

You are all set! 

## 6. Try our [beginner tutorial](../../../getting-started/beginner-tutorial/).
## 7. NOTEBOOKS USERS: Install Pachyderm JupyterLab Mount Extension

Once your cluster is up and running, you can helm install JupyterHub on your Pachyderm cluster and experiment with your data in Pachyderm from your Notebook cells. 

Check out our [JupyterHub and Pachyderm Mount Extension](../../how-tos/jupyterlab-extension/index.md){target=_blank} page for installation instructions. 

Use Pachyderm's default image and values.yaml [`jupyterhub-ext-values.yaml`](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/examples/jupyterhub-ext-values.yaml){target=_blank} or follow the instructions to update your own.

!!! Note
       Make sure to check our [data science notebook examples](https://github.com/pachyderm/examples){target=_blank} running on Pachyderm, from a market sentiment NLP implementation using a FinBERT model to pipelines training a regression model on the Boston Housing Dataset.
    

