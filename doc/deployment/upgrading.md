# Pachyderm Version Upgrades

**Important Information For v1.8**

In v1.8 we rearchitected core parts of the platform to [improve speed and scalability](http://www.pachyderm.io/2018/11/15/performance-improvements.html). Currently, this has made upgrading to 1.8 from previous versions of pachyderm not possible. Therefore, it's recommended that users create a new 1.8 deployment and [migrate data over manually](https://pachyderm.readthedocs.io/en/latest/deployment/migrations.html). Rest assured, we are working on a supported upgrade path and you can track our efforts [here](https://github.com/pachyderm/pachyderm/issues/3259).  

Pachyderm releases new major versions (1.4, 1.5, 1.6, etc.) roughly every 2-3 months and releases minor versions as needed/warranted. Upgrading the version of your Pachyderm cluster should be relatively painless, and you should try to upgrade to make sure that you benefit from the latest features, bug fixes, etc. This guide will walk you through that upgrading process.

**Note** - Occasionally, Pachyderm introduces changes that are backward-incompatible. For example, repos/commits/files created on an old version of Pachyderm may be unusable on a new version of Pachyderm. When that happens (which isn't very often), migrations of Pachyderm metadata will happen automatically upon upgrading. We try our best to make these type of changes transparent (in blog posts, changelogs, etc.), and you can read more about the migration process and best practices [here](migrations.html). 

## Before Upgrading

Pachyderm's state (the data you have/are processing and metadata associated with commits, jobs, etc.) is stored in the object store bucket and persistent volume(s) you specified at deploy time. As such, you may want to back up one or both of these storage resources before upgrading, just in case something unexpected happens. You should follow your cloud provider's recommendation for backing up these resources. For example, here are official guides on backing up persistent volumes on Google Cloud Platform and AWS, respectively:

- [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Creating snapshots of Elastic Block Store (EBS) volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

In addition or alternatively, you can utilize `pachctl extract` and `pachctl restore` to extract the state of a Pachyderm cluster and restore a Pachyderm cluster to an extracted state. This process is further described [here](migrations.html#backups).

That being said, the upgrading steps detailed below should not effect these storage resources, and it's perfectly fine to upgrade to the new version of Pachyderm with the same storage resources.

It's also good idea to version or otherwise save the Pachyderm deploy commands (`pachctl deploy ...`) that you utilize when deploying, because you can re-use those exact same commands when re-deploying, as further detailed below.

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

Once you install the new version of `pachctl` (e.g., 1.7.0 in our example), you can confirm this via:

```sh
$ pachctl version --client-only
COMPONENT           VERSION
pachctl             1.7.0
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
pachctl             1.7.0
pachd               1.7.0
```

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

When you upgrade Pachyderm versions, you may lose your local `port-forward` to connect `pachctl` to your cluster. Alternatively, if you are setting the `ADDRESS` environmental variable manually to connect `pachctl` to your cluster, the IP address for Pachyderm may have changed. To fix this, you can either:

- Re-run `pachctl port-forward &`, or
- Set the `ADDRESS` environmental variable to the update value, e.g., `export ADDRESS=<k8s master IP>:30650`.

