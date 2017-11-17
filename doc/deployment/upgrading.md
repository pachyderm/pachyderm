# Pachyderm Version Upgrades

Pachyderm releases new major versions (1.4, 1.5, 1.6, etc.) roughly every 2-3 months and releases minor versions as needed/warranted. Upgrading the version of your Pachyderm cluster should be relatively painless, and you should try to upgrade to make sure that you benefit from the latest features, bug fixes, etc. This guide will walk you through that upgrading process.

**Note** - Occasionally, Pachyderm introduces changes that are backward-incompatible. For example, repos/commits/files created on an old version of Pachyderm may be unusable on a new version of Pachyderm. When that happens (which isn't very often), we try our best to make this transparent (in blog posts, changelogs, etc.), and we write a migration script that "upgrades" your data so it’s usable by the new version of Pachyderm. More details on Pachyderm migrations can be found [here](migrations.html).

## Before Upgrading

Pachyderm's state (the data you have/are processing and metadata associated with commits, jobs, etc.) is stored in the object store bucket and persistent volume(s) you specified at deploy time. As such, you may want to back up one or both of these storage resources before upgrading, just in case something unexpected happens. You should follow your cloud provider's recommendation for backing up these resources. For example, here are official guides on backing up persistent volumes on Google Cloud Platform and AWS, respectively:

- [Creating snapshots of GCE persistent volumes](https://cloud.google.com/compute/docs/disks/create-snapshots)
- [Creating snapshots of Elastic Block Store (EBS) volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-creating-snapshot.html)

That being said, the upgrading steps detailed below should not effect these storage resources, and it's perfectly fine to re-deploy the new version of Pachyderm with the same storage resources.

It's also good idea to version or otherwise save the Pachyderm deploy commands (`pachctl deploy ...`) that you utilize when deploying, because you can re-use those exact same commands when re-deploying, as further detailed below.

## Upgrading Pachyderm

Upgrading your Pachyderm version is as easy as:

1. [Undeploying Pachyderm](#undeploying-pachyderm)
2. [Upgrading `pachctl`](#upgrading-pachctl)
3. [Re-deploying Pachyderm](#re-deploying-pachyderm)

### Undeploying Pachyderm

Let's suppose that I have a 1.6.3 Pachyderm cluster deployed, which can be confirmed as follows:

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.6.3
pachd               1.6.3
```

To undeploy this Pachyderm cluster, run:

```sh
$ pachctl undeploy
No resources found
service "etcd" deleted
service "pachd" deleted
No resources found
deployment "etcd" deleted
deployment "pachd" deletedserviceaccount "pachyderm" deleted
secret "pachyderm-storage-secret" deleted
No resources found
```

You will see output indicating the Kubernetes services and resources that are being removed, and you may see some `No resources found` statements depending on how you have Pachyderm deployed. After a few minutes, you should see that your Kubernetes cluster has returned to it's pre-Pachyderm state:

```sh
$ kubectl get all
NAME             CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
svc/kubernetes   10.96.0.1    <none>        443/TCP   39m
```

### Upgrading `pachctl`

Now, to deploy an upgraded Pachyderm, we need to retrieve the latest version of `pachctl`. Details on installing the latest version can be found [here](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html#pachctl). You should be able to upgrade via `brew` or `apt` depending on your environment.

Once you install the new version of `pachctl` (e.g., 1.6.4 in our example), you can confirm this via:

```sh
$ pachctl version
COMPONENT           VERSION
pachctl             1.6.4
context deadline exceeded
```

You will see `context deadline exceeded` here because Pachyderm isn't yet running on our cluster.

### Re-deploying Pachyderm

You can now re-deploy Pachyderm with the **same** deploy command that you originally used to deploy Pachyderm. That is, you should specify the same arguments, fields, and storage resources that you specified when deploying the previously utilized version of Pachyderm. The various deploy options/commands are further detailed [here](deploy_intro.html). However, it should look something like:

```sh
$ pachctl deploy <args>
serviceaccount "pachyderm" created
deployment "etcd" created
service "etcd" created
service "pachd" created
deployment "pachd" created
secret "pachyderm-storage-secret" created

Pachyderm is launching. Check it's status with "kubectl get all"
```

After a few minutes, you should then see a healthy Pachyderm cluster running in Kubernetes:

```sh
kubectl get all
NAME                        READY     STATUS    RESTARTS   AGE
po/etcd-1490304484-rm57h    1/1       Running   0          1m
po/pachd-3081814520-15sj5   1/1       Running   0          1m

NAME             CLUSTER-IP      EXTERNAL-IP   PORT(S)                                                   AGE
svc/etcd         10.105.75.160   <nodes>       2379:32379/TCP                                            1m
svc/kubernetes   10.96.0.1       <none>        443/TCP                                                   53m
svc/pachd        10.103.202.58   <nodes>       650:30650/TCP,651:30651/TCP,652:30652/TCP,999:30999/TCP   1m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           1m
deploy/pachd   1         1         1            1           1m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-1490304484    1         1         1         1m
rs/pachd-3081814520   1         1         1         1m
```

And you can confirm the new version of Pachyderm as follows:

```sh
pachctl version
COMPONENT           VERSION
pachctl             1.6.4
pachd               1.6.4
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

