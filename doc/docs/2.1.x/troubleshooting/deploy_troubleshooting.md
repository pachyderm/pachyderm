# Troubleshooting Deployments

A common issue related to a deployment: getting a `CrashLoopBackoff` error. 

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

If the error / recourse isn't obvious from the error message, post the error as well as the `pachd` logs in our [Slack channel](https://www.pachyderm.com/slack/){target=_blank}, or open a [GitHub Issue](https://github.com/pachyderm/pachyderm/issues/new){target=_blank} and provide the necessary details prompted by the issue template. Please do be sure provide these logs either way as it is extremely helpful in resolving the issue.

### Pod stuck in `CrashLoopBackoff` - with error attaching volume

#### Symptoms

A pod (could be the `pachd` pod or a worker pod) fails to startup, and is stuck in `CrashLoopBackoff`. If you execute `kubectl describe po/pachd-xxxx`, you'll see an error message like the following at the bottom of the output:

```
  30s        30s        1    {attachdetach }                Warning        FailedMount    Failed to attach volume "etcd-volume" on node "ip-172-20-44-17.us-west-2.compute.internal" with: Error attaching EBS volume "vol-0c1d403ac05096dfe" to instance "i-0a12e00c0f3fb047d": VolumeInUse: vol-0c1d403ac05096dfe is already attached to an instance
```

This would indicate that the [persistent volume claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/){target=_blank} is failing to get attached to the node in your kubernetes cluster.  

#### Recourse

Your best bet is to manually detach the volume and restart the pod.  

For example, to resolve this issue when Pachyderm is deployed to AWS, pull up your AWS web console and look up the node mentioned in the error message (ip-172-20-44-17.us-west-2.compute.internal in our case). Then on the bottom pane for the attached volume. Follow the link to the attached volume, and detach the volume. You may need to "Force Detach" it.

Once it's detached (and marked as available). Restart the pod by killing it, e.g:

```
kubectl delete po/pachd-xxx
```

It will take a moment for a new pod to get scheduled.
