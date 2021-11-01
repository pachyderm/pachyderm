# Troubleshooting Pipelines

## Introduction

Job failures can occur for a variety of reasons, but they generally categorize into 3 failure types: 

1. [User-code-related](#user-code-failures): An error in the user code running inside the container or the json pipeline config.
1. [Data-related](#data-failures): A problem with the input data such as incorrect file type or file name.
1. [System- or infrastructure-related](#system-level-failures): An error in Pachyderm or Kubernetes such as missing credentials, transient network errors, or resource constraints (for example, out-of-memory--OOM--killed).

In this document, we'll show you the tools for determining what kind of failure it is.  For each of the failure modes, we’ll describe Pachyderm’s and Kubernetes’s specific retry and error-reporting behaviors as well as typical user triaging methodologies. 

Failed jobs in a pipeline will propagate information to downstream pipelines with empty commits to preserve provenance and make tracing the failed job easier.  A failed job is no longer running.

In this document, we'll describe what you'll see, how Pachyderm will respond, and techniques for triaging each of those three categories of failure. 

At the bottom of the document, we'll provide specific troubleshooting steps for [specific scenarios](#specific-scenarios).

- [Pipeline exists but never runs](#pipeline-exists-but-never-runs) 
- [All your pods or jobs get evicted](#all-your-pods-or-jobs-get-evicted)

### Determining the kind of failure

First off, you can see the status of Pachyderm's jobs with `pachctl list job`, which will show you the status of all jobs.  For a failed job, use `pachctl inspect job <job-id>` to find out more about the failure.  The different categories of failures are addressed below.

### User Code Failures

When there’s an error in user code, the typical error message you’ll see is 

```
failed to process datum <UUID> with error: <user code error>
```

This means pachyderm successfully got to the point where it was running user code, but that code exited with a non-zero error code. If any datum in a pipeline fails, the entire job will be marked as failed, but datums that did not fail will not need to be reprocessed on future jobs.   You can use `pachctl inspect datum <job-id> <datum-id>` or `pachctl logs` with the `--pipeline`, `--job` or `--datum` flags to get more details.

There are some cases where users may want mark a datum as successful even for a non-zero error code by setting the `transform.accept_return_code` field in the pipeline config .

#### Retries
Pachyderm will automatically retry user code three (3) times before marking the datum as failed. This mitigates datums failing for transient connection reasons.

#### Triage
`pachctl logs --job=<job_ID>` or `pachctl logs --pipeline=<pipeline_name>` will print out any logs from your user code to help you triage the issue. Kubernetes will rotate logs occasionally so if nothing is being returned, you’ll need to make sure that you have a persistent log collection tool running in your cluster. If you set `enable_stats:true` in your pachyderm pipeline, pachyderm will persist the user logs for you. 

In cases where user code is failing, changes first need to be made to the code and followed by updating the pachyderm pipeline. This involves building a new docker container with the corrected code, modifying the pachyderm pipeline config to use the new image, and then calling `pachctl update pipeline -f updated_pipeline_config.json`. Depending on the issue/error, user may or may not want to also include the `--reprocess` flag with `update pipeline`. 

### Data Failures

When there’s an error in the data, this will typically manifest in a user code error such as 

```
failed to process datum <UUID> with error: <user code error>
```

This means pachyderm successfully got to the point where it was running user code, but that code exited with a non-zero error code, usually due to being unable to find a file or a path, a misformatted file, or incorrect fields/data within a file. If any datum in a pipeline fails, the entire job will be marked as failed. Datums that did not fail will not need to be reprocessed on future jobs.

#### Retries
Just like with user code failures, Pachyderm will automatically retry running a datum 3 times before marking the datum as failed. This mitigates datums failing for transient connection reasons.

#### Triage
Data failures can be triaged in a few different way depending on the nature of the failure and design of the pipeline. 

In some cases, where malformed datums are expected to happen occasionally, they can be “swallowed” (e.g. marked as successful using `transform.accept_return_codes` or written out to a “failed_datums” directory and handled within user code). This would simply require the necessary updates to the user code and pipeline config as described above.  For cases where your code detects bad input data, a "dead letter queue" design pattern may be needed.  Many pachyderm developers use a special directory in each output repo for "bad data" and pipelines with globs for detecting bad data direct that data for automated and manual intervention.

Pachyderm's engineering team is working on changes to the Pachyderm Pipeline System in a future release that may make implementation of design patterns like this easier.   [Take a look at the pipeline design changes for pachyderm 1.9](https://github.com/pachyderm/pachyderm/issues/3345)

If a few files as part of the input commit are causing the failure, they can simply be removed from the HEAD commit with `start commit`, `delete file`, `finish commit`. The files can also be corrected in this manner as well. This method is similar to a revert in Git -- the “bad” data will still live in the older commits in Pachyderm, but will not be part of the HEAD commit and therefore not processed by the pipeline.

If the entire commit is bad and you just want to remove it forever as if it never happened, `delete commit` will both remove that commit and all downstream commits and jobs that were created as downstream effects of that input data. 

### System-level Failures

System-level failures are the most varied and often hardest to debug. We’ll outline a few common patterns and triage steps. Generally, you’ll need to look at deeper logs to find these errors using `pachctl logs --pipeline=<pipeline_name> --raw` and/or `--master` and `kubectl logs pod <pod_name>`.

Here are some of the most common system-level failures:

- Malformed or missing credentials such that a pipeline cannot connect to object storage, registry, or other external service. In the best case, you’ll see `permission denied` errors, but in some cases you’ll only see “does not exist” errors (this is common reading from object stores)
- Out-of-memory (OOM) killed or other resource constraint issues such as not being able to schedule pods on available cluster resources. 
- Network issues trying to connect Pachd, etcd, or other internal or external resources
- Failure to find or pull a docker image from the registry

#### Retries
For system-level failures, Pachyderm or Kubernetes will generally continually retry the operation with exponential backoff. If a job is stuck in a given state (e.g. starting, merging) or a pod is in `CrashLoopBackoff`, those are common signs of a system-level failure mode.


#### Triage
Triaging system failures varies as widely as the issues do themselves. Here are options for the common issues mentioned previously.
- Credentials: check your secrets in k8s, make sure they’re added correctly to the pipeline config, and double check your roles/perms within the cluster
- OOM: Increase the memory limit/request or node size for your pipeline. If you are very resource constrained, making your datums smaller to require less resources may be necessary. 
- Network: Check to make sure etcd and pachd are up and running, that k8s DNS is correctly configured for pods to resolve each other and outside resources, firewalls and other networking configurations allow k8s components to reach each other, and ingress controllers are configured correctly
- Check your container image name in the pipeline config and image_pull_secret.

## Specific scenarios

### All pods or jobs get evicted

#### Symptom

After creating a pipeline, a job starts but never progresses through
any datums.

#### Recourse

Run `kubectl get pods` and see if the command returns pods that
are marked `Evicted`. If you run `kubectl describe <pod-name>` with
one of those evicted pods, you might get an error saying that it was
evicted due to disk pressure. This means that your nodes are not
configured with a big enough root volume size.
You need to make sure that each node's root volume is big enough to
store the biggest datum you expect to process anywhere on your DAG plus
the size of the output files that will be written for that datum.

Let's say you have a repo with 100 folders. You have a single pipeline
with this repo as an input, and the glob pattern is `/*`. That means
each folder will be processed as a single datum. If the biggest folder
is 50GB and your pipeline's output is about three times as big, then your
root volume size needs to be bigger than:

```
50 GB (to accommodate the input) + 50 GB x 3 (to accommodate the output) = 200GB
```

In this case we would recommend 250GB to be safe. If your root
volume size is less than 50GB (many defaults are 20GB), this pipeline
will fail when downloading the input. The pod may get evicted and
rescheduled to a different node, where the same thing will happen.

### Pipeline exists but never runs

#### Symptom

You can see the pipeline via:

```
pachctl list pipeline
```

But if you look at the job via:

```
pachctl list job
```

It's marked as running with `0/0` datums having been processed.  If you inspect the job via:

```
pachctl inspect job
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

First make sure that there is no parent job still running. Do `pachctl list job | grep yourPipelineName` to see if there are pending jobs on this pipeline that were kicked off prior to your job. A parent job is the job that corresponds to the parent output commit of this pipeline. A job will block until all parent jobs complete.

If there are no parent jobs that are still running, then continue debugging:

Describe the pod via:

```
$kubectl describe po/pipeline-foo-5-v1-273zc
```

If the state is `CrashLoopBackoff`, you're looking for a descriptive error message. One such cause for this behavior might be if you specified an image for your pipeline that does not exist.

If the state is `Pending` it's likely the cluster doesn't have enough resources. In this case, you'll see a `could not schedule` type of error message which should describe which resource you're low on. This is more likely to happen if you've set resource requests (cpu/mem/gpu) for your pipelines.  In this case, you'll just need to scale up your resources. If you deployed using `kops`, you'll want to do edit the instance group, e.g. `kops edit ig nodes ...` and up the number of nodes. If you didn't use `kops` to deploy, you can use your cloud provider's auto scaling groups to increase the size of your instance group. Either way, it can take up to 10 minutes for the changes to go into effect. 

For more information, see [Autoscale Your Cluster](../deploy-manage/manage/autoscaling.md).

### Cannot Delete Pipelines with an etcd Error

Failed to delete a pipeline with an `etcdserver` error.

#### Symptom

Deleting pipelines fails with the following error:

```shell
$ pachctl delete pipeline pipeline-name
etcdserver: too many operations in txn request (XXXXXX comparisons, YYYYYYY writes: hint: set --max-txn-ops on the ETCD cluster to at least the largest of those values)
```

#### Recourse

When a Pachyderm cluster reaches a certain scale, you need to adjust
the default parameters provided for certain `etcd` flags.
Depending on how you deployed Pachyderm,
you need to either edit the `etcd` `Deployment` or `StatefulSet`.

```shell
$ kubectl edit deploy etcd
```

or

```shell
$ kubectl edit statefulset etcd
```

In the `spec/template/containers/command` path, set the value for
`max-txn-ops` to a value appropriate for your cluster, in line
with the advice in the error above: *larger than the greater of XXXXXX or YYYYYYY*.

### Pipeline is stuck in `starting`

#### Symptom

After starting a pipeline, running the `pachctl list pipeline` command returns
the `starting` status for a very long time. The `kubectl get pods` command
returns the pipeline pods in a pending state indefinitely.

#### Recourse

Run the `kubectl describe pod <pipeline-pod>` and analyze the
information in the output of that command. Often, this type of error
is associated with insufficient amount of CPU, memory, and GPU resources
in your cluster.
