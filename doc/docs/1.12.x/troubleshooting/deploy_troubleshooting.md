# Troubleshooting Deployments

Here are some common issues by symptom related to certain deploys. 

- [General Pachyderm cluster deployment](#general-pachyderm-cluster-deployment)
- Environment-specific
  - [AWS](#aws-deployment)
    - [Can't connect to the Pachyderm cluster after a rolling update](#cant-connect-to-the-pachyderm-cluster-after-a-rolling-update)
    - [The one shot deploy script, `aws.sh`, never completes](#one-shot-script-never-completes)
    - [VPC limit exceeded](#vpc-limit-exceeded)
    - [GPU node never appears](#gpu-node-never-appears)

<!--  - Google - coming soon...
  - Azure - coming soon...-->

## General Pachyderm cluster deployment

- [Pod stuck in `CrashLoopBackoff`](#pod-stuck-in-crashloopbackoff)
- [Pod stuck in `CrashLoopBackoff` - with error attaching volume](#pod-stuck-in-crashloopbackoff-with-error-attaching-volume)

### Pod stuck in `CrashLoopBackoff`

#### Symptoms

The pachd pod keeps crashing/restarting:

```
kubectl get all
NAME                        READY     STATUS             RESTARTS   AGE
po/etcd-281005231-qlkzw     1/1       Running            0          7m
po/pachd-1333950811-0sm1p   0/1       CrashLoopBackOff   6          7m

NAME             CLUSTER-IP       EXTERNAL-IP   PORT(S)                       AGE
svc/etcd         100.70.40.162    <nodes>       2379:30938/TCP                7m
svc/kubernetes   100.64.0.1       <none>        443/TCP                       9m
svc/pachd        100.70.227.151   <nodes>       650:30650/TCP,651:30651/TCP   7m

NAME           DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
deploy/etcd    1         1         1            1           7m
deploy/pachd   1         1         1            0           7m

NAME                  DESIRED   CURRENT   READY     AGE
rs/etcd-281005231     1         1         1         7m
rs/pachd-1333950811   1         1         0         7m
```

#### Recourse

First describe the pod:

```
kubectl describe po/pachd-1333950811-0sm1p
```

If you see an error including `Error attaching EBS volume` or similar, see the recourse for that error here under the corresponding section below. If you don't see that error, but do see something like:

```
  1m    3s    9    {kubelet ip-172-20-48-123.us-west-2.compute.internal}                Warning    FailedSync    Error syncing pod, skipping: failed to "StartContainer" for "pachd" with CrashLoopBackOff: "Back-off 2m40s restarting failed container=pachd pod=pachd-1333950811-0sm1p_default(a92b6665-506a-11e7-8e07-02e3d74c49ac)"
```

it means Kubernetes tried running `pachd`, but `pachd` generated an internal error. To see the specifics of this internal error, check the logs for the `pachd` pod:

```
kubectl logs po/pachd-1333950811-0sm1p
```

!!! note
    If you're using a log aggregator service (e.g. the default in GKE), you won't see any logs when using `kubectl logs ...` in this way.  You will need to look at your logs UI (e.g. in GKE's case the stackdriver console).

These logs will most likely reveal the issue directly, or at the very least, a good indicator as to what's causing the problem. For example, you might see, `BucketRegionError: incorrect region, the bucket is not in 'us-west-2' region`. In that case, your object store bucket in a different region than your pachyderm cluster and the fix would be to recreate the bucket in the same region as your pachydermm cluster.

If the error / recourse isn't obvious from the error message, post the error as well as the `pachd` logs in our [Slack channel](https://www.pachyderm.com/slack/), or open a [GitHub Issue](https://github.com/pachyderm/pachyderm/issues/new) and provide the necessary details prompted by the issue template. Please do be sure provide these logs either way as it is extremely helpful in resolving the issue.

### Pod stuck in `CrashLoopBackoff` - with error attaching volume

#### Symptoms

A pod (could be the `pachd` pod or a worker pod) fails to startup, and is stuck in `CrashLoopBackoff`. If you execute `kubectl describe po/pachd-xxxx`, you'll see an error message like the following at the bottom of the output:

```
  30s        30s        1    {attachdetach }                Warning        FailedMount    Failed to attach volume "etcd-volume" on node "ip-172-20-44-17.us-west-2.compute.internal" with: Error attaching EBS volume "vol-0c1d403ac05096dfe" to instance "i-0a12e00c0f3fb047d": VolumeInUse: vol-0c1d403ac05096dfe is already attached to an instance
```

This would indicate that the [peristent volume claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) is failing to get attached to the node in your kubernetes cluster.  

#### Recourse

Your best bet is to manually detach the volume and restart the pod.  

For example, to resolve this issue when Pachyderm is deployed to AWS, pull up your AWS web console and look up the node mentioned in the error message (ip-172-20-44-17.us-west-2.compute.internal in our case). Then on the bottom pane for the attached volume. Follow the link to the attached volume, and detach the volume. You may need to "Force Detach" it.

Once it's detached (and marked as available). Restart the pod by killing it, e.g:

```
kubectl delete po/pachd-xxx
```

It will take a moment for a new pod to get scheduled.

---

## AWS Deployment

### Can't connect to the Pachyderm cluster after a rolling update

#### Symptom

After running `kops rolling-update`, `kubectl` (and/or `pachctl`) all requests hang and you can't connect to the cluster.

#### Recourse

First get your cluster name. You can easily locate that information by running `kops get clusters`. If you used the one shot deployment](http://docs.pachyderm.io/en/latest/deployment/amazon_web_services.html#one-shot-script), you can also get this info in the deploy logs you created by `aws.sh`.

Then you'll need to grab the new public IP address of your master node. The master node will be named something like `master-us-west-2a.masters.somerandomstring.kubernetes.com`

Update the etc hosts entry in `/etc/hosts` such that the api endpoint reflects the new IP, e.g:

```
54.178.87.68 api.somerandomstring.kubernetes.com
```

### One shot script never completes

#### Symptom

The `aws.sh` one shot deploy script hangs on the line:

```
Retrieving ec2 instance list to get k8s master domain name (may take a minute)
```

If it's been more than 10 minutes, there's likely an error.

#### Recourse

Check the AWS web console / autoscale group / activity history. You have probably hit an instance limit. To confirm, open the AWS web console for EC2 and check to see if you have any instances with names like:

```
master-us-west-2a.masters.tfgpu.kubernetes.com
nodes.tfgpu.kubernetes.com
```

If you don't see instances similar to the ones above the next thing to do is to navigate to "Auto Scaling Groups" in the left hand menu. Then find the ASG with your cluster name:

```
master-us-west-2a.masters.tfgpu.kubernetes.com
```

Look at the "Activity History" in the lower pane. More than likely, you'll see a "Failed" error message describing why it failed to provision the VM. You're probably run into an instance limit for your account for this region. If you're spinning up a GPU node, make sure that your region supports the instance type you're trying to spin up.

A successful provisioning message looks like:

```
Successful
Launching a new EC2 instance: i-03422f3d32658e90c
2017 June 13 10:19:29 UTC-7
2017 June 13 10:20:33 UTC-7
Description:DescriptionLaunching a new EC2 instance: i-03422f3d32658e90c
Cause:CauseAt 2017-06-13T17:19:15Z a user request created an AutoScalingGroup changing the desired capacity from 0 to 1. At 2017-06-13T17:19:28Z an instance was started in response to a difference between desired and actual capacity, increasing the capacity from 0 to 1.
```

While a failed one looks like:

```
Failed
Launching a new EC2 instance
2017 June 12 13:21:49 UTC-7
2017 June 12 13:21:49 UTC-7
Description:DescriptionLaunching a new EC2 instance. Status Reason: You have requested more instances (1) than your current instance limit of 0 allows for the specified instance type. Please visit http://aws.amazon.com/contact-us/ec2-request to request an adjustment to this limit. Launching EC2 instance failed.
Cause:CauseAt 2017-06-12T20:21:47Z an instance was started in response to a difference between desired and actual capacity, increasing the capacity from 0 to 1.
```
### VPC Limit Exceeded

#### Symptom

When running `aws.sh` or otherwise deploying with `kops`, you will see:

```
W0426 17:28:10.435315   26463 executor.go:109] error running task "VPC/5120cf0c-pachydermcluster.kubernetes.com" (3s remaining to succeed): error creating VPC: VpcLimitExceeded: The  maximum number of VPCs has been reached.
```

#### Recourse

You'll need to increase your VPC limit or delete some existing VPCs that are not in use. On the AWS web console navigate to the VPC service. Make sure you're in the same region where you're attempting to deploy.

It's not uncommon (depending on how you tear down clusters) for the VPCs not to be deleted. You'll see a list of VPCs here with cluster names, e.g. `aee6b566-pachydermcluster.kubernetes.com`. For clusters that you know are no longer in use, you can delete the VPC here.

### GPU Node Never Appears

#### Symptom

After running `kops edit ig gpunodes` and `kops update` (as outlined [here](http://docs.pachyderm.io/en/latest/cookbook/gpus.html)) the GPU node never appears, which can be confirmed via the AWS web console.

#### Recourse

It's likely you have hit an instance limit for the GPU instance type you're using, or it's possible that AWS doesn't support that instance type in the current region.

[Follow these instructions to check for and update Instance Limits](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html).  If this region doesn't support your instance type, you'll see an error message like:

```
Failed
Launching a new EC2 instance
2017 June 12 13:21:49 UTC-7
2017 June 12 13:21:49 UTC-7
Description:DescriptionLaunching a new EC2 instance. Status Reason: You have requested more instances (1) than your current instance limit of 0 allows for the specified instance type. Please visit http://aws.amazon.com/contact-us/ec2-request to request an adjustment to this limit. Launching EC2 instance failed.
Cause:CauseAt 2017-06-12T20:21:47Z an instance was started in response to a difference between desired and actual capacity, increasing the capacity from 0 to 1.
```
