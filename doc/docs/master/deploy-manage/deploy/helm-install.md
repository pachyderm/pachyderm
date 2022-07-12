# Helm Deployment

The package manager [Helm](https://helm.sh/docs/intro/install/#helm){target=_blank} is the authoritative deployment method for Pachyderm.

!!! Reminder
    **Pachyderm services are exposed on the cluster internal IP (ClusterIP) instead of each node’s IP (Nodeport)** except for LOCAL Helm installations (i.e. Services are still accessible through Nodeports on Local installations).

This page gives a high level view of the steps to follow to install Pachyderm using Helm. Find our chart on [Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm){target=_blank} or in our [GitHub repository](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/etc/helm/pachyderm){target=_blank}.

!!! Important "Before your start your installation process." 
      - Refer to this generic page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/). You will find instructions on setting up an ingress controller, a TCP load balancer, or connecting an Identity Provider for access control. 
      - If you are planning to install Pachyderm UI, read our [Console deployment](../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires the set up of a DNS, an Ingress, and the activation of authentication.
## Install
### Prerequisites
1. Install [`Helm`](https://helm.sh/docs/intro/install/){target=_blank}. 

1. Install [`pachctl`](../../../getting-started/local-installation/#install-pachctl), the command-line utility for interacting with a Pachyderm cluster. 

1. Choose the deployment [guidelines](https://docs.pachyderm.com/{{ config.pach_branch }}/deploy-manage/deploy/){target=_blank} that apply to you:
    * **Find the deployment page that applies to your Cloud provider** (or custom deployment, or on-premises deployment).
    It will help list the various installation prerequisites, and Kubernetes deployment instructions that fit your own use case:
    
        For example, if your Cloud provider is Google Cloud Platform, follow the **Prerequisites** and **Deploy Kubernetes** sections of the [deployment on Google Cloud Platform](https://docs.pachyderm.com/{{ config.pach_branch }}/deploy-manage/deploy/google-cloud-platform/#google-cloud-platform){target=_blank} page.

    * Additionally, those instructions will help you configure the various elements (object store, postgreSQL instance, secret names...) that relate to your deployment needs. Those parameters values will **be specified in your YAML configuration file** as follows.

### Edit a Values.yaml File
Create a personalized `my_pachyderm_values.yaml` out of this [example repository](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/etc/helm/examples){target=_blank}. Pick the example that fits your target deployment and update the relevant values according to the parameters gathered in the previous step.

Please refer to this section to [understand Pachyderm's main configuration values (License Key, IdP configuration, etc...)](#read-before-any-install-or-upgrade-pachyderm-configuration-values-and-platform-secrets) and how you can provide them hard-coded in your values or by using secrets. 

See the reference [values.yaml](../../../reference/helm-values/) for the list of all available helm values at your disposal.

!!! Warning
    **No default k8s CPU and memory requests and limits** are created for pachd.  If you don't provide values in the values.yaml file, then those requests and limits are simply not set. 
    
    For Production deployments, Pachyderm strongly recommends that you **[create your values.yaml file with CPU and memory requests and limits for both pachd and etcd](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/pachyderm/values.yaml){target=_blank}** set to values appropriate to your specific environment. For reference, 1 CPU and 2 GB memory for each is a sensible default. 
     
###  Install Pachyderm's Helm Chart
1. Get your Helm Repo Info
    ```shell
    helm repo add pach https://helm.pachyderm.com
    helm repo update
    ```

1. Install Pachyderm

    You are ready to deploy Pachyderm on the environment of your choice.
    ```shell
    helm install pachd -f my_pachyderm_values.yaml pach/pachyderm --version <your_chart_version>
    ```
    !!! Info "To choose a specific helm chart version"
        **Each chart version is associated with a given version of Pachyderm**. You will find the list of all available chart versions and their associated version of Pachyderm on  [Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm){target=_blank}.
        

        - You can choose a specific helm chart version by adding a `--version` flag (for example, `--version 0.3.0`) to your `helm install.`
        - No additional flag will install the latest GA release of Pachyderm by default. 
        - You can choose the latest pre-release version of the chart by using the flag `--devel` (pre-releases are versions of the chart that correspond to releases of Pachyderm that don't have the GA status yet).


        For example: When the 2.0 version of Pachyderm was a release candidate, using the flag `--devel` would let you install the latest RC of 2.0 while no flag would retrieve the newest GA (1.13.4). 
     
    

1. Check your deployment
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
### Have 'pachctl' and your Cluster Communicate

Assuming your `pachd` is running as shown above, make sure that `pachctl` can talk to the cluster.

If you are exposing your cluster publicly:

  1. Retrieve the external IP address of your TCP load balancer or your domain name:
    ```shell
    kubectl get services | grep pachd-lb | awk '{print $4}'
    ```

  1. Update the context of your cluster with their direct url, using the external IP address/domain name above:

      ```shell
      echo '{"pachd_address": "grpc://<external-IP-address-or-domain-name>:30650"}' | pachctl config set 
      ```
      ```shell
      context "<your-cluster-context-name>" --overwrite
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

Verify that `pachctl` and your cluster are connected:

```shell
pachctl version
```

**System Response:**

```
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

## Uninstall Pachyderm's Helm Chart
[Helm uninstall](https://helm.sh/docs/helm/helm_uninstall/){target=_blank} a release as easily as you installed it.
```shell
helm uninstall pachd 
```

We recommend making sure that everything is properly removed following a helm uninstall:

- The uninstall leaves your persistent volumes. To clean them up, run `kubectl get pvc` and delete the claims `data-postgres-0` and `etcd-storage-etcd-0`. 

!!! Attention
     Deleting pvs will result in the loss of your data.

- All other resources should have been removed by Helm. Run `kubectl get all | grep "etcd\|\pachd\|postgres\|pg-bouncer"` to make sure of it and delete any remaining resources where necessary.

- If your uninstall failed, there might be config jobs still running. Run `kubectl get jobs.batch | grep pachyderm` and delete any remaining job.

## Upgrade Pachyderm's Helm Chart
When a new version of Pachyderm's chart is released, or when you want to change the configuration of your release, use the [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/){target=_blank} command:

```shell
helm upgrade pachd -f my_new_pachyderm_values.yaml pach/pachyderm --version <your_chart_version>        
```

!!! Warning

        Make sure that your platform's secret names have been set properly. Refer to this [section](#edit-a-valuesyaml-file) for the list of secrets we recommend creating prior to installing the product, and what to do if you have not. **Failing to provide those values might cause your upgrade to fail.**


## READ BEFORE ANY INSTALL OR UPGRADE: Pachyderm Configuration Values and Platform Secrets

### Values Used By Pachyderm At Installation/Upgrade Time

Find the complete list of helm values that can control secret values here. 
We have grouped those keys in two categories:

- **Values that you will need to provide:**

    - `pachd.enterpriseLicenseKey`: Your enterprise license (When installing Enterprise)
    - `oidc.upstreamIDPs` : The list of dex connectors, each containing Oauth client info connecting to an upstream IDP (When setting up the Authentication).

- **Values autogenerated when left blank**. Can be any value. As an example, use the following command to generate your own `openssl rand -hex 16`:

    - `pachd.rootToken`: Password of your root user.
    - `pachd.enterpriseSecret`: Needed if you connect to an enterprise server.
    - `pachd.oauthClientSecret`: Pachd oidc config to connect to dex.
    - `console.config.oauthClientSecret`: Oauth client secret for Console. Required if you set Console Enterprise.  
    - `global.postgresqlPassword`: Password to your postgreSQL.   
    - `pachd.enterpriseRootToken`: Provide this token when installing a cluster that is under an Enterprise Server's umbrella.

Those configuration values can be provided to Pachyderm in various ways:

- A - [**Hard-coded in your values.yaml**](#a-provide-credentials-directly-in-your-valuesyaml)
- B - In a secret, by [**referencing a secret name in your values.yaml**](#b-use-secrets)
#### **A - Provide credentials directly in your values.yaml**

You can provide credentials and configuration values directly in your values.yaml or set them with a `--set` argument during the installation/upgrade. The [(Second table, Column B)](#mapping-table-between-external-secrets-fields-values-fields-and-pachyderm-platform-secrets) below lists the names of the fields in which you can hard code your values directly).

-  *A-1 - Provide the minimal amount of values and let Pachyderm generate the rest*

    - Provide your [Enterprise License](../../../enterprise/deployment/#activate-the-enterprise-edition){target=_blank} (For Enterprise users only). 
    - Optionally, your [IDPs configuration](../../../enterprise/auth/authentication/idp-dex/#create-your-idp-pachyderm-connection){target=_blank} (For Enterprise users using the Authentication feature)
    - If you are using a built-in PostgreSQL, we strongly recommend to set its password before your initial installation.

    !!! Warning "Built-in PostgreSQL users"
    
        ```yaml
        postgresql:
                postgresqlPassword: "abc123321cba"
                postgresqlPostgresPassword: "npLvsvhGcF"
        ```
        
        If no value is provided, Pachyderm will auto-generate a password during the first installation. That password will be reset at each upgrade, and new default value created, **causing your helm upgrade to fail unless you retrieve the initial password ( `{{"kubectl get secret postgres -o go-template='{{.data.postgresql-password | base64decode }}'"}}`) then inject it back into your values.

    The rest of the values will be generated for you.

- *A-2 - Choose and set your values rather than relying on autogeneration*
        
    Generate your own value using:

    ```shell
    openssl rand -hex 16
    f438329fc98302779be65eef226d32c1
    ```

#### **B - Use Secret(s)** 

If your organization uses tools like ArgoCD for Gitops, you might want to [Create those secrets](../../../how-tos/advanced-data-operations/secrets/#create-a-secret) ahead of time then supply their names in the `secretName` field of your values.yaml. 

Find the secret name field to reference your secret in your values.yaml (Column A) and its corresponding Secret Key (First column) in the second table below.
### Pachyderm Platform Secrets

Pachyderm inject the values hard coded in your values.yaml into **"platform secrets"** (see the complete list of secrets in the table below)  at the time of the deployment or upgrade (such as Postgresql admin login username and password, OAuth information to set up your IdP, or your enterprise license key).

!!! Note
    If no secret name is provided for the  secret name fields, Pachyderm will retrieve the dedicated plain-text secret values in the helm values and populate generic platform secrets (see list below) at the time of the installation, then re-use those at each upgrade. If no value is found in either one of those two cases, default values are used.


|Secret Name  <div style="width:190px"> | Key |	Description <div style="width:290px">|
|---------------------------------|-----|-------------|
|`pachyderm-auth`                 |	- root-token <div> - auth-config <div> - cluster-role-bindings | - Password of the "root" user of your cluster. <div> - Pachd oidc config to connect to dex.<div> - [Role Based Access declaration](../../enterprise/auth/authorization/index) at the cluster level. Used to define access control at the cluster level when deploying. For example: Give a specific group `ClusterAdmin` access at once, or give an entire company a default `RepoReader` Access to all repos on this cluster.|
|**`pachyderm-console-secret`**     | OAUTH_CLIENT_SECRET | Oauth client secret for Console. Required if you set Console Enterprise. |
|**`pachyderm-deployment-id-secret`** | CLUSTER_DEPLOYMENT_ID | Cluster identifier. |
|**`pachyderm-enterprise`** | enterprise-secret | For internal use. Used as a shared secret between an Enterprise Server and a Cluster to communicate. Always present when enterprise is on but used only when an Enterprise Server is set.|
|`pachyderm-identity` | upstream-idps | The list of dex connectors, each containing Oauth client info connecting to an upstream IDP. |
|`pachyderm-license` | enterprise-license-key | Your enterprise license. |
|`pachyderm-storage-secret` | *This content depends on what object store backs your installation of Pachyderm.*|Credentials for Pachyderm to access your object store.|
|`postgres` | - postgresql-password <div> - postgresql-postgres-password| Password for Pachyderm to Access Postgres. |

### Mapping Table Between External Secrets Fields, Values Fields, and Pachyderm Platform Secrets

Find the complete list of the secret key and secret name fields to reference a secret in a values.yaml, the secret values fields if you chose to hard code your values in your values.yaml rather than pass them in a secret, and Pachyderm's platform secrets and keys those values will be injected into.

!!! Important "Order of operations."
       Note that if no secret name is provided for the fields mentioned in **A** (see table above), Pachyderm will retrieve the dedicated plain-text secret values in the helm values (Column **B**) and populate (or autogenerate when lfet blank) its own platform secrets at the time of the installation/upgrade (Column **C**). 

|Secret KEY name| <div style="width:290px"> Description </div>| A - Create your secrets ahead <br> of your cluster creation| B - Pass credentials in values.yaml| <div style="width:250px"> C - Corresponding (Platform Secret, Key) in which the values provided in A or B will be injected.</div>| 
|------------|------------|-----|--------|---------|
|root-token| Root clusterAdmin| pachd.rootTokenSecretName |pachd.rootToken|(pachyderm-auth, rootToken)|
|root-token|Users give this token in their values.yaml when they install a cluster that is under an [Enterprise Server's](../../enterprise/auth/enterprise-server/setup) umbrella.|pachd.enterpriseRootTokenSecretName|pachd.enterpriseRootToken|????|
|postgresql-password|Password to your database|global.postgresql.postgresqlExistingSecretName <br> global.postgresql.postgresqlExistingSecretKey |global.postgresql.postgresqlPassword|(postgres, postgresql-password) , (postgres, postgresql-postgres-password)|
|OAUTH_CLIENT_SECRET|Oauth client secret for Console <br> Required if you set Console|console.config.oauthClientSecretSecretName |console.config.oauthClientSecret|(pachyderm-console-secret, OAUTH_CLIENT_SECRET)|
|enterprise-license-key|Your enterprise license|pachd.enterpriseLicenseKeySecretName |pachd.enterpriseLicenseKey|(pachyderm-license, enterprise-license-key)|
|auth-config| Oauth client secret for pachd| pachd.oauthClientSecretSecretName|pachd.oauthClientSecret|(pachyderm-auth, auth-config)|
|enterprise-secret|Needed if you connect to an enterprise server|pachd.enterpriseSecretSecretName  |pachd.enterpriseSecret| (pachyderm-enterprise, enterprise-secret)  |
|upstream-idps|The list of dex connectors, each containing Oauth client info connecting to an upstream IDP|oidc.upstreamIDPsSecretName|oidc.upstreamIDPs|(pachyderm-identity, upstream-idps)|


        