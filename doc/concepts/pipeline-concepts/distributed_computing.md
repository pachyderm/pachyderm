# Distributed Computing

Distributing computation workload across multiple workers
is a fundamental part of any big data processing.
When you design your
Pachyderm pipeline system for production, you need to
define the number of Pachyderm workers across which you want
to spread your computations and which workers are responsible
for which data.

## Pachyderm Workers

A Pachyderm worker is an identical Kubernetes pod that runs
the Docker image that you specified in the
[pipeline spec](../reference/pipeline_spec.html). Your analysis code
does not affect how Pachyderm distributes the workload among workers.
Instead, Pachyderm spreads out the data that needs to be processed
across the various workers and makes that data available for your code.

When you create a pipeline, Pachyderm spins up worker pods that
continuously run in the cluster waiting for new data to be available
for processing. Therefore, you do not need to recreate and
schedule workers for every new job.

<!-- The following diagram shows how distributed computing works in
Pachyderm - TBA -->

## Controlling the Number of Workers

You can control the number of worker pods that Pachyderm runs in a
pipeline by defining the `parallelism` parameter in the
[pipeline specification](../reference/pipeline_spec.html).

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
* [Pipeline Specification](../../reference/pipeline_spec)
