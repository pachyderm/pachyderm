# Quickstart

!!! Important  
    These instructions will help create a simple and easy first time deployment of the latest version of Pachyderm in the cloud (Azure, Google, AWS). 
    
    For each cloud provider, we will give you the option to install Pachyderm with or without Console (Pachyderm UI).

    While this is **not recommended in production settings**, this page can be useful for a quick test setup. For a production grade installation, read our [infrastructure recommendations](/deploy-manage/deploy/ingress/). In particular, we recommend to deploy a TCP Load Balancer in front of your pachd service and an Ingress Controller in front of console.


## 1. Prerequisistes

Before you start creating your cluster, install the following
clients on your machine. Use the
latest available version of the components listed below.

* [kubectl](https://docs.microsoft.com/cli/azure/aks?view=azure-cli-latest#az_aks_install_cli)
* [pachctl](../../../../getting_started/local_installation#install-pachctl)
* Install [`Helm`](https://helm.sh/docs/intro/install/). 

!!! Warning "Optional - Deployment of Pachyderm with Console"
    - The deployment of Console (Pachyderm UI) **requires a valid enterprise token**. To get your free-trial token, fill in [this form](https://www.pachyderm.com/trial), or get in touch with us at [sales@pachyderm.io](mailto:sales@pachyderm.io) or on our [Slack](http://slack.pachyderm.io/). 
    - Additionally, you will need to set up an [Ingress](../ingress/#ingress). To get you started quickly, check our [Traefik installation](../ingress/pach-ui-ingress/) documentation. 
    - When deploying, a mock user is setup for you to authenticate to Console:
        - username:`admin`
        - password: `password`
    - Attention: You will need to run `pachctl auth login` to use pachctl.
    
    More information about the [deployment of Pachyderm with Console](/deploy-manage/deploy/console/#deploy-in-the-cloud). 
    Once your Ingress is configured, jump to the Cloud provider of your choice:

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
        deployTarget: AMAZON

        pachd:
            storage:
                amazon:
                    bucket: "bucket_name"
                    
                    # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
                    id: AKIAIOSFODNN7EXAMPLE
                    
                    # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
                    secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
                    region: us-east-2
            service:
                type: "LoadBalancer"

        postgresql:
            # Enables the built in Postgres.
            enabled: true
            persistence:
                size: 500Gi
    ```
=== "Deploy Pachyderm with Console"

    ```yaml
        deployTarget: AMAZON

        pachd:
            storage:
                amazon:
                    bucket: "bucket_name"
                    
                    # this is an example access key ID taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html (AWS Credentials)
                    id: AKIAIOSFODNN7EXAMPLE
                    
                    # this is an example secret access key taken from https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html  (AWS Credentials)          
                    secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
                    region: us-east-2
            activateEnterprise: true
            # pachyderm enterprise key
            enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
            service:
                type: "LoadBalancer"

        console:
            enabled: true

        oidc:
            mockIDP: true

        postgresql:
            # Uses the built in Postgres.
            enabled: true
            persistence:
                size: 500Gi
    ```

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
    deployTarget: GOOGLE

    pachd:
        storage:
            google:
                bucket: "bucket_name"
            cred: INSERT JSON TO YOUR SERVICE ACCOUNT HERE
        service:
                type: "LoadBalancer"

    postgresql:
        # Uses the built in Postgres.
        enabled: true
    ```
=== "Deploy Pachyderm with Console"

    ```yaml
    deployTarget: GOOGLE

    pachd:
        storage:
            google:
                bucket: "bucket_name"
            cred: INSERT JSON TO YOUR SERVICE ACCOUNT HERE

        activateEnterprise: true
        # pachyderm enterprise key
        enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
        service:
            type: "LoadBalancer"

    console:
        enabled: true

    oidc:
        mockIDP: true

    postgresql:
        # Uses the built in Postgres.
        enabled: true
    ```

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
        deployTarget: MICROSOFT

        pachd:
            storage:
                microsoft:
                # storage container name
                container: blah

                # storage account name
                id: AKIAIOSFODNN7EXAMPLE

                # storage account key
                secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
            service:
                type: "LoadBalancer"

        postgresql:
            # Uses the built in Postgres.
            enabled: true
    ```
=== "Deploy Pachyderm with Console"

    ```yaml    
        deployTarget: MICROSOFT

        pachd:
            storage:
                microsoft:
                # storage container name
                container: blah

                # storage account name
                id: AKIAIOSFODNN7EXAMPLE

                # storage account key
                secret: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
            activateEnterprise: true
            # pachyderm enterprise key
            enterpriseLicenseKey: "YOUR_ENTERPRISE_TOKEN"
            service:
                type: "LoadBalancer"

        console:
            enabled: true

        oidc:
            mockIDP: true

        postgresql:
            # Uses the built in Postgres.
            enabled: true
    ```

Jump to [Helm install](#3-helm-install)

## 3. [Helm Install](../helm_install/#install-the-pachyderm-helm-chart)
- Your will be deploying the [latest GA release](../../../contributing/supported-releases/#generally-available-ga) of Pachyderm:

    ```shell
    $ helm repo add pach https://helm.pachyderm.com
    $ helm repo update
    $ helm install pachd -f my_pachyderm_values.yaml pach/pachyderm 
    ```


- Check your deployment:

    ```shell
    $ kubectl get pods
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

    - Retrieve the external IP address of pachd service (run `kubectl get services` and find the EXTERNAL_IP in front of pachd), then **update your context for pachctl to point at your cluster**:

        ```shell
        $ echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set context "<your-cluster-context-name>" --overwrite
        ```
    
    - **Attention**: If you have deployed Console, before you can use `pachctl`, you will need to run `pachct auth login` then authenticate with your Mock User.

    - Then run `pachctl port-forward` (Background this process in a new tab of your terminal).

- Check That Your Cluster Is Up And Running

    ```shell
    $ pachctl version
    ```

    **System Response:**

    ```shell
    COMPONENT           VERSION
    pachctl             {{ config.pach_latest_version }}
    pachd               {{ config.pach_latest_version }}
    ```

## 5. Connect to Console
To connect to your Console (Pachyderm UI):

- Point your browser to `http://localhost:80` 
- Authenticate as the mock User using `admin` & `password` 

You are all set! 

## 6. Or try our [beginner tutorial](../../-getting_started/beginner_tutorial/).


    

