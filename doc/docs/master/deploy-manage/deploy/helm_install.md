# Helm Deployment

Currently, the `pachctl deploy` command is the authoritative deployment method for Pachyderm.
However, you can now deploy Pachyderm using the package manager [Helm](https://helm.sh/docs/intro/install/#helm).

!!! Note
      Helm support for Pachyderm is a **beta** release. 
      See our [supported releases documentation](https://docs.pachyderm.com/latest/contributing/supported-releases/#release-status) for details.

This page gives you a high level view of the steps to follow to install Pachyderm using Helm.

As a general rule, those steps would be:

## Prerequisites
1. Install [`Helm`](https://helm.sh/docs/intro/install/). 

1. Choose the deployment [guidelines](https://docs.pachyderm.com/latest/deploy-manage/deploy/) that apply to you:
    * **Find the deployment page that applies to your Cloud provider** (Custom deployment, or on-premises deployment).
    It will help list the various installation prerequisites, Kubernetes deployment instructions, and kubectl installation applying to your own use case:
    
        For example, if your Cloud provider is Google Cloud Platform, follow the **Prerequisites** and **Deploy Kubernetes** sections of the [deployment on Google Cloud Platform](https://docs.pachyderm.com/latest/deploy-manage/deploy/google_cloud_platform/#google-cloud-platform) page.

    * Additionaly, those instructions will also help you define and configure the various elements (persistent volume, object store, credentials...) that relate to your deployment needs. 
    In the case of a deployment using `pachctl deploy`, the values of those parameters would ultimately be passed to the 
    corresponding arguments and flags of the command.

        For example, in the case of a generic `pachctl deploy google` command:"
        ```shell
        pachctl deploy google <bucket-name> <disk-size> [<credentials-file>] [flags]
        ```
        This command comes with an exhaustive [list of available flags](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_deploy_google/).
    
        In the case of an installation using Helm, those same parameters values will now **populate a .yaml configuration file** as follow.

## Edit a values.yaml file
Create a personalized `my_pachyderm_values.yaml` out of this [example repository](https://github.com/pachyderm/helmchart/tree/master/examples). Pick the example that fits your target deployment and update the relevant fields according to the parameters gathered in the previous step.   

See the [conversion table](#conversion-table) at the end of this page. It should help you pass easily from the `pachctl deploy` arguments and flags to their attributes counterpart in values.yaml.

See also the reference [values.yaml](https://github.com/pachyderm/helmchart/blob/master/pachyderm/values.yaml) for an exhaustive list of all parameters.

## Get your Helm Repo Info
```shell
$ helm repo add pachyderm https://pachyderm.github.io/helmchart
```
```shell
$ helm repo update
```

## Install the Pachyderm helm chart ([helm v3](https://helm.sh/docs/intro/))
You are ready to deploy Pachyderm on the environment of your choise.
```shell
$ helm install pachd -f my_pachyderm_values.yaml pachyderm/pachyderm
```

## Check your deployment 

```shell
$ kubectl get pods
```

**System Response:**

```
NAME                     READY     STATUS    RESTARTS   AGE
dash-6c9dc97d9c-89dv9    2/2       Running   0          1m
etcd-0                   1/1       Running   0          4m
pachd-65fd68d6d4-8vjq7   1/1       Running   0          4m
```

## Verify that the Pachyderm cluster is up and running

```shell
$ pachctl version
```

**System Response:**

```
COMPONENT           VERSION
pachctl             {{ config.pach_latest_version }}
pachd               {{ config.pach_latest_version }}
```

## Conversion table
| FLAG OPTION | Values.yaml ATTRIBUTE |
|-------------|-----------------------|
|--tls|tls.crt|
||tls.key|
|--image-pull-secret|imagePullSecret|
|--dash-image|dash.image.repository|
||dash.image.tag|
|opposite of --no-dashboard|dash.enabled |
|--dynamic-etcd-nodes|etcd.dynamicNodes|
|--etcd-storage-class|etcd.storageClass|
|--etcd-cpu-request|etcd.cpuRequest|
|--etcd-memory-request|etcd.memoryRequest|
|--block-cache-size|pachd.blockCacheBytes|
|--cluster-deployment-id|pachd.clusterDeploymentID|
|--pachd-cpu-request|pachd.cpuRequest|
|inverse of --no-expose-docker-socket|pachd.exposeDockerSocket|
|--expose-object-api|pachd.exposeObjectAPI|
|--pachd-memory-request|pachd.memoryRequest|
|--require-critical-servers-only|pachd.requireCriticalServersOnly|
|--put-file-concurrency-limit|pachd.storage.putFileConcurrencyLimit|
|--upload-concurrency-limit|pachd.storage.uploadConcurrencyLimit|
|--shards|pachd.numShards|
|--cloudfront-distribution|pachd.storage.amazon.cloudFrontDistribution|
|--disable-ssl|pachd.storage.amazon.disableSSL|
|--iam-role|pachd.storage.amazon.iamRole|
|--credentials|pachd.storage.amazon.id|
|| pachd.storage.amazon.secret|
|| pachd.storage.amazon.token|
|--obj-log-options|pachd.storage.amazon.logOptions|
|--max-upload-parts|pachd.storage.amazon.maxUploadParts|
|--no-verify-ssl|pachd.storage.amazon.noVerifySSL|
|--part-size|pachd.storage.amazon.partSize|
|--retries|pachd.storage.amazon.retries|
|--reverse|pachd.storage.amazon.reverse|
|--timeout|pachd.storage.amazon.timeout|
|--upload-acl|pachd.storage.amazon.uploadACL|
|--host-path|pachd.storage.local.hostPath|
|--put-file-concurrency-limit|pachd.storage.minio.putFileConcurrencyLimit|
|--upload-concurrency-limit|pachd.storage.minio.uploadConcurrencyLimit|
|--worker-service-account|pachd.workerServiceAccount.name|
|--no-rbac|opposite of rbac.create|
|--local-roles|rbac.clusterRBAC|

