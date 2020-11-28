# General Troubleshooting

This section outlines general troubleshooting steps that you can
attempt to repair your cluster.
Here are some common issues by symptom along with steps to resolve them. 

- [Connecting to a Pachyderm cluster](#connecting-to-a-pachyderm-cluster)
  - [Cannot connect via `pachctl` - context deadline exceeded](#cannot-connect-via-pachctl-context-deadline-exceeded)
  - [Certificate error when using `kubectl`](#certificate-error-when-using-kubectl)
  - [Uploads and Downloads are slow](#uploads-and-downloads-are-slow)
  - [Naming a Repo with an Unsupported Symbol](#naming-a-repo-with-an-unsupported-symbol)
  - [Failed Uploads](#failed-uploads)
  
---

## Connecting to a Pachyderm Cluster

### Cannot connect via `pachctl` - context deadline exceeded

#### Symptom

You may be using the pachd address config value or environment variable to specify how `pachctl` talks to your Pachyderm cluster, or you may be forwarding the pachyderm port.  In any event, you might see something similar to:

```
pachctl version
COMPONENT           VERSION                                          
pachctl             {{ config.pach_latest_version }}   
context deadline exceeded
```

Also, you might get this message if `pachd` is not running.

#### Recourse

It's possible that the connection is just taking a while. Occasionally this can happen if your cluster is far away (deployed in a region across the country). Check your internet connection.

It's also possible that you haven't poked a hole in the firewall to access the node on this port. Usually to do that you adjust a security rule (in AWS parlance a security group). For example, on AWS, if you find your node in the web console and click on it, you should see a link to the associated security group. Inspect that group. There should be a way to "add a rule" to the group. You'll want to enable TCP access (ingress) on port 30650. You'll usually be asked which incoming IPs should be whitelisted. You can choose to use your own, or enable it for everyone (0.0.0.0/0).


### Certificate Error When Using Kubectl

#### Symptom

This can happen on any request using `kubectl` (e.g. `kubectl get all`). In particular you'll see:

```
kubectl version
Client Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.4", GitCommit:"d6f433224538d4f9ca2f7ae19b252e6fcb66a3ae", GitTreeState:"clean", BuildDate:"2017-05-19T20:41:24Z", GoVersion:"go1.8.1", Compiler:"gc", Platform:"darwin/amd64"}
Unable to connect to the server: x509: certificate signed by unknown authority
```

#### Recourse

Check if you're on any sort of VPN or other egress proxy that would break SSL.  Also, there is a possibility that your credentials have expired. In the case where you're using GKE and gcloud, renew your credentials via:

```
kubectl get all
Unable to connect to the server: x509: certificate signed by unknown authority
```

```
gcloud container clusters get-credentials my-cluster-name-dev
Fetching cluster endpoint and auth data.
kubeconfig entry generated for my-cluster-name-dev.
```

```
kubectl config current-context
gke_my-org_us-east1-b_my-cluster-name-dev
```

### Uploads and Downloads are Slow

#### Symptom

Any `pachctl put file` or `pachctl get file` commands are slow.

#### Recourse

If you do not explicitly set the pachd address config value, `pachctl` will default to using port forwarding, which throttles traffic to ~1MB/s. If you need to do large downloads/uploads you should consider using pachd address config value. You'll also want to make sure you've allowed ingress access through any firewalls to your k8s cluster.

### Naming a Repo with an Unsupported Symbol

#### Symptom

A Pachyderm repo was accidentally named starting with a dash (`-`) and the repository
is treated as a command flag instead of a repository.

#### Recourse

Pachyderm supports standard `bash` utilities that you can
use to resolve this and similar problems. For example, in this case,
you can specify double dashes (`--`) to delete the repository. Double dashes
signify the end of options and tell the shell to process the
rest arguments as filenames and objects.

For more information, see `man bash`.

### Failed Uploads

#### Symptom

A file upload, particularly a recursive one of many files, fails. You may see log messages containing the following in either pipeline logs, pachd logs, or from the pachctl command locally:
- pachctl errror: ``an error occurred forwarding XXXXX -> 650: error forwarding port 650``
- pachctl error: ``EOF``
- pachd or worker: ``all SubConns are in TransientFailure, latest connection error: connection error: desc = \"transport: Error while dialing dial tcp  127.0.0.1:653: connect: connection refused\"; retrying in XXXX.XXXXXs"}``

#### Recourse

Either ``pachd`` or your pipeline's worker sidecar may be getting OOM killed as it grows while getting data from object storage. 

You can give the storage container more resources by increasing the ``cache_size`` parameter in your pipeline spec. Increase it to what you can afford; its default is 64M.(If you’re using a release prior to 1.10.0 and you have cluster-wide or namepace policies on resource limits, you may need to manually edit the pipeline RC.)

If it still gets OOM killed by k8s, there are a couple of environment variables you can set in your pachd deployment to limit the amount of memory the sidecar and pachd use.

- ``STORAGE_UPLOAD_CONCURRENCY_LIMIT`` limits the parallelism to put files into the storage backend. Default is 100.
- ``STORAGE_PUT_FILE_CONCURRENCY_LIMIT`` limits the number of parallel downloads pachd will initiate. Default is also 100.

You may use a binary search technique to hone in on a value appropriate for a production pipeline:

for ``cache_size``, max it out. If it works, halve it. If its OOM killed, increase the value by 50%. and so on
for the ``CONCURRENCY_LIMITS``, halve and increase by 50% until you get a value that works.
