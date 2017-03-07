## ./pachctl create-pipeline

Create a new pipeline.

### Synopsis


Create a new pipeline from a spec

# Pipeline Specification

This document discusses each of the fields present in a pipeline specification.
To see how to use a pipeline spec, refer to the [pachctl
create-pipeline](../pachctl/pachctl_create-pipeline.html) doc.

## JSON Manifest Format

```
{
  "pipeline": {
    "name": string
  },
  "transform": {
    "image": string,
    "cmd": [ string ],
    "stdin": [ string ]
    "env": {
        "foo": "bar"
    },
    "secrets": [ {
        "name": "secret_name",
        "mountPath": "/path/in/container"
    } ],
    "imagePullSecrets": [ "my_secret" ],
    "overwrite": bool
  },
  "parallelism_spec": {
    "strategy": "CONSTANT"|"COEFFICIENT"
    "constant": int        // if strategy == CONSTANT
    "coefficient": double  // if strategy == COEFFICIENT
  },
  "inputs": [
    {
      "repo": {
        "name": string
      },
      "runEmpty": false,
      "lazy": false,
      "method": "map"/"reduce"/"global"
      // alternatively, method can be specified as an object.
      // this is only for advanced use cases; most of the time, one of the three
      // strategies above should suffice.
      "method": {
        "partition": "BLOCK"/"FILE"/"REPO",
        "incremental": "NONE"/"DIFF"/"FILE",
      }
    }
  ],
  "output": {
    "URL": "s3://bucket/dir"
  },
  "gc_policy": {
    "success": string,
    "failure": string
  }
}
```

### Name (required)

`pipeline.name` is the name of the pipeline that you are creating.  Each
pipeline needs to have a unique name.

### Transform (required)

`transform.image` is the name of the Docker image that your jobs run in.

`transform.cmd` is the command passed to the Docker run invocation.  Note that
as with Docker, cmd is not run inside a shell which means that things like
wildcard globbing (`*`), pipes (`|`) and file redirects (`>` and `>>`) will not
work.  To get that behavior, you can set `cmd` to be a shell of your choice
(e.g. `sh`) and pass a shell script to stdin.

`transform.stdin` is an array of lines that are sent to your command on stdin.
Lines need not end in newline characters.

`transform.env` is a map from key to value of environment variables that will be
injected into the container

`transform.secrets` is an array of secrets, secrets reference Kubernetes
secrets by name and specify a path that the secrets should be mounted to.
Secrets are useful for embedding sensitive data such as credentials. Read more
about secrets in Kubernetes
[here](http://kubernetes.io/docs/user-guide/secrets/).

`transform.imagePullSecrets` is an array of image pull secrets, image pull
secrets are similar to secrets except that they're mounted before the
containers are created so they can be used to provide credentials for image
pulling. Read more about image pull secrets
[here](http://kubernetes.io/docs/user-guide/images/#specifying-imagepullsecrets-on-a-pod).

`transform.overwrite` is a boolean flag that controls whether the output of
this pipeline overwrites the previous output, as opposed to appending to it
(the default).  For instance, if `overwrite` is set to true and the pipeline
outputs the same file in two subsequent runs, the second write to that file
will overwrite the first one.  This flag defaults to false.

### Parallelism Spec (optional)

`parallelism_spec` describes how Pachyderm should parallelize your pipeline.
Currently, Pachyderm has two parallelism strategies: `CONSTANT` and
`COEFFICIENT`.

If you use the `CONSTANT` strategy, Pachyderm will start a number of workers
that you give it. To use this strategy, set the field `strategy` to `CONSTANT`,
and set the field `constant` to an integer value (e.g. `10` to start 10 workers).

If you use the `COEFFICIENT` strategy, Pachyderm will start a number of workers
that is a multiple of your Kubernetes cluster's size. To use this strategy, set
the field `coefficient` to a double. For example, if your Kubernetes cluster
has 10 nodes, and you set `coefficient` to 0.5, Pachyderm will start five
workers. If you set it to 2.0, Pachyderm will start 20 workers (two per
Kubernetes node).

By default, we use the parallelism spec "coefficient=1".

__Note:__ Pachyderm treats this config as an upper bound. Pachyderm may choose
to start fewer workers than specified if the pipeline's input data set is small
or otherwise doesn't parallelize well (for example, if you use an input method
of file and the input repo only has one file in it).

### GC Policy (optional)

`gc_policy` specifies when completed jobs can be garbage collected.  Completed
jobs are typically kept around for a while because the logs can be of interest
to developers, especially if a job has failed.

`gc_policy` has two string fields: `success` specifies how long successful jobs
are kept around, and `failure` specifies how long failed jobs are kept around.  
The string needs to be sequence of decimal numbers, each with optional fraction
and a unit suffix, such as "300ms", "1.5h" or "2h45m". Valid time units are "s", 
"m", "h".

By default, successful jobs are kept for 24 hours (a day), and failed jobs are
kept for 168 hours (a week).

### Inputs (optional)

`inputs` specifies a set of Repos that will be visible to the jobs during
runtime. Commits to these repos will automatically trigger the pipeline to
create new jobs to process them.

`inputs.runEmpty` specifies what happens when an empty commit (i.e. no data)
comes into the input repo of this pipeline (for example, if an input pipeline
produced no data). If this flag is set to true, Pachyderm will still
run your pipeline even if it has no new input data to process.  Specifically:
if this flag is set to false (the default), then an empty commit won't trigger
a job; if set to true, an empty commit will trigger a job.

`inputs.lazy` controls how the data is exposed to jobs. The default is `false`
which means the job will eagerly download the data it needs to process and it
will be exposed as normal files on disk. If `true` data will be exposed as
named pipes instead and no data will be downloaded until the job opens the pipe
and reads it, if the pipe is never opened then no will be downloaded. Some
applications won't work with pipes, for example if they make syscalls such as
`Seek` which pipes don't support. Applications that can work with pipes should
use them since they're more performant, the difference will be especially
notable if the job only reads a subset of the files that are available to it.

`inputs.method` specifies two different properties:

- Partition unit: How input data  will be partitioned across parallel
    containers.
- Incrementality: Whether the entire all of the data or just the new data
    (diff) is processed.

The next section explains input methods in detail.

### Output (optional)
`output` allows you to push the results of a Pipeline to an external data store
such as s3, Google Cloud Storage or Azure Storage. Data will be pushed after
the user code has finished running but before the job is marked as successful.
Pipelines and jobs with `output`s will still output to a Pachyderm repo.

## Pipeline Input Methods

For each pipeline input, you may specify a "method".  A method dictates exactly
what happens in the pipeline when a commit comes into the input repo.

A method consists of two properties: partition unit and incrementality.

### Partition Unit
Partition unit ("BLOCK", "FILE", or "REPO") specifies the granularity at which
input data is parallelized across containers.  It can be of three values:

1.  `BLOCK`: different blocks of the same file may be parelleized across
    containers.
2. `FILE`: the files and/or directories residing under the root directory (/)
   must be grouped together.  For instance, if you have four files in a
   directory structure like:

```
/foo
/bar
/buzz
   /a
   /b
```
then there are only three top-level objects, `/foo`, `/bar`, and `/buzz`, each
of which will remain grouped in the same container.
3. `REPO`: the entire repo.  In this case, the input won't be partitioned at
   all.

### Incrementality

Incrementality ("NONE", "DIFF" or "FILE") describes what data needs to be
available when a new commit is made on an input repo. Namely, do you want to
process _only the new data_ in that commmit (the "DIFF"), only files with any
new data ("FILE"), or does all of the data need to be reprocessed ("NONE")?

For instance, if you have a repo with the file `/foo` in commit 1 and file
`/bar` in commit 2, then:

* If the input incrementality is "DIFF", the first job sees file `/foo` and the
second job sees file `/bar`.

* If the input is non-incremental("NONE"), every job sees all the data. The
first job sees file `/foo` and the second job sees file `/foo` and file `/bar`.

* "FILE" (Top-level objects) means that if any part in a file (or alternatively
any file within a directory) changes, then show all the data in that file
(directory). For example, you may have vendor data files in separate
directories by state -- the California directory contains a file for each
california vendor, etc.  `Incremental: "FILE"` would mean that your job will
see the entire directory if at least one file in that directory has changed. If
only one vendor file in the whole repo was was changed and it was in the
Colorado directory, all Colorado vendor files would be present, but that's it.

### Combining Partition unit and Incrementality

For convenience, we have defined aliases for the three most commonly used (and
most familiar) input methods: "map", "reduce", and "global".

* A `map` (BLOCK + DIFF), for example, can partition files at the block level
and jobs only need to see the new data.

* `reduce` (FILE + NONE) as it's typically seen in Hadoop, requires all parts
of a file to be seen by the same container ("FILE") and your job needs to
reprocess _all_ the data in the whole repo ("NONE").

* `global` (REPO + NONE), means that the entire repo needs to be seen by
_every_ container. This is commonly used if you had a repo with just
parameters, and every container needed to see all the parameter data and pull
out the ones that are relevant to it.

They are defined below:

```
                             +-----------------------------------------+
                             |             Partition Unit              |
  +--------------------------+---------+----------------------+--------+
  |     Incrementality       | "Block" | "FILE" (Top-lvl Obj) | "REPO" |
  +==========================+=========+======================+========+
  | "NONE" (non-incremental) |         |       "reduce"       |"global"|
  +--------------------------+---------+----------------------+--------+
  | "DIFF" (incremental)     |  "map"  |                      |        |
  +--------------------------+---------+----------------------+--------+
  | "FILE" (top-lvl object)  |         |                      |        |
  +--------------------------+---------+----------------------+--------+
```

### Defaults
If no method is specified, the `map` method (BLOCK + DIFF) is used by default.

## Multiple Inputs

A pipeline is allowed to have multiple inputs.  The important thing to
understand is what happens when a new commit comes into one of the input repos.
In short, a pipeline processes the **cross product** of its inputs.  We will
use an example to illustrate.

Consider a pipeline that has two input repos: `foo` and `bar`.  `foo` uses the
`file/incremental` method and `bar` uses the `reduce` method.  Now let's say
that the following events occur:

```
1. PUT /file-1 in commit1 in foo -- no jobs triggered
2. PUT /file-a in commit1 in bar -- triggers job1
3. PUT /file-2 in commit2 in foo -- triggers job2
4. PUT /file-b in commit2 in bar -- triggers job3
```

The first time the pipeline is triggered will be when the second event
completes.  This is because we need data in both repos before we can run the
pipeline.

Here is a breakdown of the files that each job sees:

```
job1:
    /pfs/foo/file-1
    /pfs/bar/file-a

job2:
    /pfs/foo/file-2
    /pfs/bar/file-a

job3:
    /pfs/foo/file-1
    /pfs/foo/file-2
    /pfs/bar/file-a
    /pfs/bar/file-b
```

`job1` sees `/pfs/foo/file-1` and `/pfs/bar/file-a` because those are the only
files available.

`job2` sees `/pfs/foo/file-2` and `/pfs/bar/file-a` because it's triggered by
commit2 in `foo`, and `foo` uses an incremental input method
(`file/incremental`).

`job3` sees all the files because it's triggered by commit2 in `bar`, and `bar`
uses a non-incremental input method (`reduce`).

## Examples

```json
{
  "pipeline": {
    "name": "my-pipeline"
  },
  "transform": {
    "image": "my-image",
    "cmd": [ "my-binary", "arg1", "arg2"],
    "stdin": [
        "my-std-input"
    ]
  },
  "parallelism": "4",
  "inputs": [
    {
      "repo": {
        "name": "my-input"
      },
      "method": "map"
    }
  ]
}
```

This pipeline runs when the repo `my-input` gets a new commit.  The pipeline
will spawn 4 parallel jobs, each of which runs the command `my-binary` in the
Docker image `my-image`, with `arg1` and `arg2` as arguments to the command and
`my-std-input` as the standard input.  Each job will get a set of blocks from
the new commit as its input because `method` is set to `map`.

## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_repo` which is where you would find the latest commit from each
    input repo you specified.
  - Each input repo will be found here by name
  - Note: Unlike when mounting locally for debugging, there is no `Commit` ID
  in the path. This is because the commit will always change, and the ID isn't
  relevant to the processing. The commit that is exposed is configured based on
  the `incrementality` flag above
- `/pfs/out` which is where you write any output
- `/pfs/prev` which is this `Job` or `Pipeline`'s previous output, if it
    exists. (You can think of it as this job's output commit's parent).

### Output Formats

PFS supports data to be delimited by line, JSON, or binary blobs. [Refer here
for more information on
delimiters](../pachyderm_file_system.html#block-delimiters)

## Environment Variables

When the pipeline runs, the input and output commit IDs are exposed via
environment variables:

- `$PACH_OUTPUT_COMMIT_ID` contains the output commit of the job itself
- For each of the job's input repositories, there will be a corresponding
    environment variable w the input commid ID:
  - e.g. if there are two input repos `foo` and `bar`, the following will be
      populated:
    - `$PACH_FOO_COMMIT_ID`
    - `$PACH_BAR_COMMIT_ID`


## Flash-crowd behavior

In distributed systems, a flash-crowd behavior occurs when a large number of
nodes send traffic to a particular node in an uncoordinated fashion, causing
the node to become a hotspot, resulting in performance degradation.

To understand how such a behavior can occur in Pachyderm, it's important to
understand the way requests are sharded in a Pachyderm cluster.  Pachyderm
currently employs a simple sharding scheme that shards based on file names.
That is, requests pertaining to a certain file will be sent to a specific node.
As a result, if you have a number of nodes processing a large dataset in
parallel, it's advantageous for them to process files in a random order.

For instance, imagine that you have a dataset that contains `file_A`, `file_B`,
and `file_C`, each of of which is 1TB in size.  Now, each of your nodes will
get a portion of each of these files.  If your nodes independently start
processing files in alphanumeric order, they will all start with `file_A`,
causing all traffic to be sent to the node that handles `file_A`.  In contrast,
if your nodes process files in a random order, traffic will be distributed
between three nodes.



```
./pachctl create-pipeline -f pipeline.json
```

### Options

```
  -f, --file string       The file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
      --password string   Your password for the registry being pushed to.
  -p, --push-images       If true, push local docker images into the cluster registry.
  -r, --registry string   The registry to push images to. (default "docker.io")
  -u, --username string   The username to push images as, defaults to your OS username.
```

### Options inherited from parent commands

```
      --no-metrics   Don't report user metrics for this command
  -v, --verbose      Output verbose logs
```

### SEE ALSO
* [./pachctl](./pachctl.md)	 - 

###### Auto generated by spf13/cobra on 6-Mar-2017
