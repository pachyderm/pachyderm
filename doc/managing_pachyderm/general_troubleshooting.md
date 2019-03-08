# General Troubleshooting

Here are some common issues by symptom along with steps to resolve them. 

- [Connecting to a Pachyderm cluster](#connecting-to-a-pachyderm-cluster)
  - [Cannot connect via `pachctl` - context deadline exceeded](#cannot-connect-via-pachctl-context-deadline-exceeded)
  - [Certificate error when using `kubectl`](#certificate-error-when-using-kubectl)
  - [Uploads/downloads are slow](#uploads-downloads-are-slow)


---

## Connecting to a Pachyderm Cluster

### Cannot connect via `pachctl` - context deadline exceeded

#### Symptom

You may be using the environment variable `PACHD_ADDRESS` to specify how `pachctl` talks to your Pachyderm cluster, or you may be forwarding the pachyderm port.  In any event, you might see something similar to:

```
$ echo $PACHD_ADDRESS
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

This can happen on any request using `kubectl` (e.g. `kubectl get all`). In particular you'll see:

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

If you do not explicitly set the `PACHD_ADDRESS` environment variable, `pachctl` will default to using port forwarding, which throttles traffic to ~1MB/s. If you need to do large downloads/uploads you should consider using the `PACHD_ADDRESS` variable instead to connect directly to your k8s master node. You'll also want to make sure you've allowed ingress access through any firewalls to your k8s cluster.

