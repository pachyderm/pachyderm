# Distributed Computing

Distributing your computations across multiple workers
is a fundamental part of any big data processing.
When you build production-scale pipelines, you need
to adjust the number of workers and resources that are
allocated to each job to optimize throughput.

## Pachyderm Workers

A Pachyderm worker is an identical Kubernetes pod that runs
the Docker image that you specified in the
[pipeline spec](../../reference/pipeline_spec.md). Your analysis code
does not affect how Pachyderm distributes the workload among workers.
Instead, Pachyderm spreads out the data that needs to be processed
across the various workers and makes that data available for your code.

When you create a pipeline, Pachyderm spins up worker pods that
continuously run in the cluster waiting for new data to be available
for processing. You can change this behavior by setting `"standby" :true`.
Therefore, you do not need to recreate and
schedule workers for every new job.

For each job, all the datums are queued up and then distributed
across the available workers. When a worker finishes processing
its datum, it grabs a new datum from the queue until all datums
complete processing. If a worker pod crashes, its datums are
redistributed to other workers for maximum fault tolerance.

<!-- The following diagram shows how distributed computing works in
Pachyderm - TBA Possibly could be a gif. :) Show queue of
datums and 3 workers running things in parallel. Technically,
each worker can download a datum, process a datum, and upload a
completed datum all in parallel. May or may not want to show
this, but wouldn't be too hard. We can draw this out in TOH.-->

## Controlling the Number of Workers

You can control the number of worker pods that Pachyderm runs in a
pipeline by defining the `parallelism` parameter in the
[pipeline specification](../../reference/pipeline_spec.md).

```
  "parallelism_spec": {
    // Exactly one of these two fields should be set
    "constant": int
    "coefficient": double
```

Pachyderm has the following parallelism strategies that you
can set in the pipeline spec:

| Strategy       | Description        |
| -------------- | ------------------ |
| `constant`     | Pachyderm starts the specified number of workers. For example, <br> if you set `"constant":10`, Pachyderm spreads the computation workload among ten workers. |
| `coefficient`  | Pachyderm starts a number of workers that is a multiple of <br> your Kubernetes cluster size. For example, if your Kubernetes cluster has ten nodes, <br> and you set `"coefficient": 0.5`, Pachyderm starts five workers. If you set parallelism to `"coefficient": 2.0`, Pachyderm starts twenty workers. |

By default, Pachyderm sets `parallelism` to `“coefficient": 1`, which means
that it spawns one worker per Kubernetes node for this pipeline.

**See also:**

* [Glob Pattern](../datum/glob-pattern)
* [Pipeline Specification](../../reference/pipeline_spec.md)
