# Pachyderm Version Upgrades

As new versions of Pachyderm are released, you may need to update your cluster to get access to bug fixes and new features. These updates fall into two categories, [Upgrades](./upgrades.md) and [Migrations](./migrations).

A Pachyderm deployment upgrade is moving between point releases within the same major release (e.g. 1.7.2 --> 1.7.3). Upgrades are typically a simple process that require little to no downtime. 

__ Note:__ If you are moving between major versions (e.g. 1.8.x -->1 .9.x), you need to follow [Migration](./migrations.md) procedures, not upgrading. 

* [Before upgrading](#before-upgrading), we recommend you follow standard backup procedures to ensure no data is lost in the case of an error. 
* [Upgrade Procedures](#upgrading-pachyderm)
* [Common Issues](#common-issues-questions)



## Before Upgrading

Refer to [Backing up your cluster](./backups.md)...

<!-- Pachyderm's state (the data you have/are processing and metadata associated with commits, jobs, etc.) is stored in the object store bucket and persistent volume(s) you specified at deploy time. As such, you may want to back up one or both of these storage resources before upgrading, just in case something unexpected happens. You should follow your cloud provider's recommendation for backing up these resources. For example, here are official guides on backing up persistent volumes on Google Cloud Platform and AWS, respectively:

- [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Creating snapshots of Elastic Block Store (EBS) volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

In addition or alternatively, you can utilize `pachctl extract` and `pachctl restore` to extract the state of a Pachyderm cluster and restore a Pachyderm cluster to an extracted state. This process is further described [here](migrations.html#backups).

That being said, the upgrading steps detailed below should not effect these storage resources, and it's perfectly fine to upgrade to the new version of Pachyderm with the same storage resources.

It's also good idea to version or otherwise save the Pachyderm deploy commands (`pachctl deploy ...`) that you utilize when deploying, because you can re-use those exact same commands when re-deploying, as further detailed below.

**Important Information For v1.8**

In v1.8 we rearchitected core parts of the platform to [improve speed and scalability](http://www.pachyderm.io/2018/11/15/performance-improvements.html). The 1.7.x 1.8.x migration is a fairly indepth process (see our [Migrations docs](./backup_retore_and_migrate.md)) Therefore, it may actually be easier if you don't have tons of data to create a new 1.8 deployment and reingress data. Please come chat with us our [Public Slack Channel](slack.pachyderm.io) if you have any questions. 

Pachyderm releases new major versions (1.4, 1.5, 1.6, etc.) roughly every 3 months and releases point releases as needed/warranted. 

**Note** - Occasionally, Pachyderm introduces changes that are backward-incompatible. For example, repos/commits/files created on an old version of Pachyderm may be unusable on a new version of Pachyderm. When that happens (which isn't very often), migrations of Pachyderm metadata will happen automatically upon upgrading. We try our best to make these type of changes transparent (in blog posts, changelogs, etc.), and you can read more about the migration process and best practices [here](migrations.html).  --> 


## Upgrading Pachyderm

Upgrading your Pachyderm version is as easy as:

1. [Spin down current pachd server](#spin-down-old-cluster)
2. [Upgrading `pachctl`](#upgrading-pachctl)
3. [Re-deploying Pachyderm](#re-deploying-pachyderm)


### Spin down old cluster

```
pachctl undeploy
```

### Upgrading `pachctl`

To deploy an upgraded Pachyderm, we need to retrieve the latest version of `pachctl`. Details on installing the latest version can be found [here](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html#pachctl). You should be able to upgrade via `brew upgrade` or `apt` depending on your environment.

Once you install the new version of `pachctl` (e.g., 1.8.4 in our example), you can confirm this via:

```sh
$ pachctl version --client-only
COMPONENT           VERSION
pachctl             1.8.4
```

### Re-deploying Pachyderm

You can now re-deploy Pachyderm with the **same** deploy command that you originally used to deploy Pachyderm. That is, you should specify the same arguments, fields, and storage resources that you specified when deploying the previously utilized version of Pachyderm. The various deploy options/commands are further detailed [here](deploy_intro.html). However, it should look something like:

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

You'll want to make sure your pachd and pachctl versions are both matching the new version.

## Common Issues, Questions

### Dynamic/static volumes

It is recommended that you deploy Pachyderm using dynamically etcd volumes when possible. If you have deployed Pachyderm using dynamic volumes, you can still use the *same* deploy command to re-deploy Pachyderm (i.e., the one specifying dynamic etcd volumes). Kubernetes is smart enough to see the previously utilized volumes and re-use them.

### etcd re-deploy problems

Depending on the cloud you are deploying to and the previous deployment configuration, we have seen certain cases in which volumes don't get attached to the right nodes on re-deploy (especially when using AWS). In these scenarios, you may see the etcd pod stuck in a Pending, CrashLoopBackoff, or other failed state. Most often, deleting the corresponding etcd pod(s) or nodes (to redeploy them) or re-deploying all of Pachyderm again will fix the issue. 

### `AlreadyExists` errors on re-deploy

Occasionally, you might see errors similar to the following:

```
Error from server (AlreadyExists): error when creating "STDIN": secrets "pachyderm-storage-secret" already exists
```

This might happen when re-deploying the enterprise dashboard, for example. These warning are benign.

### Reconnecting `pachctl`

When you upgrade Pachyderm versions, you may lose your local `port-forward` to connect `pachctl` to your cluster. Alternatively, if you are setting the `PACHD_ADDRESS` environmental variable manually to connect `pachctl` to your cluster, the IP address for Pachyderm may have changed. To fix this, you can either:

- Re-run `pachctl port-forward &`, or
- Set the `PACHD_ADDRESS` environmental variable to the update value, e.g., `export PACHD_ADDRESS=<k8s master IP>:30650`.

