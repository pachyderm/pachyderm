## ./pachctl update-pipeline

Update an existing Pachyderm pipeline.

### Synopsis


Update a Pachyderm pipeline with a new spec

# Pipeline Specification

This document discusses each of the fields present in a pipeline specification.
To see how to use a pipeline spec, refer to the [pachctl
create-pipeline](../pachctl/pachctl_create-pipeline.html) doc.

## JSON Manifest Format

```json
{
  "pipeline": {
    "name": string
  },
  "transform": {
    "image": string,
    "cmd": [ string ],
    "stdin": [ string ]
    "env": {
        string: string
    },
    "secrets": [ {
        "name": string,
        "mountPath": string
    } ],
    "imagePullSecrets": [ string ],
    "acceptReturnCode": [ int ]
  },
  "parallelism_spec": {
    "strategy": "CONSTANT"|"COEFFICIENT"
    "constant": int        // if strategy == CONSTANT
    "coefficient": double  // if strategy == COEFFICIENT
  },
  "inputs": [
    {
      "name": string,
      "repo": {
        "name": string
      },
      "branch": string,
      "glob": string,
      "lazy": bool,
      "from": {
        "repo": {
          "name": string
        },
        ID: string
      }
    }
  ],
  "outputBranch": string,
  "egress": {
    "URL": "s3://bucket/dir"
  }
}
```

In practice, you rarely need to specify all the fields.  Most fields either come with sensible defaults or can be nil.  Following is an example of a minimal spec:

```json
{
  "pipeline": {
    "name": "wordcount"
  },
  "transform": {
    "image": "wordcount-image",
    "cmd": ["/binary", "/pfs/data", "/pfs/out"]
  },
  "inputs": [
    {
      "repo": {
        "name": "data"
      },
      "glob": "/*"
    }
  ]
}
```

Following is a walk-through of all the fields.

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

`transform.acceptReturnCode` is an array of return codes (i.e. exit codes)
from your docker command that are considered acceptable, which means that
if your docker command exits with one of the codes in this array, it will
be considered a successful run for the purpose of setting job status.  `0`
is always considered a successful exit code.

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

By default, we use the parallelism spec "coefficient=1", which means that
we spawn one worker per node for this pipeline.

### Inputs (optional)

`inputs` specifies a set of Repos that will be visible to the jobs during
runtime. Commits to these repos will automatically trigger the pipeline to
create new jobs to process them.

`inputs.name` is the name of the input.  An input with name `XXX` will be
visible under the path `/pfs/XXX` when a job runs.  Input names must be
unique.  If an input's name is not specified, it's default to the name of
the repo.

`inputs.repo` is a repo that contains input data for this pipeline.

`inputs.branch` is the name of a branch in the input repo.  Only commits on
this branch trigger the pipeline.  By default, it's set to `master`.

`inputs.glob` is a glob pattern that's used to determine how the input data
is partitioned.  It's explained in detail in the next section.

`inputs.lazy` controls how the data is exposed to jobs. The default is
`false` which means the job will eagerly download the data it needs to
process and it will be exposed as normal files on disk. If lazy is set
to `true`, data will be exposed as named pipes instead and no data will
be downloaded until the job opens the pipe and reads it, if the pipe is
never opened then no will be downloaded. Some applications won't work with
pipes, for example if they make syscalls such as `Seek` which pipes don't
support. Applications that can work with pipes should use them since they're
more performant, the difference will be especially notable if the job only 
reads a subset of the files that are available to it.

`inputs.from` specifies the starting point of the input branch.  If `from`
is not specified, then the entire input branch will be processed.  Otherwise,
only commits since the `from` commit (not including the `from` commit itself)
 will be processed.

### OutputBranch

This is the branch where the pipeline outputs new commits.  By default,
it's "master".

### Egress (optional)

`egress` allows you to push the results of a Pipeline to an external data
store such as s3, Google Cloud Storage or Azure Storage. Data will be pushed
after the user code has finished running but before the job is marked as
successful.

## The Input Glob Pattern

Each input needs to specify a [glob pattern](http://man7.org/linux/man-pages/man7/glob.7.html).

Pachyderm uses the glob pattern to determine how many "datums" an input
consists of.  Datums are the unit of parallelism in Pachyderm.  That is,
Pachyderm attemps to process datums in parallel whenever possible.

Intuitively, you may think of the input repo as a file system, and you are
applying the glob pattern to the root of the file system.  The files and
directories that match the glob pattern are considered datums.

For instance, let's say your input repo has the following structure:

```
/foo-1
/foo-2
/bar
  /bar-1
  /bar-2
```

Now let's consider what the following glob patterns would match respectively:

* `/`: this pattern matches `/`, the root directory itself
* `/*`:  this pattern matches everything under the root directory:
`/foo-1`, `/foo-2`, and `/bar`
* `/foo*`:  this pattern matches files under the root directory that start
with `/foo`: `/foo-1` and `/foo-2`
* `/*/*`:  this pattern matches everything that's two levels deep relative
to the root: `/bar/bar-1` and `/bar/bar-2`

Whatever that matches the glob pattern is a datum.  For instance, if we used
`/*`, then the job will process three datums (potentially in parallel):
`/foo-1`, `/foo-2`, and `/bar`.

## Multiple Inputs

A pipeline is allowed to have multiple inputs.  The important thing to
understand is what happens when a new commit comes into one of the input repos.

In short, a pipeline processes the **cross product** of the datums in its
inputs.  We will
use an example to illustrate.

Consider a pipeline that has two input repos: `foo` and `bar`.

```
1. PUT /file-1 in commit1 in foo -- no jobs triggered
2. PUT /file-a in commit1 in bar -- triggers job1
3. PUT /file-2 in commit2 in foo -- triggers job2
4. PUT /file-b in commit2 in bar -- triggers job3
```

The first time the pipeline is triggered will be when the second event
completes.  This is because we need data in both repos before we can run the
pipeline.

Here is a breakdown of the datums that each job sees:

```
job1:

Datum1:

    /pfs/foo/file-1
    /pfs/bar/file-a

job2:

Datum1:
    /pfs/foo/file-1
    /pfs/bar/file-a

Datum2:
    /pfs/foo/file-2
    /pfs/bar/file-a

job3:

Datum1:
    /pfs/foo/file-1
    /pfs/bar/file-a

Datum2:
    /pfs/foo/file-2
    /pfs/bar/file-a

Datum3:
    /pfs/foo/file-1
    /pfs/bar/file-b

Datum4:
    /pfs/foo/file-2
    /pfs/bar/file-b
```

## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_name` which is where you would find the datum.
  - Each input will be found here by its name, which defaults to the repo
  name if not specified.
- `/pfs/out` which is where you write any output.

### Output Formats

PFS supports data to be delimited by line, JSON, or binary blobs. [Refer here
for more information on
delimiters](../pachyderm_file_system.html#block-delimiters)


```
./pachctl update-pipeline -f pipeline.json
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

###### Auto generated by spf13/cobra on 14-Mar-2017
