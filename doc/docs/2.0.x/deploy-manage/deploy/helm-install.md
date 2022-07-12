# Helm Deployment

The package manager [Helm](https://helm.sh/docs/intro/install/#helm){target=_blank} is the authoritative deployment method for Pachyderm.

!!! Reminder
    **Pachyderm services are exposed on the cluster internal IP (ClusterIP) instead of each node’s IP (Nodeport)** except for LOCAL Helm installations (i.e. Services are still accessible through Nodeports on Local installations).

This page gives a high level view of the steps to follow to install Pachyderm using Helm. Find our chart on [Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm){target=_blank} or in our [GitHub repository](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/etc/helm/pachyderm){target=_blank}.

!!! Important "Before your start your installation process." 
      - Refer to this generic page for more information on  how to install and get started with `Helm`.
      - Read our [infrastructure recommendations](../ingress/). You will find instructions on setting up an ingress controller, a load balancer, or connecting an Identity Provider for access control. 
      - If you are planning to install Pachyderm UI. Read our [Console deployment](../console/) instructions. Note that, unless your deployment is `LOCAL` (i.e., on a local machine for development only, for example, on Minikube or Docker Desktop), the deployment of Console requires the set up of an Ingress and the activation of authentication.
## Install
### Prerequisites
1. Install [`Helm`](https://helm.sh/docs/intro/install/){target=_blank}. 

1. Install [`pachctl`](../../../getting-started/local-installation/#install-pachctl), the command-line utility for interacting with a Pachyderm cluster. 

1. Choose the deployment [guidelines](https://docs.pachyderm.com/{{ config.pach_branch }}/deploy-manage/deploy/){target=_blank} that apply to you:
    * **Find the deployment page that applies to your Cloud provider** (or custom deployment, or on-premises deployment).
    It will help list the various installation prerequisites, Kubernetes deployment instructions, and kubectl installation that fit your own use case:
    
        For example, if your Cloud provider is Google Cloud Platform, follow the **Prerequisites** and **Deploy Kubernetes** sections of the [deployment on Google Cloud Platform](https://docs.pachyderm.com/{{ config.pach_branch }}/deploy-manage/deploy/google-cloud-platform/#google-cloud-platform){target=_blank} page.

    * Additionally, those instructions will also help you configure the various elements (object store, credentials...) that relate to your deployment needs. Those parameters values will **be specified in a YAML configuration file** as follows.

### Edit a Values.yaml File
Create a personalized `my_pachyderm_values.yaml` out of this [example repository](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/etc/helm/examples){target=_blank}. Pick the example that fits your target deployment and update the relevant values according to the parameters gathered in the previous step.   

See the reference [values.yaml](../../../reference/helm-values/) for the list of all available helm values at your disposal.

!!! Warning
    **No default k8s CPU and memory requests and limits** are created for pachd.  If you don't provide values in the values.yaml file, then those requests and limits are simply not set. 
    
    For Production deployments, Pachyderm strongly recommends that you **[create your values.yaml file with CPU and memory requests and limits for both pachd and etcd](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/etc/helm/pachyderm/values.yaml){target=_blank}** set to values appropriate to your specific environment. For reference, 1 CPU and 2 GB memory for each is a sensible default. 

!!! Note "**Platform Secrets: READ BEFORE ANY INSTALL OR UPGRADE**" 
        Pachyderm recommends using **"platform secrets"** to hold the values needed by a cluster at the time of the deployment (such as Postgresql admin login username and password, OAuth information to set up your IdP, or your enterprise license key). 
        You have the option to: 

        1. [Create those secrets](../../../how-tos/advanced-data-operations/secrets/#create-a-secret) ahead of time then supply their names in the `secretName` field of your values.yaml (Recommended option). 
        1. For a quick installation, put the secrets' values in the dedicated fields of your values.yaml. In such cases, those will populate Pachyderm's generic default `pachyderm-bootstrap-config` secret. 

        Find the complete list of helm values that can control secret values here: 
        ``` 
        global.postgresqlExistingSecretName 
        console.config.oauthClientSecretSecretName 
        pachd.enterpriseLicenseKeySecretName 
        pachd.rootTokenSecretName 
        pachd.enterpriseSecretSecretName 
        pachd.oauthClientSecretSecretName 
        pachd.enterpriseRootTokenSecretName 
        oidc.upstreamIDPsSecretName 
        ``` 
        
        It is important to note that if no secret name is provided for the fields mentioned above, Pachyderm will retrieve the dedicated plain-text secret values in the helm values and populate a generic, default, auto-generated secret (`pachyderm-bootstrap-config`) at the time of the installation. If no value is found in either one of those two cases, default values are used in `pachyderm-bootstrap-config`. Check the list of all secret values fields and  `pachyderm-bootstrap-config` keys in our [upgrade section](../../manage/upgrades/#troubleshoot-upgrades){target=_blank}.

        This generic secret `pachyderm-bootstrap-config` is reset at each upgrade, and new default values are created, causing the **helm upgrade to fail unless you retrieve your default values (for example: `{{"kubectl get secret pachyderm-bootstrap-config -o go-template='{{.data.rootToken | base64decode }}'"}}`), create a dedicated secret for each, then manually set each secret name back into their corresponding secret name field above.**
       
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
