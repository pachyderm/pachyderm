# Trouble Shooting Guide

Here we list some common gotchas by symptom and steps you can do to resolve the issue.

- Connecting pachctl to the cluster
- Starting up Pachyderm cluster
- AWS Deployment
- Miscellaneous

---





## Misc

### Pod failed to attach volume

#### Symptoms

A pod (could be the pachd pod or a worker pod) fails to startup, and is stuck in `CrashLoopBackoff`. If you do `kubectl describe po/pachd-xxxx` you'll see an error message like the following at the bottom of the output:

```
  30s        30s        1    {attachdetach }                Warning        FailedMount    Failed to attach volume "etcd-volume" on node "ip-172-20-44-17.us-west-2.compute.internal" with: Error attaching EBS volume "vol-0c1d403ac05096dfe" to instance "i-0a12e00c0f3fb047d": VolumeInUse: vol-0c1d403ac05096dfe is already attached to an instance
```

#### Recourse

Your best bet is to manually detach the volume and restart the pod.

To find the volume, you first need the node. In the output of the `kubectl describe po/pachd-xxx` command above, the output should include the name of the node it's running on.

On the AWS UI, find that node the name provided in the step above is the internal DNS value of the node in the AWS UI. Once you have the right node, look in the bottom pane for the attached volume. Follow that link, and detach the volume. You may need to 'Force Detach' it.

Once it's detached (and marked as available). Restart the pod by killing it, e.g:

```
$kubectl delete po/pachd-xxx
```

It'll take a moment for a new one to get rescheduled.

### Certificate Error When Using Kubectl

This can happen on any request using `kubectl` (e.g. `kubectl get all`) but can also be seen when running `pachctl port-forward` because it uses `kubectl` under the hood.

#### Symptom

On any `kubectl` command you'll see a certificate error:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.4", GitCommit:"d6f433224538d4f9ca2f7ae19b252e6fcb66a3ae", GitTreeState:"clean", BuildDate:"2017-05-19T20:41:24Z", GoVersion:"go1.8.1", Compiler:"gc", Platform:"darwin/amd64"}
Unable to connect to the server: x509: certificate signed by unknown authority
```

#### Recourse

Check if you're on any sort of VPN or other egress proxy that would break SSL

### Upload/Download is Slow

#### Symptom

Any `pachctl put-file` or `pachctl get-file`s are slow.

#### Recourse

Check if you're using port-forwarding. Port forwarding throttles traffic to ~1MB/s. If you need to do large downloads/uploads you should consider using the `ADDRESS` variable instead to connect directly to your k8s master node. [See this note](./getting_started/other_installation.html?highlight=ADDRESS#usage)

You'll want to make sure you've allowed ingress access through any firewalls to your k8s cluster. (See below)

### Cannot connect to cluster using ADDRESS variable

#### Symptom

```
$echo $ADDRESS
1.2.3.4:30650
$pachctl version
COMPONENT           VERSION                                          
pachctl             1.4.8   
context deadline exceeded
```

#### Recourse

It's possible that you are connecting, it's just taking a while. Occasionally this can happen if your cluster is far away (deployed in a region across the country). Check your internet connection.

It's also possible that you haven't poked a hole in the firewall to access this node on this port. Usually to do that you adjust a security rule (in AWS parlance a security group). First find your node on the UI. Then click it. You should see a link to the associated security group. Inspect that group. There should be a way to 'add a rule' to the group. You'll want to enable TCP access (ingress) on port 30650. You'll usually be asked which incoming IPs should be whitelisted. You can choose to use your own. Or enable it for everyone (0.0.0.0/0).


## AWS Deployment

### Can't connect to cluster after rolling update

#### Symptom

After a `kops rolling-update` kubectl cannot connect to the cluster. All `kubectl` requests hang.

#### Recourse

First get your cluster name. This will be in the deploy logs you saved from running `aws.sh`.

Then you'll need to grab the new public IP address of your master node. The master node will be named something like `master-us-west-2a.masters.somerandomstring.kubernetes.com`

Update the etc hosts entry for the api endpoint, e.g:

```
54.178.87.68 api.somerandomstring.kubernetes.com
```

That might be your old entry. Update the IP address on the left to the value of your master node's public IP address.

Update /etc/hosts

### One shot script never completes

#### Symptom

It hangs on the line:

```
Retrieving ec2 instance list to get k8s master domain name (may take a minute)
```

If it's been more than 10 minutes, there's likely an error.

#### Recourse

Check the AWS UI / autoscale group / activity history. You probably hit an instance limit.

To navigate there, open the AWS UI for EC2. Check to see if you have any instances w your cluster name:

```
master-us-west-2a.masters.tfgpu.kubernetes.com
nodes.tfgpu.kubernetes.com
```

If not, navigate to "Auto Scaling Groups" in the left hand menu. Then find the ASG with your cluster name:

```
master-us-west-2a.masters.tfgpu.kubernetes.com
```

And look at the 'Activity History' in the lower pane. More than likely, you'll see a 'Failed' error message describing why it failed to provision the VM. You're probably run into an instance limit for your account for this region. If you're spinning up a GPU node, make sure that your region supports the instance type you're trying to spin up.

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

```
W0426 17:28:10.435315   26463 executor.go:109] error running task "VPC/5120cf0c-pachydermcluster.kubernetes.com" (3s remaining to succeed): error creating VPC: VpcLimitExceeded: The  maximum number of VPCs has been reached.
```

#### Recourse

You'll need to increase your VPC limit or delete some existing VPCs that are not in use. On the AWS UI navigate to the VPC service. Make sure you're in the same region where you're attempting to deploy.

It's not uncommon (depending on how you tear down clusters) for the VPCs not to be deleted. You'll see a list of VPCs here with cluster names, e.g. `aee6b566-pachydermcluster.kubernetes.com`. For clusters that you know are no longer in use, you can delete the VPC here.

### GPU Node Never Appears

#### Symptom

After running the `kops edit ig gpunodes` and `kops update` the node never appears in the AWS UI.

#### Recourse

It's likely you have hit an instance limit for the gpu instance type you're using, or it's possible that AWS doesn't support that instance type in the current region.

[Follow the instructions to check for Instance Limit Error messages here]()

If this region doesn't support your instance type you'll see an error message like:

```
Failed
Launching a new EC2 instance
2017 June 12 13:21:49 UTC-7
2017 June 12 13:21:49 UTC-7
Description:DescriptionLaunching a new EC2 instance. Status Reason: You have requested more instances (1) than your current instance limit of 0 allows for the specified instance type. Please visit http://aws.amazon.com/contact-us/ec2-request to request an adjustment to this limit. Launching EC2 instance failed.
Cause:CauseAt 2017-06-12T20:21:47Z an instance was started in response to a difference between desired and actual capacity, increasing the capacity from 0 to 1.
```
