# Trouble Shooting Guide

Here we list some common gotchas by symptom and steps you can do to resolve the issue.

## Misc

### Certificate Error When Using Kubectl

This can happen on any request using `kubectl` (e.g. `kubectl get all`) but can also be seen when running `pachctl port-forward` because it uses `kubectl` under the hood.

#### Symptom

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

Update /etc/hosts

### One shot script never completes

#### Symptom


#### Recourse

Check the AWS UI / autoscale group / activity history. You probably hit an instance limit

### VPC Limit Exceeded

#### Symptom

#### Recourse

### GPU Node Never Appears

#### Symptom

#### Recourse



