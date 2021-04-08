# Upgrade Pachyderm

!!! Info
      - If you need to upgrade Pachyderm between major versions,
      such as from `1.12.2` to `2.0.0`, follow the
      instructions in the [Migrate between major versions](./migrations.md).
      - Prior to 1.11, minor releases required a migration. This is no longer the case.

Upgrades between minor releases or point releases, such as from version `1.12.5` to
version `1.13.0` do not introduce breaking changes. Therefore, the upgrade
procedure is simple and requires little to no downtime.

!!! Warning
    Do not use these steps to upgrade between major versions because
    it might result in data corruption.

To upgrade Pachyderm from one minor release to another, complete the following steps:

1. Back up your cluster as described in the [Backup and Restore](../backup_restore/#backup-your-cluster)
section.

1. Destroy your Pachyderm cluster:

      ```shell
      pachctl undeploy
      ```

1. Upgrade `pachctl` by using `brew` for macOS or `apt` for Linux:

      **Example:**

      ```shell
      brew upgrade pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}
      ```

      **System response:**

      ```shell
      ==> Upgrading 1 outdated package:
      pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}
      ==> Upgrading pachyderm/tap/pachctl@{{ config.pach_major_minor_version }}
      ...
      ```

      **Note:** You need to specify the major/minor version of `pachctl` to which
      you want to upgrade. For example, if you want to upgrade `1.12.0` to
      the latest point release of the 1.12, add `@1.12` at the end of the upgrade path.

1. Confirm that the new version has been successfully installed by running
the following command:

      ```shell
      pachctl version --client-only
      ```

      **System response:**

      ```shell
      COMPONENT           VERSION
      pachctl             {{ config.pach_latest_version }}
      ```

1. Redeploy Pachyderm by running the `pachctl deploy` command
with the same arguments, fields, and storage resources
that you specified when you deployed the previous version
of Pachyderm:

      ```shell
      pachctl deploy <args>
      ```

      **System response:**

      ```shell
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

      The deployment takes some time. You can run `kubectl get pods` periodically
      to check the status of the deployment. When Pachyderm is deployed, the command
      shows all pods as `READY`:


      ```shell
      kubectl get pods
      ```

      **System response:**

      ```shell
      NAME                     READY     STATUS    RESTARTS   AGE
      dash-482120938-np8cc     2/2       Running   0          4m
      etcd-0                   1/1       Running   0          4m
      pachd-3677268306-9sqm0   1/1       Running   0          4m
      ```

1. Verify that the new version has been deployed:

      ```shell
      pachctl version
      ```

      **System response:**

      ```shell
      COMPONENT           VERSION
      pachctl             {{ config.pach_latest_version }}
      pachd               {{ config.pach_latest_version }}
      ```

      The `pachd` and `pachctl` versions must both match the new version.

## Troubleshooting point release Upgrades

<!-- We might want to move this section to Troubleshooting -->

This section describes issues that you might run into when
upgrading Pachyderm and provides guidelines on how to resolve
them.

### StatefulSets vs static persistent volumes

[StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) are a mechanism provided in Kubernetes 1.9 and newer to manage the deployment and scaling of applications. 
It can use Persistent Volume Provisioning or pre-provisioned PV’s,
both of which are dynamically allocated from Pachyderm's point of view.
Thus, the `--dynamic-etcd-nodes` flag to `pachctl deploy` is used to deploy Pachyderm using StatefulSets.

!!! Tip 
      It is recommended that you deploy Pachyderm using StatefulSets when possible. 
      All of the instructions for cloud provider deployments do this by default.
      We also provide [instructions for on-premises deployments using StatefulSets](../../deploy/on_premises/#statefulsets).

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

   ```shell
   Error from server (AlreadyExists): error when creating "STDIN": secrets "pachyderm-storage-secret" already exists
   ```

This might happen when re-deploying the enterprise dashboard, for example. These warning are benign.

### `pachctl` connnection problems

When you upgrade Pachyderm versions, you may lose your local `port-forward` to connect `pachctl` to your cluster. 
If you are not using `port-forward` and you are instead setting pachd address config value to connect `pachctl` to your cluster, 
the IP address for Pachyderm may have changed. 

To fix problems with connections to `pachd` after upgrading, you can perform the appropriate remedy for your situation:

- Re-run `pachctl port-forward`, or
- Set the pachd address config value to the updated value, e.g.:
 
```shell
pachctl config update context `pachctl config get active-context` --pachd-address=<cluster ip>:30650
```









