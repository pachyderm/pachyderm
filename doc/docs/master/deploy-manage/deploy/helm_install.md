# Helm Deployment

Currently, the `pachctl deploy` command is the authoritative deployment method for Pachyderm.
However, you can now deploy Pachyderm using the package manager [Helm](https://helm.sh/docs/intro/install/#helm).

!!! Note
      Helm support for Pachyderm is a **beta** release. 
      See our [supported releases documentation](https://docs.pachyderm.com/latest/contributing/supported-releases/#release-status) for details on our product release life cycle.

This page gives a you a quick overview of the steps to follow to "helm install" Pachyderm.

As a general rules, those steps would be:

## Prerequisites
1. Install [`Helm`](https://helm.sh/docs/intro/install/). 

1. Follow all installation prerequisites, Kubernetes deployment instructions, and kubectl installation applying to the cloud provider   of your choice:
    Choose the [guidelines](https://docs.pachyderm.com/latest/deploy-manage/deploy/) that apply to you.
    
    For example, if your Cloud provider is Google Cloud Platform, look at the **Prerequisites** and **Deploy Kubernetes** sections for a [deployment on Google Cloud Platform](https://docs.pachyderm.com/latest/deploy-manage/deploy/google_cloud_platform/#google-cloud-platform).

1. The Deployment page that applies to your Cloud provider (or Custom deployment, or On premises deployment) will help you define the various parameters that relate to your own deployment use case.

    In the case of a deployment using `pachctl deploy`, the values of those parameters would ultimately be passed to the 
    corresponding arguments and flags of the command.

    !!! "Example of Google generic pachctl deploy command:"
            ```shell
            pachctl deploy google <bucket-name> <disk-size> [<credentials-file>] [flags]
            ```
            This command comes with an exhaustive [list of available flags](https://docs.pachyderm.com/latest/reference/pachctl/pachctl_deploy_google/) allowing to personalize this deployment.
    
    Those same parameters values will now populate a .yaml configuration file as follow.

## Edit a values.yaml file
Create a personalized `my_pachyderm_values.yaml` out of the [reference values.yaml](https://github.com/pachyderm/helmchart/blob/master/pachyderm/values.yaml) and update the relevant fields according to the parameters gathered in th eprevious step.   

Here is a conversion table that should help you pass easily from the `pachctl deploy` flags to their values.yaml counterpart:

| FLAG OPTION | Values.yaml ATTRIBUTE |
|-------------|-----------------------|

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