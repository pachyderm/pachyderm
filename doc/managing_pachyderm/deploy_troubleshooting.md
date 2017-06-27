# Deploy Specific Troubleshooting

Here are some common issues by symptom related to certain deploys.  They are organized into the following categories:

- [AWS](#aws-deployment)
- Google - coming soon...
- Azure - coming soon...

---

## AWS Deployment

- [Can't connec to to the Pachyderm cluster after a rolling update](#cant-connect-to-the-Pachyderm-cluster-after-a-rolling-update)
- [The one shot deploy script, `aws.sh`, never completes](#one-shot-script-never-completes)
- [VPC limit exceeded](#vpc-limit-exceeded)
- [GPU node never appears](#gpu-node-never-appears)

### Can't connect to the Pachyderm cluster after a rolling update

#### Symptom

After running `kops rolling-update`, `kubectl` (and/or `pachctl`) cannot connect to the cluster. All `kubectl` requests hang.

#### Recourse

First get your cluster name. This will be in the deploy logs you saved from running `aws.sh` (if you utilized the [one shot deployment](http://docs.pachyderm.io/en/latest/deployment/amazon_web_services.html#one-shot-script)), or can be retrieved via `kops get clusters`.

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

Check the AWS web console / autoscale group / activity history. You have probably hit an instance limit.  To navigate there, open the AWS web console for EC2. Check to see if you have any instances with names like::

```
master-us-west-2a.masters.tfgpu.kubernetes.com
nodes.tfgpu.kubernetes.com
```

If not, navigate to "Auto Scaling Groups" in the left hand menu. Then find the ASG with your cluster name:

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

After running `kops edit ig gpunodes` and `kops update` (as outlined [here](http://docs.pachyderm.io/en/latest/cookbook/gpus.html)) the GPU node never appears, which can be confirmed via the AWS web console..

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
