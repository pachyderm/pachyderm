# Quickstart

 
These instructions provide a simple deployment configuration of the latest GA version of Pachyderm on AWS (EKS), Google (GKS), and Azure (AKS).

<!--For each cloud provider, we will give you the option to install Pachyderm with or without Console(Pachyderm UI).-->

!!! Important 
    The deployment steps highlighted in this document are **not intended for production**. If you wish to deploy Pachyderm in production, please read our [infrastructure recommendations](../ingress/). In particular, we recommend:

     - the use of a **managed PostgreSQL server** (RDS, CloudSQL, or PostgreSQL Server) rather than Pachyderm's default bundled PostgreSQL.
     - the setup of a **TCP Load Balancer** in front of your pachd service.
     - the setup of an **Ingress Controller** in front of console. 


## 1. Prerequisites

Before you start creating your cluster, install the following
clients on your machine. Use the
latest available version of the components listed below.

* [kubectl](https://docs.microsoft.com/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli)
* [pachctl](../../../../getting_started/local_installation#install-pachctl)
* Install [`Helm`](https://helm.sh/docs/intro/install/). 


!!! Warning "Optional - Deployment of Pachyderm with Console"
    - The deployment of Console (Pachyderm UI) **requires a valid enterprise token**. To get your free-trial token, fill in [this form](https://www.pachyderm.com/trial), or get in touch with us at [sales@pachyderm.io](mailto:sales@pachyderm.io) or on our [Slack](http://slack.pachyderm.io/). 
    - When deploying with `mockIDP: true` (see your values.yaml below), a mock user (username:`admin`, password: `password`) is created to authenticate to Console.
    
    More information about the [deployment of Pachyderm with Console](/deploy-manage/deploy/console/#deploy-in-the-cloud). 


Select your favorite cloud provider.

## 2. Create Your Values.yaml
### AWS

1. Additional client installation:
Install [AWS CLI](https://aws.amazon.com/cli/)

1. [Create an EKS cluster](../amazon_web_services/deploy-eks/) 

1. [Create an S3 bucket](../amazon_web_services/aws-deploy-pachyderm/#1-create-an-s3-bucket) for your data

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
=== "Coming Soon... Deploy Pachyderm with Console"

     In the meantime... 

      - Try [Console on Hub](https://hub.pachyderm.com/)
      - Deploy Pachyderm and Console [locally](../../../getting_started/local_installation/) 
      - Or [contact us](mailto:support@pachyderm.io)! We are happy to Help.
     

Jump to [Helm install](#3-helm-install)

### Google
1. Additional client installation:
Install [Google Cloud SDK](https://cloud.google.com/sdk/) 

1. [Create a GKE cluster](../google_cloud_platform/#2-deploy-kubernetes)
Note: 
Add `--scopes storage-rw` to your `gcloud container clusters create` command. 

1. [Create a GCS Bucket](../google_cloud_platform/#create-a-gcs-bucket) for your data

1. Create a values.yaml

=== "Deploy Pachyderm without Console"

    ```yaml
    deployTarget: "GOOGLE"
    pachd:
      storage:
        google:
          bucket: "bucket_name"
          cred: "INSERT JSON TO YOUR SERVICE ACCOUNT HERE"
      externalService:
        enabled: true
    ```
=== "Coming Soon... Deploy Pachyderm with Console"

     In the meantime... 

      - Try [Console on Hub](https://hub.pachyderm.com/)
      - Deploy Pachyderm and Console [locally](../../../getting_started/local_installation/) 
      - Or [contact us](mailto:support@pachyderm.io)! We are happy to Help.

Jump to [Helm install](#3-helm-install)

### Azure

!!! Note
    - This section assumes that you have an [Azure Subsciption](https://docs.microsoft.com/en-us/azure/guides/developer/azure-developer-guide#understanding-accounts-subscriptions-and-billing).

1. Additional client installation:
Install [Azure CLI 2.0.1 or later](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli).

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
=== "Coming Soon...  Deploy Pachyderm with Console"

     In the meantime... 

      - Try [Console on Hub](https://hub.pachyderm.com/)
      - Deploy Pachyderm and Console [locally](../../../getting_started/local_installation/) 
      - Or [contact us](mailto:support@pachyderm.io)! We are happy to Help.

Jump to [Helm install](#3-helm-install)

## 3. [Helm Install](../helm_install/#install-the-pachyderm-helm-chart)
- Your will be deploying the [latest GA release](../../../contributing/supported-releases/#generally-available-ga) of Pachyderm:

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
    
    **System Response:**

    ```
    NAME                           READY   STATUS    RESTARTS   AGE
    etcd-0                         1/1     Running   0          18h
    pachd-5db79fb9dd-b2gdq         1/1     Running   2          18h
    postgres-0                     1/1     Running   0          18h
    ```

## 4. Have 'pachctl' And Your Cluster Communicate

- Make sure that `pachctl` can talk to the cluster:

    - Retrieve the external IP address of pachd service (run `kubectl get services | grep pachd-lb | awk '{print $4}'`), then **update your context for pachctl to point at your cluster**:

        ```shell
        echo '{"pachd_address": "grpc://<external-IP-address>:30650"}' | pachctl config set context "<choose-a-cluster-context-name>" --overwrite
        pachctl config set active-context "<choose-a-cluster-context-name>"
        ```

    - Then run `pachctl port-forward` (Background this process in a new tab of your terminal).

- Check That Your Cluster Is Up And Running

   !!! Note
    If Authentication is activated (When you deploy Console, for example), you will need to run `pachct auth login`, then authenticate to Pachyderm with your User, before you use `pachctl`. 


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

## 6. Try our [beginner tutorial](../../../getting_started/beginner_tutorial/).


    

