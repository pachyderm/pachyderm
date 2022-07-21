# Quickstart

 
On this page, you will find simplified deployment instructions and Helm values to get you started with the latest release of Pachyderm on the Kubernetes Engine of your choice (AWS (EKS), Google (GKS), and Azure (AKS)).


For each cloud provider, we will give you the option to "quick deploy" Pachyderm with or without Console (Pachyderm UI).

!!! Important 
    The deployment steps highlighted in this document are **not intended for production**. For production settings, please read our [infrastructure recommendations](../ingress/). In particular, we recommend:

     - the use of a **managed PostgreSQL server** (RDS, CloudSQL, or PostgreSQL Server) rather than Pachyderm's default bundled PostgreSQL.
     - the setup of a **TCP Load Balancer** in front of your pachd service.
     - the setup of an **Ingress Controller** in front of Console. 

    Then find your targeted Cloud provider in the [Deploy and Manage](../) ection of this documentation.


## 1. Prerequisites

Pachyderm in deployed on a Kubernetes Cluster.

Just before you start creating your cluster, install the following
clients on your machine. Use the
latest available version of the components listed below.

* [kubectl](https://docs.microsoft.com/en-us/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli){target=_blank}: the cli to interact with your cluster.
* [pachctl](../../../getting-started/local-installation#install-pachctl): the cli to interact with Pachyderm.
* Install [`Helm`](https://helm.sh/docs/intro/install/){target=_blank} for your deployment. 


!!! Warning "Optional - Quick deployment of Pachyderm with Console"
    - The deployment of Console (Pachyderm UI) **requires a valid enterprise token**. To get your free-trial token, fill in [this form](https://www.pachyderm.com/trial/){target=_blank}, get in touch with us at [sales@pachyderm.io](mailto:sales@pachyderm.io), or on our [Slack](https://www.pachyderm.com/slack/){target=_blank}. 
    - When deploying with Console, we create a default mock user (username:`admin`, password: `password`) to authenticate to Console without the hassle of connecting your Identity Provider. 

    For a better understanding of the additional steps and helm values needed when deploying with Console in a production environment, read about the [deployment of Pachyderm with Console](../console/#deploy-in-the-cloud) page. 


Select your favorite cloud provider.

!!! Important "Definition"
    Note that we often use the acronym `CE` for Community Edition.

## 2. Create Your Values.yaml
### AWS

1. Additional client installation:
Install [AWS CLI](https://aws.amazon.com/cli/){target=_blank}

1. [Create an EKS cluster](../aws-deploy-pachyderm/#2-deploy-kubernetes-by-using-eksctl) 

1. [Create an S3 bucket](../aws-deploy-pachyderm/#3-create-an-s3-bucket) for your data

1. Create a values.yaml

=== "Deploy Pachyderm without Console"

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
    ```
=== "Deploy Pachyderm with Console"

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

=== "Deploy Pachyderm without Console"

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
    ```
=== "Deploy Pachyderm with Console"

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
    - This section assumes that you have an [Azure Subsciption](https://docs.microsoft.com/en-us/azure/guides/developer/azure-developer-guide#understanding-accounts-subscriptions-and-billing){target=_blank}.

1. Additional client installation:
Install [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli){target=_blank}.

1. [Create an AKS cluster](../azure/#2-deploy-kubernetes) 

1. [Create a Storage Container](../azure/#3-create-an-azure-storage-container-for-your-data) for your data

1. Create a values.yaml

=== "Deploy Pachyderm without Console"

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
    ```
=== "Deploy Pachyderm with Console"

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
    helm install pachyderm -f my_pachyderm_values.yaml pach/pachyderm 
    ```

- Check your deployment:

    ```shell
    kubectl get pods
    ```

    Once the pods are up, you should see a pod for `pachd` running 
    (alongside etcd, pg-bouncer or postgres, console, depending on your installation). 
    If you are curious about the architecture of Pachyderm, take a look at our [high-level architecture diagram](../../).
    
    **System Response:**

    ```
    NAME                           READY   STATUS    RESTARTS   AGE
    etcd-0                         1/1     Running   0          18h
    pachd-5db79fb9dd-b2gdq         1/1     Running   2          18h
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

=== "You have deployed Pachyderm with Console"
    - To connect to your new Pachyderm instance, run:

        ```shell
        pachctl config import-kube local --overwrite
        ```
        ```shell
        pachctl config set active-context local
        ```

    - Then run `pachctl port-forward` (Background this process in a new tab of your terminal).

    - Note that you will need to run `pachctl auth login` then authenticate to Pachyderm with the mock User (`username`, `password`) to use `pachctl`

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
- Authenticate as the mock User using `admin` & `password` 

You are all set! 

## 6. Try our [beginner tutorial](../../../getting-started/beginner-tutorial/).


    

