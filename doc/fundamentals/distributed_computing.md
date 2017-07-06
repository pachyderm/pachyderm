# Distributed Computing

Distributing computation across multiple workers is a fundamental part of processing any big data or computationally intensive workload. There are two main questions to think about when trying to distribute computation:

1. [How many workers to spread computation across?](#controlling-the-number-of-workers-parallelism)

2. [How to define which workers are responsible for which data?](#spreading-data-across-workers-glob-patterns)

## Pachyderm Workers

Before we dive into the above questions, there are a few details you should understand about Pachyderm workers.

Every worker for a given pipeline is an identical pod running the Docker image you specified in the [pipeline spec](../reference/pipeline_spec.html). Your analysis code does not need do anything special to run in a distributed fashion. Instead, Pachyderm will spread out the data that needs to be processed across the various workers and make that data available for your code.

Pachyderm workers are spun up when you create the pipeline and are left running in the cluster waiting for new jobs (data) to be available for processing (committed). This saves having to recreate and schedule the worker for every new job.

## Controlling the Number of Workers (Parallelism)

The number of workers that are used for a given pipeline is controlled by the `parallelism_spec` defined in the [pipeline specification](../reference/pipeline_spec.html).

```
  "parallelism_spec": {
    // Exactly one of these two fields should be set
    "constant": int
    "coefficient": double
```
Pachyderm has two parallelism strategies: `constant` and `coefficient`. You should set one of the two corresponding fields in the `parallelism_spec`, and pachyderm chooses a parallelism strategy based on which field is set.

If you set the `constant` field, Pachyderm will start the number of workers that you specify. For example, set `"constant":10` to use 10 workers.

If you set the `coefficient` field, Pachyderm will start a number of workers that is a multiple of your Kubernetes cluster’s size. For example, if your Kubernetes cluster has 10 nodes, and you set `"coefficient": 0.5`, Pachyderm will start five workers. If you set it to 2.0, Pachyderm will start 20 workers (two per Kubernetes node).

**NOTE:** The parallelism_spec is optional and will default to `“coefficient": 1`, which means that it'll spawn one worker per Kubernetes node for this pipeline if left unset.

## Spreading Data Across Workers (Glob Patterns)

Defining how your data is spread out among workers is arguably the most important aspect of distributed computation and is the fundamental idea around concepts like Map/Reduce.

Instead of confining users to just data-distribution patterns such as Map (split everything as much as possible) and Reduce (_all_ the data must be grouped together), Pachyderm uses [Glob Patterns](https://en.wikipedia.org/wiki/Glob_(programming)) to offer incredible flexibility in defining your data distribution.

 Glob patterns are defined by the user for each `atom` within the `input` of a pipeline, and they tell Pachyderm how to divide the input data into individual "datums" that can be processed independently.

```
"input": {
    "atom": {
        "repo": "string",
        "glob": "string",
    }
}
```

That means you could easily define multiple "atoms", one with the data highly distributed and another where it's grouped together.  You can then join the datums in these atoms via a cross product or union (as shown above) for combined, distributed processing.

```
"input": {
    "cross" or "union": [
        {
            "atom": {
                "repo": "string",
                "glob": "string",
            }
        },
        {
            "atom": {
                "repo": "string",
                "glob": "string",
            }
        },
        etc...
    ]
}
```

More information about "atoms," unions, and crosses can be found in our [Pipeline Specification](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html).

### Datums

Pachyderm uses the glob pattern to determine how many “datums” an input atom consists of. Datums are the unit of parallelism in Pachyderm. That is, Pachyderm attempts to process datums in parallel whenever possible.

If you have two workers and define 2 datums, Pachyderm will send one datum to each worker. In a scenario where there are more datums than workers, Pachyderm will queue up extra datums and send them to workers as they finish processing previous datums.

### Defining Datums via Glob Patterns

Intuitively, you should think of the input atom repo as a file system where the glob pattern is being applied to the root of the file system. The files and directories that match the glob pattern are considered datums.

For example, a glob pattern of just `/` would denote the entire input repo as a single datum. All of the input data would be given to a single worker similar to a typical reduce-style operation.

Another commonly used glob pattern is `/*`. `/*` would define each top level object (file or directory) in the input atom repo as its own datum. If you have a repo with just 10 files in it and no directory structure, every file would be a datum and could be processed independently. This is similar to a  typical map-style operation.

But Pachyderm can do anything in between too. If you have a directory structure with each state as a directory and a file for each city such as:

```
/California
   /San-Francisco.json
   /Los-Angeles.json
   ...
/Colorado
   /Denver.json
   /Boulder.json
   ...
...
```

and you need to process all the data for a given state together, `/*` would also be the desired glob pattern. You'd have one datum per state, meaning all the cities for a given state would be processed together by a single worker, but each state can be processed independently.

If we instead used the glob pattern `/*/*` for the states example above, each `<city>.json` would be it's own datum.

Glob patterns also let you take only a particular directory (or subset of directories) as an input atom instead of the whole input repo. If we create a pipeline that is specifically only for California, we can use a glob pattern of `/California/*` to only use the data in that directory as input to our pipeline.

### Only Processing New Data

A datum defines the granularity at which Pachyderm decides what data is new and what data has already been processed. Pachyderm will never reprocess datums it's already seen with the same analysis code. But if any part of a datum changes, the entire datum will be reprocessed.

**Note:** If you change your code (or pipeline spec), Pachyderm will of course allow you to process all of the past data through the new analysis code.

Let's look at our states example with a few different glob patterns to demonstrate what gets processed and what doesn't. Suppose we have an input data layout such as:

```
/California
   /San-Francisco.json
   /Los-Angeles.json
   ...
/Colorado
   /Denver.json
   /Boulder.json
   ...
...
```

If our glob pattern is `/`, then the entire input atom is a single datum, which means anytime any file or directory is changed in our input, all the the data will be processed from scratch. There are plenty of usecases where this is exactly what we need (e.g. some machine learning training algorithms).

If our glob pattern is `/*`, then each state directory is it's own datum and we'll only process the ones that have changed. So if we add a  new city file, `Sacramento.json` to the `/California` directory, _only_ the California datum, will be reprocessed.

If our glob pattern was `/*/*` then each `<city>.json` file would be it's own datum. That means if we added a `Sacramento.json` file, only that specific file would be processed by Pachyderm.

