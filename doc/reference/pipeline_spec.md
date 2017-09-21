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
  "description": string,
  "transform": {
    "image": string,
    "cmd": [ string ],
    "stdin": [ string ]
    "env": {
        string: string
    },
    "secrets": [ {
        "name": string,
        "mount_path": string
    } ],
    "image_pull_secrets": [ string ],
    "accept_return_code": [ int ]
  },
  "parallelism_spec": {
    // Set at most one of the following:
    "constant": int
    "coefficient": double
  },
  "resource_spec": {
    "memory": string
    "cpu": double
  },
  "input": {
    <"atom" or "cross" or "union", see below>
  },
  "output_branch": string,
  "egress": {
    "URL": "s3://bucket/dir"
  },
  "scale_down_threshold": string,
  "incremental": bool,
  "cache_size": string,
  "enable_stats": bool,
  "service": {
    "internal_port": int,
    "external_port": int
  }
}

------------------------------------
"atom" input
------------------------------------

"atom": {
  "name": string,
  "repo": string,
  "branch": string,
  "glob": string,
  "lazy" bool,
  "from_commit": string
}

------------------------------------
"cross" or "union" input
------------------------------------

"cross" or "union": [
  {
    "atom": {
      "name": string,
      "repo": string,
      "branch": string,
      "glob": string,
      "lazy" bool,
      "from_commit": string
    }
  },
  {
    "atom": {
      "name": string,
      "repo": string,
      "branch": string,
      "glob": string,
      "lazy" bool,
      "from_commit": string
    }
  }
  etc...
]

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
  "input": {
        "atom": {
            "repo": "data",
            "glob": "/*"
        }
    }
}
```

Following is a walk-through of all the fields.

### Name (required)

`pipeline.name` is the name of the pipeline that you are creating.  Each
pipeline needs to have a unique name.

### Description (optional)

`description` is an optional text field where you can put documentation about the pipeline.

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
[here](https://kubernetes.io/docs/concepts/configuration/secret/).

`transform.image_pull_secrets` is an array of image pull secrets, image pull
secrets are similar to secrets except that they're mounted before the
containers are created so they can be used to provide credentials for image
pulling. For example, if you are using a private Docker registry for your
images, you can specify it via:

```sh
$ kubectl create secret docker-registry myregistrykey --docker-server=DOCKER_REGISTRY_SERVER --docker-username=DOCKER_USER --docker-password=DOCKER_PASSWORD --docker-email=DOCKER_EMAIL
```

And then tell your pipeline about it via `"image_pull_secrets": [ "myregistrykey" ]`. Read more about image pull secrets
[here](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod).

`transform.accept_return_code` is an array of return codes (i.e. exit codes)
from your docker command that are considered acceptable, which means that
if your docker command exits with one of the codes in this array, it will
be considered a successful run for the purpose of setting job status.  `0`
is always considered a successful exit code.

### Parallelism Spec (optional)

`parallelism_spec` describes how Pachyderm should parallelize your pipeline.
Currently, Pachyderm has two parallelism strategies: `constant` and
`coefficient`.

If you set the `constant` field, Pachyderm will start the number of workers
that you specify. For example, set `"constant":10` to use 10 workers.

If you set the `coefficient` field, Pachyderm will start a number of workers
that is a multiple of your Kubernetes cluster’s size. For example, if your
Kubernetes cluster has 10 nodes, and you set `"coefficient": 0.5`, Pachyderm
will start five workers. If you set it to 2.0, Pachyderm will start 20 workers
(two per Kubernetes node).

By default, we use the parallelism spec "coefficient=1", which means that
we spawn one worker per node for this pipeline.

### Resource Spec (optional)

`resource_spec` describes the amount of resources you expect the
workers for a given pipeline to consume. Knowing this in advance
lets us schedule big jobs on separate machines, so that they don't
conflict and either slow down or die.

The `memory` field is a string that describes the amount of memory, in bytes,
each worker needs (with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc). For
example, a worker that needs to read a 1GB file into memory might set
`"memory": "1.2G"` (with a little extra for the code to use in addition to the
file. Workers for this pipeline will only be placed on machines with at least
1.2GB of free memory, and other large workers will be prevented from using it
(if they also set their `resource_spec`).

The `cpu` field is a double that describes the amount of CPU time (in (cpu
seconds)/(real seconds) each worker needs. Setting `"cpu": 0.5` indicates that
the worker should get 500ms of CPU time per second. Setting `"cpu": 2`
indicates that the worker should get 2000ms of CPU time per second (i.e. it's
using 2 CPUs, essentially, though worker threads might spend e.g. 500ms on four
physical CPUs instead of one second on two physical CPUs).

In both cases, the resource requests are not upper bounds. If the worker uses
more memory than it's requested, it will not (necessarily) be killed.  However,
if the whole node runs out of memory, Kubernetes will start killing pods that
have been placed on it and exceeded their memory request, to reclaim memory.
To prevent your worker getting killed, you must set your `memory` request to
a sufficiently large value. However, if the total memory requested by all
workers in the system is too large, Kubernetes will be unable to schedule new
workers (because no machine will have enough unclaimed memory). `cpu` works
similarly, but for CPU time.

By default, workers are scheduled with an effective resource request of 0 (to
avoid scheduling problems that prevent users from being unable to run
pipelines).  This means that if a node runs out of memory, any such worker
might be killed.

### Input (required)

`input` specifies repos that will be visible to the jobs during runtime.
Commits to these repos will automatically trigger the pipeline to create new
jobs to process them. Input is a recursive type, there are multiple different
kinds of inputs which can be combined together. The `input` object is a
container for the different input types with a field for each, only one of
these fields be set for any insantiation of the object.

```
{
    "atom": atom_input,
    "union": [input],
    "cross": [input],
}
```

#### Atom Input
Atom inputs are the simplest inputs, they take input from a single branch on a
single repo.

```
{
    "name": string,
    "repo": string,
    "branch": string,
    "glob": string,
    "lazy" bool,
    "from_commit": string
}
```

`input.atom.name` is the name of the input.  An input with name `XXX` will be visible
under the path `/pfs/XXX` when a job runs.  Input names must be unique. If an
input's name is not specified, it defaults to the name of the repo. Therefore,
if you have two inputs from the same repo, you'll need to give at least one
of them a unique name.

`input.atom.repo` is the `repo` to be used for the input.

`input.atom.branch` is the `branch` to watch for commits on, it may be left blank in
which case `"master"` will be used.

`input.atom.commit` is the `repo` and `branch` (specified as `id`) to be used for the
input, `repo` is required but `id` may be left blank in which case `"master"`
will be used.

`input.atom.glob` is a glob pattern that's used to determine how the input data
is partitioned.  It's explained in detail in the next section.

`input.atom.lazy` controls how the data is exposed to jobs. The default is `false`
which means the job will eagerly download the data it needs to process and it
will be exposed as normal files on disk. If lazy is set to `true`, data will be
exposed as named pipes instead and no data will be downloaded until the job
opens the pipe and reads it, if the pipe is never opened then no data will be
downloaded. Some applications won't work with pipes, for example if they make
syscalls such as `Seek` which pipes don't support. Applications that can work
with pipes should use them since they're more performant, the difference will
be especially notable if the job only reads a subset of the files that are
available to it.  Note that `lazy` currently doesn't support datums that
contain more than 10000 files.

`input.atom.from_commit` specifies the starting point of the input branch.  If
`from_commit` is not specified, then the entire input branch will be
processed.  Otherwise, only commits since the `from_commit` (not including
the commit itself) will be processed.

#### Union Input

Union inputs take the union of other inputs. For example:

```
| inputA | inputB | inputA ∪ inputB |
| ------ | ------ | --------------- |
| foo    | fizz   | foo             |
| bar    | buzz   | fizz            |
|        |        | bar             |
|        |        | buzz            |
```

Notice that union inputs, do not take a name and maintain the names of the
sub-inputs. In the above example you would see files under
`/pfs/inputA/...` or `/pfs/inputB/...`, but never both at the same time.
This can be annoying to write code for since the first thing your code
needs to do is figure out which input directory is present. As of 1.5.3
the recommended way to fix this is to give your inputs the same `Name`,
that way your code only needs to handle data being present in that
directory. This, of course, only works if your code doesn't need to be
aware of which of the underlying inputs the data comes from.

`input.union` is an array of inputs to union, note that these need not be
`atom` inputs, they can also be `union` and `cross` inputs. Although there's no
reason to take a union of unions since union is associative.

#### Cross Input

Cross inputs take the cross product of other inputs, in other words it creates
tuples of the datums in the inputs. For example:

```
| inputA | inputB | inputA ⨯ inputB |
| ------ | ------ | --------------- |
| foo    | fizz   | (foo, fizz)     |
| bar    | buzz   | (foo, buzz)     |
|        |        | (bar, fizz)     |
|        |        | (bar, buzz)     |
```

Notice that cross inputs, do not take a name and maintain the names of the sub-inputs.
In the above example you would see files under `/pfs/inputA/...` and `/pfs/inputB/...`.

`input.cross` is an array of inputs to cross, note that these need not be
`atom` inputs, they can also be `union` and `cross` inputs. Although there's no
reason to take a cross of crosses since cross products are associative.

#### Cron Input

Cron inputs allow you to trigger pipelines based on time. It's based on the
unix utility `cron`. When you create a pipeline with one or more Cron Inputs
pachd will create a repo for each of them. When a cron input triggers,
that is when the present time satisfies its spec, pachd will commit
a single file, called "time" to the repo which contains the time which
satisfied the spec. The time is formatted according to [RFC
3339](https://www.ietf.org/rfc/rfc3339.txt).

```
{
    "name": string,
    "spec": string,
    "repo": string,
    "start": time,
}
```

`input.cron.name` is the name for the input, its semantics are similar to
those of `input.atom.name`. Except that it's not optional.

`input.cron.spec` is a cron expression which specifies the schedule on
which to trigger the pipeline. To learn more about how to write schedules
see the [Wikipedia page on cron](https://en.wikipedia.org/wiki/Cron).
Pachyderm supports Nonstandard schedules such as `"@daily"`.

`input.cron.repo` is the repo which will be created for the input. It is
optional, if it's not specified then `"<pipeline-name>_<input-name>"` will
be used.

`input.cron.start` is the time to start counting from for the input. It is
optional, if it's not specified then the present time (when the pipeline
is created) will be used. Specifying a time allows you to run on matching
times from the past or, skip times from the present and only start running
on matching times in the future. Times should be formatted according to [RFC
3339](https://www.ietf.org/rfc/rfc3339.txt).

### OutputBranch (optional)

This is the branch where the pipeline outputs new commits.  By default,
it's "master".

### Egress (optional)

`egress` allows you to push the results of a Pipeline to an external data
store such as s3, Google Cloud Storage or Azure Storage. Data will be pushed
after the user code has finished running but before the job is marked as
successful.

### Scale-down threshold (optional)

`scale_down_threshold` specifies when the worker pods of a pipeline should be terminated.

by default, a pipeline’s worker pods are always running.  when `scale_down_threshold` is set, all but one worker pods are terminated after the pipeline has not seen a new job for the given duration (we still need one worker pod to subscribe to new input commits).  when a new input commit comes in, the worker pods are then re-created.

`scale_down_threshold` is a string that needs to be sequence of decimal numbers with a unit suffix, such as “300ms”, “1.5h” or “2h45m”. valid time units are “s”, “m”, “h”.

### Incremental (optional)

Incremental, if set will cause the pipeline to be run "incrementally". This
means that when a datum changes it won't be reprocessed from scratch, instead
`/pfs/out` will be populated with the previous results of processing that datum
and instead of seeing the full datum under `/pfs/repo` you will see only
new/modified values.

Incremental processing is useful for [online
algorithms](https://en.wikipedia.org/wiki/Online_algorithm), a canonical
example is summing a set of numbers since the new numbers can be added to the
old total without having to reconsider the numbers which went into that old
total. Incremental is design to work nicely with the `--split` flag to
`put-file` because it will cause only the new chunks of the file to be
displayed to each step of the pipeline.

### Cache Size (optional)

`cache_size` controls how much cache a pipeline worker uses.  In general,
your pipeline's performance will increase with the cache size, but only
up to a certain point depending on your workload.

### Enable Stats (optional)

`enable_stats` turns on stat tracking for the pipeline. This will cause the
pipeline to commit to a second branch in its output repo called `"stats"`. This
branch will have information about each datum that is processed including:
timing information, size information, logs and a `/pfs` snapshot. This
information can be accessed through the `inspect-datum` and `list-datum`
pachctl commands and through the webUI.

Note: enabling stats will use extra storage for logs and timing information.
However it will not use as much extra storage as it appears to due to the fact
that snapshots of the `/pfs` directory, which are generally the largest thing
stored, don't actually require extra storage because the data is already stored
in the input repos.

### Service (optional)

`service` specifies that the pipeline should be treated as a long running
service rather than a data transformation. This means that `transform.cmd` is
not expected to exit, if it does it will be restarted. Furthermore, the service
will be exposed outside the container using a kubernetes service.
`"internal_port"` should be a port that the user code binds to inside the
container, `"external_port"` is the port on which it is exposed, via the
NodePorts functionality of kubernetes services. After a service has been
created you should be able to access it at
`http://<kubernetes-host>:<external_port>`.

## The Input Glob Pattern

Each atom input needs to specify a [glob pattern](../fundamentals/distributed_computing.html).

Pachyderm uses the glob pattern to determine how many "datums" an input
consists of.  Datums are the unit of parallelism in Pachyderm.  That is,
Pachyderm attempts to process datums in parallel whenever possible.

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

* `/`: this pattern matches `/`, the root directory itself, meaning all the data would be a single large datum.
* `/*`:  this pattern matches everything under the root directory given us 3 datums:
`/foo-1.`, `/foo-2.`, and everything under the directory `/bar`.
* `/bar/*`: this pattern matches files only under the `/bar` directory: `/bar-1` and `/bar-2`
* `/foo*`:  this pattern matches files under the root directory that start with the characters `foo`
* `/*/*`:  this pattern matches everything that's two levels deep relative
to the root: `/bar/bar-1` and `/bar/bar-2`

The datums are defined as whichever files or directories match by the glob pattern. For instance, if we used
`/*`, then the job will process three datums (potentially in parallel):
`/foo-1`, `/foo-2`, and `/bar`. Both the `bar-1` and `bar-2` files within the directory `bar` would be grouped together and always processed by the same worker.

## Multiple Inputs

It's important to note that if a pipeline takes multiple atom inputs (via cross
or union) then the pipeline will not get triggered until all of the atom inputs
have at least one commit on the branch.

## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_name` which is where you would find the datum.
  - Each input will be found here by its name, which defaults to the repo
  name if not specified.
- `/pfs/out` which is where you write any output.

### Output Formats

PFS supports data to be delimited by line, JSON, or binary blobs.
