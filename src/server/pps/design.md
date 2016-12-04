# PPS Design Doc



Drawing inspirations from Kubernetes, PPS stores the **desired state** in a persistent database (RethinkDB), and use a set of **controllers** to make the actual state match the desired state.

To explain how PPS works, we are going to go through the entire lifecycle of a pipeline:

1. Create a pipeline
2. Run the pipeline
3. Receive input commits from PFS
4. Create chunks and spawn jobs
5. Monitor the liveness of pods
6. Finish jobs

## Create a pipeline

When a user creates a pipeline (e.g. via `pachctl create-pipeline`), the pipeline spec is written to the database.  That's it.

## Run the pipeline

We use etcd to shard pipelines among multiple PPS nodes.  When a PPS node starts up, it's assigned a set of shards.  Each pipeline also has a shard number, which is basically determined from the pipeline name.  Thus, each PPS node knows which subset of pipelines it's supposed to run. 

For each shard assigned, a PPS node launches a goroutine that subscribes to [a RethinkDB changefeed](https://www.rethinkdb.com/docs/changefeeds/ruby/), which delivers the pipelines for that shard.  Thus, when a new pipeline spec is inserted into the database, one (and only one) PPS node receives the pipeline.

When a PPS node receives a pipeline, it spawns a `pipelineManager` goroutine that subscribes to new input commits and launch jobs accordingly, as described in the next section.

## Receive input commits from PFS

A `pipelineManager` receives input commits from PFS via `ListCommit`.  Specifically, it uses the `block` flag with `ListCommit` such that it blocks until new commits are added.  

If the pipeline has only one input repo, then each commit in the input repo triggers a job.  It's worth explaining what happens if the pipeline has more than one input repos.  Consider this example:

```
A1   B2
A3   B4
A5
```

Here the pipeline has two input repos A and B.  The order in which the commits are made is specified by the number: A1 -> B2 -> A3 -> B4 -> A5.  Now let's examine what happens as each commit is made.

1. When A1 is made, the pipeline doesn't spawn a job, because there's no commit in B yet.
2. When B2 is made, the pipeline spawns a job with the input commits {A1, B2}.
3. When A3 is made, the pipeline spawns a job with the input commits {A3, B2}.
4. When B4 is made, the pipeline spawns a job with the input commits {A1+A3, B4}.
5. When A5 is made, the pipeline spawns a job with the input commits {A5, B2+B4}.

Basically, the goal is to ensure that every pair of input commits is processed.  Therefore when a new commit is made, a job is spawn with the new commit as input, along with **all** commits in the other input repos.

## Create chunks and spawn jobs

To enable parallelism, we divide a job's input into `chunks`.  A chunk is a portion of a set of input commits.

Each job in PPS corresponds to a [Kubernetes job](http://kubernetes.io/docs/user-guide/jobs/).  A Kubernetes job consists of a number of `pods`, each of which runs a copy of the user's code.

When a pod starts up, it sends a request to PPS to ask for a chunk to process.  Currently, a pod exits once it has processed a chunk.  This works because currently the number of pods is always equal to the number of chunks.  In the future however, we should decouple pods and chunks to enable more advanced features such as straggler mitigation.  For instance, if a chunk is taking too long to process, we can simply split the chunk into smaller chunks so the other pods can share the work.  See [#442](https://github.com/pachyderm/pachyderm/issues/442).

## Monitor the liveness of pods

Normally, if a pod crashes, Kubernetes will instantiate a new pod, due to the fact that a Kubernetes job always ensures that its pods successfully terminate.  However, the new pod will have a completely new identity, and thus when it comes to PPS to ask for a chunk, PPS needs to somehow know to hand it the chunk that was originally being processed by the crashed pod.

Furthermore, it's possible for a pod to get into a state where it's still "alive" from the perspective of Kubernetes (and thus a new pod is not created), but is not making any progress from the perspective of the pipeline.  For instance, there could be a network partition between a pod and PPS/PFS, but no partition between the pod and the Kubernetes master.

To handle these challenges, PPS uses a lease system to manage chunks.  When a pod requests for a chunk, PPS "leases" the chunk to the pod, and starts an internal timer that revokes the lease after a configurable period.  To keep a lease, a pod needs to continuously send heartbeats to PPS to report that it's alive.  Everytime such a heartbeat is received, PPS renews the lease and restarts the timer.

Now let's consider what happens when a pod crashes.  Because the pod crashed, it stops sending heartbeats to PPS, so the lease on the chunk is automatically revoked after a while.  As soon as the lease expires, PPS can give the chunk to the new pod who has been requesting a chunk.  Thus we've completed the handoff between the crashed pod and the new pod.

In the case of network partitions between PPS and a pod, PPS will again stop receiving heartbeats and thus revoke the lease.  Pods are also programmed to crash if they can't reach PPS, causing new pods to be spawned in their place.  Thus, we again have a handoff of chunks from old, bad pods to new, functioning pods.
