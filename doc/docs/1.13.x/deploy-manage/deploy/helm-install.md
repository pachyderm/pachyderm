# Helm Deployment

Currently, the `pachctl deploy` command is the authoritative deployment method for Pachyderm.
However, you can now deploy Pachyderm using the package manager [Helm](https://helm.sh/docs/intro/install/#helm).

!!! Note
     - Helm support for Pachyderm is a **beta** release. 
        See our [supported releases documentation](https://docs.pachyderm.com/latest/reference/supported-releases/#release-status) for details.
     - Changes coming with Helm:
        For improved security, **Pachyderm services are now exposed on the cluster internal IP (ClusterIP) instead of each node’s IP (Nodeport)**. These changes do not apply to LOCAL Helm installations (i.e. Services are still accessible through Nodeports on Local installations)

This page gives you a high level view of the steps to follow to install Pachyderm using Helm. Find our charts on [Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm) or in our [GitHub repository](https://github.com/pachyderm/helmchart).

## Install
### Prerequisites
1. Install [`Helm`](https://helm.sh/docs/intro/install/). 

1. Choose the deployment [guidelines](https://docs.pachyderm.com/latest/deploy-manage/deploy/) that apply to you:
    * **Find the deployment page that applies to your Cloud provider** (or custom deployment, or on-premises deployment).
    It will help list the various installation prerequisites, Kubernetes deployment instructions, and kubectl installation that fit your own use case:
    
        For example, if your Cloud provider is Google Cloud Platform, follow the **Prerequisites** and **Deploy Kubernetes** sections of the [deployment on Google Cloud Platform](https://docs.pachyderm.com/latest/deploy-manage/deploy/google-cloud-platform/#google-cloud-platform) page.

    * Additionally, those instructions will also help you configure the various elements (persistent volume, object store, credentials...) that relate to your deployment needs. 
    In the case of a deployment using `pachctl deploy`, the values of those parameters would ultimately be passed to the corresponding arguments and flags of the command.

        For example, in the case of a generic `pachctl deploy google` command:"
        ```shell
        pachctl deploy google <bucket-name> <disk-size> [<credentials-file>] [flags]
        ```
        This command comes with an exhaustive [list of available flags](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_deploy_google/).
    
        In the case of an installation using Helm, those same parameters values will now **be specified in a YAML configuration file** as follows.

### Edit a values.yaml file
Create a personalized `my_pachyderm_values.yaml` out of this [example repository](https://github.com/pachyderm/helmchart/tree/pachyderm-0.6.5/examples). Pick the example that fits your target deployment and update the relevant fields according to the parameters gathered in the previous step.   

See the [conversion table](#conversion-table) at the end of this page. It should help you pass easily from the `pachctl deploy` arguments and flags to their attributes counterpart in values.yaml.

See also the reference [values.yaml](https://github.com/pachyderm/helmchart/blob/pachyderm-0.6.5/pachyderm/values.yaml) for an exhaustive list of all parameters.

!!! Warning
    When deploying Pachyderm using the "pachctl deploy" command (which was the only option prior to 1.13), **k8s CPU and memory requests and limits** were created for pachd, even if they were not explicitly set by flags on the command line. With the Helm Chart deployment, that is not the case.  If you don't provide values in the values.yaml file, then those requests and limits are simply not set. 
    
    For Production deployments, Pachyderm strongly recommends that you **create your values.yaml file with CPU and memory requests and limits for both pachd and etcd** set to values appropriate to your specific environment. For reference, with `pachctl deploy` the defaults used were 1 CPU and 2 GB memory for each.
###  Install the Pachyderm Helm Chart
1. Get your Helm Repo Info
    ```shell
    helm repo add pachyderm https://helm.pachyderm.com
    helm repo update
    ```

1. Install Pachyderm

    You are ready to deploy Pachyderm on the environment of your choice.
    ```shell
    helm install pachd -f my_pachyderm_values.yaml pachyderm/pachyderm --version 0.6.5
    ```
You can choose a specific helm chart version by adding a `--version` flag (for example, `--version 0.3.0`). 
Each version of a chart is associated with a given version of Pachyderm. 
Here, the `--version 0.6.5` will install Pachyderm 1.13.3.
No mention of the version will install the latest available version of Pachyderm by default. 
[Artifacthub](https://artifacthub.io/packages/helm/pachyderm/pachyderm) lists all available chart versions and their associated version of Pachyderm. 

### Check your installation

1. Check your deployment
    ```shell
    kubectl get pods
    ```

    **System Response:**

    ```
    NAME                     READY     STATUS    RESTARTS   AGE
    dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
    etcd-0                   1/1       Running   0          4m
    pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
    ```

1. Verify that the Pachyderm cluster is up and running

    ```shell
    pachctl version
    ```

    **System Response:**

    ```
    COMPONENT           VERSION
    pachctl             {{ config.pach_latest_version }}
    pachd               {{ config.pach_latest_version }}
    ```

## Uninstall the Pachyderm Helm Chart
[Helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) a release as easily as you installed it.
```shell
helm uninstall pachd 
```

## Conversion table
| FLAG OPTION | Values.yaml ATTRIBUTE | DEFAULT|
|-------------|-----------------------|-------------------|
|**common** |----|---|
|--image-pull-secret|imagePullSecret| ""|
|**dash**|----|---|
|--dash-image|dash.image.repository|pachyderm/dash|
|--registry|dash.image.tag |"0.5.57"|
|**etcd**|----|---|
|--dynamic-etcd-nodes|etcd.dynamicNodes|1|
|--etcd-storage-class|etcd.storageClass|""|
|--etcd-cpu-request|etcd.resources.requests.cpu|"1"|
|--etcd-memory-request|etcd.resources.requests.memory|"2G"|
|**pachd**|----|---|
|--block-cache-size|pachd.blockCacheBytes|"1G"|
|--cluster-deployment-id|pachd.clusterDeploymentID|""|
|inverse of --no-expose-docker-socket|pachd.exposeDockerSocket|false|
|--expose-object-api|pachd.exposeObjectAPI|false|
|--pachd-cpu-request|pachd.resources.requests.cpu|"1"|
|--pachd-memory-request|pachd.resources.requests.memory|"2G"|
|--require-critical-servers-only|pachd.requireCriticalServersOnly|false|
|--shards|pachd.numShards|16|
|--worker-service-account|pachd.workerServiceAccount.name|pachyderm-worker|
|**pachd.storage**|----|---|
|--put-file-concurrency-limit|pachd.storage.putFileConcurrencyLimit|100|
|--upload-concurrency-limit|pachd.storage.uploadConcurrencyLimit|100|
|**pachd.storage.amazon**|----|---|
|--cloudfront-distribution|pachd.storage.amazon.cloudFrontDistribution|""|
|--credentials|`id` together with `secret` and `token`, <br> implements the functionality of the <br> --credentials argument to pachctl deploy.<br><ul><li>pachd.storage.amazon.id</li><li>pachd.storage.amazon.secret</li><li> pachd.storage.amazon.token</li></ul>|all 3 default to " "|
|--disable-ssl|pachd.storage.amazon.disableSSL|false|
|--iam-role|pachd.storage.amazon.iamRole|""|
|--max-upload-parts|pachd.storage.amazon.maxUploadParts|10000|
|--no-verify-ssl|pachd.storage.amazon.noVerifySSL|false|
|--obj-log-options|pachd.storage.amazon.logOptions|Comma-separated list containing zero or more of: 'Debug', 'Signing', 'HTTPBody', 'RequestRetries','RequestErrors', 'EventStreamBody', or 'all' (case-insensitive).  See 'AWS SDK for Go' docs for details. Default to: ""|
|--part-size|pachd.storage.amazon.partSize|5242880|
|--retries|pachd.storage.amazon.retries|10|
|--reverse|pachd.storage.amazon.reverse|true|
|--timeout|pachd.storage.amazon.timeout|"5m"|
|--upload-acl|pachd.storage.amazon.uploadACL|bucket-owner-full-control|
| **pachd.storage.local**|----|---|
|--host-path|pachd.storage.local.hostPath|"/var/pachyderm/"|
|**rbac**|----|---|
|--local-roles|rbac.clusterRBAC|true|
|--no-rbac|opposite of rbac.create|true|
|**tls**|----|---|
|--tls|<ul><li>tls.crt</li><li>tls.key</li></ul>|both default to ""|

## `pachctl deploy` flag deprecation

!!! Info "Deprecation notice"
        With the addition of the Helm chart, the following `pachctl deploy`
        flags are deprecated and will be removed in the future:

        - dash-image
        - dashboard-only
        - no-dashboard
        - expose-object-api
        - storage-v2
        - shards
        - no-rbac
        - no-guaranteed
        - static-etcd-volume
        - disable-ssl
        - max-upload-parts
        - no-verify-ssl
        - obj-log-options
        - part-size
        - retries
        - reverse
        - timeout
        - upload-acl



