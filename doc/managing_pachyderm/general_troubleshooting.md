# General Troubleshooting

Here are some common issues by symptom along with steps to resolve them.  They are organized into the following categories:

- [Deploying a Pachyderm cluster](#starting-up-a-pachyderm-cluster)
- [Connecting to a Pachyderm cluster](#connecting-to-a-pachyderm-cluster)
- [Problems running pipelines](#problems-running-pipelines)

---

## Deploying A Pachyderm Cluster

- [Pod stuck in `CrashLoopBackoff`](#pod-stuck-in-crashloopbackoff)
- [Pod stuck in `CrashLoopBackoff` - with error attaching volume](#pod-stuck-in-crashloopbackoff-with-error-attaching-volume)

### Pod stuck in `CrashLoopBackoff`

#### Symptoms

The pachd pod keeps crashing/restarting:

```
$ kubectl get all
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
$ kubectl describe po/pachd-1333950811-0sm1p
```

If you see an error including `Error attaching EBS volume` or similar, see the recourse for that error here under the corresponding section below this one.  If you don't see that error, but do see something like:

```
  1m    3s    9    {kubelet ip-172-20-48-123.us-west-2.compute.internal}                Warning    FailedSync    Error syncing pod, skipping: failed to "StartContainer" for "pachd" with CrashLoopBackOff: "Back-off 2m40s restarting failed container=pachd pod=pachd-1333950811-0sm1p_default(a92b6665-506a-11e7-8e07-02e3d74c49ac)"
```

That means Kubernetes tried running `pachd`, but `pachd` generated an internal error. To see the specifics of this internal error, check the logs for the `pachd` pod:

```
$kubectl logs po/pachd-1333950811-0sm1p
```

**Note**: If you're using a log aggregator service (e.g. the default in GKE), you won't see any logs when using `kubectl logs ...` in this way.  You will need to look at your logs UI (e.g. in GKE's case the stackdriver console).

These logs will likely reveal a misconfiguration in your deploy.  For example, you might see, `BucketRegionError: incorrect region, the bucket is not in 'us-west-2' region`.  In that case, you've deployed your bucket in a different region than your cluster.

If the error / recourse isn't obvious from the error message, you can now provide the content of the `pachd` logs when getting help in our Slack channel or by opening a [GitHub Issue](github.com/pachyderm/pachyderm/issues/new). Please provide these logs either way as it is extremely helpful in resolving the issue..

### Pod stuck in `CrashLoopBackoff` - with error attaching volume

#### Symptoms

A pod (could be the `pachd` pod or a worker pod) fails to startup, and is stuck in `CrashLoopBackoff`. If you execute `kubectl describe po/pachd-xxxx`, you'll see an error message like the following at the bottom of the output:

```
  30s        30s        1    {attachdetach }                Warning        FailedMount    Failed to attach volume "etcd-volume" on node "ip-172-20-44-17.us-west-2.compute.internal" with: Error attaching EBS volume "vol-0c1d403ac05096dfe" to instance "i-0a12e00c0f3fb047d": VolumeInUse: vol-0c1d403ac05096dfe is already attached to an instance
```

#### Recourse

Your best bet is to manually detach the volume and restart the pod.  

For example, to resolve this issue when Pachyderm is deployed to AWS, first find the node of which the pod is scheduled. In the output of the `kubectl describe po/pachd-xxx` command above, you should see the name of the node on which the pod is running.  In the AWS web console, find that node.. Once you have the right node, look in the bottom pane for the attached volume. Follow the link to the attached volume, and detach the volume. You may need to "Force Detach" it.

Once it's detached (and marked as available). Restart the pod by killing it, e.g:

```
$kubectl delete po/pachd-xxx
```

It will take a moment for a new pod to get scheduled.

---

## Connecting to a Pachyderm Cluster

- [Cannot connect via `pachctl` - context deadline exceeded](#cannot-connect-via-pachctl-context-deadline-exceeded)
- [Certificate error when using `kubectl`](#certificate-error-when-using-kubectl)
- [Uploads/downloads are slow](#uploads-downloads-are-slow)

### Cannot connect via `pachctl` - context deadline exceeded

#### Symptom

You may be using the environmental variable `ADDRESS` to specify how `pachctl` talks to your Pachyderm cluster, or you may be forwarding the pachyderm port via `pachctl port-forward`.  In any event, you might see something similar to:

```
$ echo $ADDRESS
1.2.3.4:30650
$ pachctl version
COMPONENT           VERSION                                          
pachctl             1.4.8   
context deadline exceeded
```

#### Recourse

It's possible that the connection is just taking a while. Occasionally this can happen if your cluster is far away (deployed in a region across the country). Check your internet connection.

It's also possible that you haven't poked a hole in the firewall to access the node on this port. Usually to do that you adjust a security rule (in AWS parlance a security group). For example, on AWS, if you find your node in the web console and click on it, you should see a link to the associated security group. Inspect that group. There should be a way to "add a rule" to the group. You'll want to enable TCP access (ingress) on port 30650. You'll usually be asked which incoming IPs should be whitelisted. You can choose to use your own, or enable it for everyone (0.0.0.0/0).


### Certificate Error When Using Kubectl

#### Symptom

This can happen on any request using `kubectl` (e.g. `kubectl get all`), but it can also be seen when running `pachctl port-forward` because it uses `kubectl` under the hood. In particular you'll see:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.4", GitCommit:"d6f433224538d4f9ca2f7ae19b252e6fcb66a3ae", GitTreeState:"clean", BuildDate:"2017-05-19T20:41:24Z", GoVersion:"go1.8.1", Compiler:"gc", Platform:"darwin/amd64"}
Unable to connect to the server: x509: certificate signed by unknown authority
```

#### Recourse

Check if you're on any sort of VPN or other egress proxy that would break SSL.  Also, there is a possibility that your credentials have expired. In the case where you're using GKE and gcloud, renew your credentials via:

```
$ kubectl get all
Unable to connect to the server: x509: certificate signed by unknown authority
$ gcloud container clusters get-credentials my-cluster-name-dev
Fetching cluster endpoint and auth data.
kubeconfig entry generated for my-cluster-name-dev.
$ kubectl config current-context
gke_my-org_us-east1-b_my-cluster-name-dev
```

### Uploads/Downloads are Slow

#### Symptom

Any `pachctl put-file` or `pachctl get-file` commands are slow.

#### Recourse

Check if you're using port-forwarding. Port forwarding throttles traffic to ~1MB/s. If you need to do large downloads/uploads you should consider using the `ADDRESS` variable instead to connect directly to your k8s master node. [See this note](./getting_started/other_installation.html?highlight=ADDRESS#usage)

You'll also want to make sure you've allowed ingress access through any firewalls to your k8s cluster.

---

## Problems Running Pipelines

### All your pods / jobs get evicted

#### Symptom

Running:

```
$ kubectl get all
```

shows a bunch of pods that are marked `Evicted`. If you `kubectl describe ...` one of those evicted pods, you see an error saying that it was evicted due to disk pressure.


#### Recourse

Your nodes are not configured with a big enough root volume size.  You need to make sure that each node's root volume is big enough to store the biggest datum you expect to process anywhere on your DAG plus the size of the output files that will be written for that datum.

Let's say you have a repo with 100 folders. You have a single pipeline with this repo as an input, and the glob pattern is `/*`. That means each folder will be processed as a single datum. If the biggest folder is 50GB and your pipeline's output is about 3 times as big, then your root volume size needs to be bigger than:

```
50 GB (to accommodate the input) + 50 GB x 3 (to accommodate the output) = 200GB
```

In this case we would recommend 250GB to be safe. If your root volume size is less than 50GB (many defaults are 20GB), this pipeline will fail when downloading the input. The pod may get evicted and rescheduled to a different node, where the same thing will happen.

### Pipeline Exists But Never Runs

#### Symptom

You can see the pipeline via:

```
$ pachctl list-pipeline
```

But if you look at the job via:

```
$ pachctl list-job
```

It's marked as running with `0/0` datums having been processed.  If you inspect the job via:

```
$ pachctl inspect-job
```

You don't see any worker set. E.g:

```
Worker Status:
WORKER              JOB                 DATUM               STARTED             
...
```

If you do `kubectl get pod` you see the worker pod for your pipeline, e.g:

```
po/pipeline-foo-5-v1-273zc
```

But it's state is `Pending` or `CrashLoopBackoff`.

#### Recourse

First make sure that there is no parent job still running. Do `pachctl list-job | grep yourPipelineName` to see if there are pending jobs on this pipeline that were kicked off prior to your job. A parent job is the job that corresponds to the parent output commit of this pipeline. A job will block until all parent jobs complete.

If there are no parent jobs that are still running, then continue debugging:

Describe the pod via:

```
$kubectl describe po/pipeline-foo-5-v1-273zc
```

If the state is `CrashLoopBackoff`, you're looking for a descriptive error message. One such cause for this behavior might be if you specified an image for your pipeline that does not exist.

If the state is `Pending` it's likely the cluster doesn't have enough resources. In this case, you'll see a `could not schedule` type of error message which should describe which resource you're low on. This is more likely to happen if you've set resource requests (cpu/mem/gpu) for your pipelines.  In this case, you'll just need to scale up your resources. If you deployed using `kops`, you'll want to do edit the instance group, e.g. `kops edit ig nodes ...` and up the number of nodes. If you didn't use `kops` to deploy, you can use your cloud provider's auto scaling groups to increase the size of your instance group. Either way, it can take up to 10 minutes for the changes to go into effect. 

You can read more about autoscaling [here](../cookbook/autoscaling.html)
