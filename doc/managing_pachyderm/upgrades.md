# Upgrades

<!-- This section needs a thorough cleanup-->

- [Introduction](#introduction)
- [General upgrade procedure](#general-upgrade-procedure)
  - [Before you start: backups](#before-you-start-backups)
  - [Migration steps](#migration-steps)
    - [1. Spin down current pachd server](#spin-down-old-cluster)
    - [2. Upgrading `pachctl`](#upgrading-pachctl)
    - [3. Re-deploying Pachyderm](#re-deploying-pachyderm)
- [Common Issues](#common-issues)
  - [StatefulSets vs static persistent volumes](#statefulsets-vs-static-persistent-volumes)
  - [`etcd` re-deploy problems](#etcd-re-deploy-problems)
  - [`AlreadyExists` errors on re-deploy](#alreadyexists-errors-on-re-deploy)
  - [`pachctl` connnection problems](#pachctl-connnection-problems)

These updates fall into two categories, upgrades and migrations.

Migrations involve moving between major releases, 
like 1.8.6 to 1.9.0.
They're covered in a [separate document](./migrations.html).

An upgrade is moving between point releases within the same major release, 
like 1.7.2 to 1.7.3.
Upgrades are typically a simple process that require little to no downtime.
They're covered in this document.

*Important*: Performing an _upgrade_ when going between _major releases_ may lead to corrupted data. 
*You must perform a [migration](./migrations.html) when going between major releases!*

To upgrade your Pachyderm cluster between minor releases, you should follow the steps below

## General upgrade procedure

### Before you start: backups

Please refer to [the documentation on backing up your cluster](./backup_restore.html#general-backup-procedure).

### Upgrade steps

[Back up your cluster](./backups.md) using Pachyderm's recommended procedures.

### 1. Spin down old cluster

```
pachctl undeploy
```

### 2. Upgrading `pachctl`

To deploy a new version of Pachyderm, you need to first upgrade
`pachctl`. Use `brew` for macOS or `apt` for Linux to upgrade
`pachctl` to the latest version.

To upgrade `pachctl`, complete the following steps:

1. Use your favorite package manager to install the latest version
of `pachctl`. Example:

     ```bash
     $ brew upgrade pachyderm/tap/pachctl@1.9
     ==> Upgrading 1 outdated package:
     pachyderm/tap/pachctl@1.9 v1.9.0 -> v1.9.2
     ==> Upgrading pachyderm/tap/pachctl@1.9
     ...
     ```

   **Note:** You need to specify the version of `pachctl` to which
   you want to upgrade. For example, if you want to upgrade `1.9.0` to
   `1.9.2`, add `@1.9` at the end of the upgrade path.

1. Confirm that the new version has been successfully installed by running
the following command:

   ```sh
   $ pachctl version --client-only
   COMPONENT           VERSION
   pachctl             1.9.2
   ```

### 3. Re-deploying Pachyderm

You can now re-deploy Pachyderm with the **same** deploy command that you originally used to deploy Pachyderm.
That is,
you should specify the same arguments, fields, and storage resources that you specified when deploying the previously utilized version of Pachyderm. 
The various deploy options/commands are further detailed [here](deploy_intro.html). 
However, it should look something like:

```sh
$ pachctl deploy <args>
serviceaccount "pachyderm" created
storageclass "etcd-storage-class" created
service "etcd-headless" created
statefulset "etcd" created
service "etcd" created
service "pachd" created
deployment "pachd" created
service "dash" created
deployment "dash" created
secret "pachyderm-storage-secret" created

Pachyderm is launching. Check its status with "kubectl get all"
Once launched, access the dashboard by running "pachctl port-forward"
```

After a few minutes, you should then see a healthy Pachyderm cluster running in Kubernetes:

```sh
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
dash-482120938-np8cc     2/2       Running   0          4m
etcd-0                   1/1       Running   0          4m
pachd-3677268306-9sqm0   1/1       Running   0          4m
```

And you can confirm the new version of Pachyderm as follows:

```sh
pachctl version
COMPONENT           VERSION
pachctl             1.8.4
pachd               1.8.4
```

You'll want to make sure your pachd and pachctl versions both match the new version.

## Common Issues

### StatefulSets vs static persistent volumes

[StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) are a mechanism provided in Kubernetes 1.9 and newer to manage the deployment and scaling of applications. 
It can use Persistent Volume Provisioning or pre-provisioned PV’s,
both of which are dynamically allocated from Pachyderm's point of view.
Thus, the `--dynamic-etcd-nodes` flag to `pachctl deploy` is used to deploy Pachyderm using StatefulSets.

It is recommended that you deploy Pachyderm using StatefulSets when possible. 
All of the instructions for cloud provider deployments do this by default.
We also provide [instructions for on-premises deployments using StatefulSets](http://docs.pachyderm.io/en/latest/deployment/on_premises.html#statefulsets).

If you have deployed Pachyderm using StatefulSets, 
you can still use the *same* deploy command to re-deploy Pachyderm. 
Kubernetes is smart enough to see the previously utilized volumes and re-use them.

### `etcd` re-deploy problems

Depending on the cloud you are deploying to and the previous deployment configuration, 
we have seen certain cases in which volumes don't get attached to the right nodes on re-deploy (especially when using AWS). 
In these scenarios, you may see the `etcd` pod stuck in a `Pending`, `CrashLoopBackoff`, or other failed state. 
Most often, deleting the corresponding `etcd` pod(s) or nodes to redeploy them 
or re-deploying all of Pachyderm again will fix the issue. 

### `AlreadyExists` errors on re-deploy

Occasionally, you might see errors similar to the following:

```
Error from server (AlreadyExists): error when creating "STDIN": secrets "pachyderm-storage-secret" already exists
```

This might happen when re-deploying the enterprise dashboard, for example. These warning are benign.

### `pachctl` connnection problems

When you upgrade Pachyderm versions, you may lose your local `port-forward` to connect `pachctl` to your cluster. 
If you are not using `port-forward` and you are instead setting pachd address config value to connect `pachctl` to your cluster, 
the IP address for Pachyderm may have changed. 

To fix problems with connections to `pachd` after upgrading, you can perform the appropriate remedy for your situation:

- Re-run `pachctl port-forward &`, or
- Set the pachd address config value to the updated value, e.g.: ```pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650```









