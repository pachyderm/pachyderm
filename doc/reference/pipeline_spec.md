# Pipeline Specification

This document discusses each of the fields present in a pipeline specification.
To see how to use a pipeline spec to create a pipeline, refer to the [pachctl
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
    },
    {
        "name": string,
        "env_var": string,
        "key": string
    } ],
    "image_pull_secrets": [ string ],
    "accept_return_code": [ int ],
    "debug": bool,
    "user": string,
    "working_dir": string,
  },
  "parallelism_spec": {
    // Set at most one of the following:
    "constant": int,
    "coefficient": number
  },
  "resource_requests": {
    "memory": string,
    "cpu": number,
    "disk": string,
  },
  "resource_limits": {
    "memory": string,
    "cpu": number,
    "gpu": number,
    "disk": string,
  },
  "datum_timeout": string,
  "datum_tries": int,
  "job_timeout": string,
  "input": {
    <"pfs", "cross", "union", "cron", or "git" see below>
  },
  "output_branch": string,
  "egress": {
    "URL": "s3://bucket/dir"
  },
  "standby": bool,
  "cache_size": string,
  "enable_stats": bool,
  "service": {
    "internal_port": int,
    "external_port": int
  },
  "max_queue_size": int,
  "chunk_spec": {
    "number": int,
    "size_bytes": int
  },
  "scheduling_spec": {
    "node_selector": {string: string},
    "priority_class_name": string
  },
  "pod_spec": string
}

------------------------------------
"pfs" input
------------------------------------

"pfs": {
  "name": string,
  "repo": string,
  "branch": string,
  "glob": string,
  "lazy" bool,
  "empty_files": bool
}

------------------------------------
"cross" or "union" input
------------------------------------

"cross" or "union": [
  {
    "pfs": {
      "name": string,
      "repo": string,
      "branch": string,
      "glob": string,
      "lazy" bool,
      "empty_files": bool
    }
  },
  {
    "pfs": {
      "name": string,
      "repo": string,
      "branch": string,
      "glob": string,
      "lazy" bool,
      "empty_files": bool
    }
  }
  etc...
]

------------------------------------
"cron" input
------------------------------------

"cron": {
    "name": string,
    "spec": string,
    "repo": string,
    "start": time
}

------------------------------------
"git" input
------------------------------------

"git": {
  "URL": string,
  "name": string,
  "branch": string
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
  "input": {
        "pfs": {
            "repo": "data",
            "glob": "/*"
        }
    }
}
```

Following is a walk-through of all the fields.

### Name (required)

`pipeline.name` is the name of the pipeline that you are creating.  Each
pipeline needs to have a unique name. Pipeline names must:

- contain only alphanumeric characters, `_` and `-`
- begin or end with only alphanumeric characters (not `_` or `-`)
- be no more than 50 characters in length

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

Note: there are environment variables that are automatically injected into the
container, for a comprehensive list of them see the [Environment
Variables](#environment-variables) section below.

`transform.secrets` is an array of secrets, they are useful for embedding
sensitive data such as credentials. Secrets reference Kubernetes secrets by
name and specify a path that the secrets should be mounted to, or an
environment variable (`env_var`) that the value should be bound to. Secrets
must set `name` which should be the name of a secret in Kubernetes. Secrets
must also specify either `mount_path` or `env_var` and `key`. See more information about kubernetes secrets [here](https://kubernetes.io/docs/concepts/configuration/secret/).

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

`transform.debug` turns on added debug logging for the pipeline.

`transform.user` sets the user that your code runs as, this can also be
accomplished with a `USER` directive in your Dockerfile.

`transform.working_dir` sets the directory that your command will be run from,
this can also be accomplished with a `WORKDIR` directive in your Dockerfile.

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

### Resource Requests (optional)

`resource_requests` describes the amount of resources you expect the
workers for a given pipeline to consume. Knowing this in advance
lets us schedule big jobs on separate machines, so that they don't
conflict and either slow down or die.

The `memory` field is a string that describes the amount of memory, in bytes,
each worker needs (with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc). For
example, a worker that needs to read a 1GB file into memory might set
`"memory": "1.2G"` (with a little extra for the code to use in addition to the
file. Workers for this pipeline will only be placed on machines with at least
1.2GB of free memory, and other large workers will be prevented from using it
(if they also set their `resource_requests`).

The `cpu` field is a number that describes the amount of CPU time (in (cpu
seconds)/(real seconds) each worker needs. Setting `"cpu": 0.5` indicates that
the worker should get 500ms of CPU time per second. Setting `"cpu": 2`
indicates that the worker should get 2000ms of CPU time per second (i.e. it's
using 2 CPUs, essentially, though worker threads might spend e.g. 500ms on four
physical CPUs instead of one second on two physical CPUs).

The `disk` field is a string that describes the amount of ephemeral disk space,
in bytes, each worker needs (with allowed SI suffixes (M, K, G, Mi, Ki, Gi,
etc).

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

For more information about resource requests and limits see the
[Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/)
on the subject.

### Resource Limits (optional)

`resource_limits` describes the upper threshold of allowed resources a given 
worker can consume. If a worker exceeds this value, it will be evicted.

The `gpu` field is a number that describes how many GPUs each worker needs.
Only whole number are supported, Kubernetes does not allow multiplexing of
GPUs. Unlike the other resource fields, GPUs only have meaning in Limits, by
requesting a GPU the worker will have sole access to that GPU while it is
running. It's recommended to enable `standby` if you are using GPUs so other
processes in the cluster will have access to the GPUs while the pipeline has
nothing to process. For more information about scheduling GPUs see the
[Kubernetes docs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/)
on the subject.

### Datum Timeout (optional)

`datum_timeout` is a string (e.g. `1s`, `5m`, or `15h`) that determines the 
maximum execution time allowed per datum. So no matter what your parallelism
or number of datums, no single datum is allowed to exceed this value.

### Datum Tries (optional)

`datum_tries` is a int (e.g. `1`, `2`, or `3`) that determines the number of retries that a job should attempt given failure was observed. Only failed datums are retries in retry attempt. The the operation succeeds in retry attempts then job is successful, otherwise the job is marked as failure.


### Job Timeout (optional)

`job_timeout` is a string (e.g. `1s`, `5m`, or `15h`) that determines the 
maximum execution time allowed for a job. It differs from `datum_timeout`
in that the limit gets applied across all workers and all datums. That 
means that you'll need to keep in mind the parallelism, total number of
datums, and execution time per datum when setting this value. Keep in 
mind that the number of datums may change over jobs. Some new commits may
have a bunch of new files (and so new datums). Some may have fewer.

### Input (required)

`input` specifies repos that will be visible to the jobs during runtime.
Commits to these repos will automatically trigger the pipeline to create new
jobs to process them. Input is a recursive type, there are multiple different
kinds of inputs which can be combined together. The `input` object is a
container for the different input types with a field for each, only one of
these fields be set for any instantiation of the object.

```
{
    "pfs": pfs_input,
    "union": [input],
    "cross": [input],
    "cron": cron_input
}
```

#### PFS Input
PFS inputs are the simplest inputs, they take input from a single branch on a
single repo.

```
{
    "name": string,
    "repo": string,
    "branch": string,
    "glob": string,
    "lazy" bool,
    "empty_files": bool
}
```

`input.pfs.name` is the name of the input.  An input with name `XXX` will be
visible under the path `/pfs/XXX` when a job runs.  Input names must be unique
if the inputs are crossed, but they may be duplicated between `PfsInput`s that are unioned.  This is because when `PfsInput`s are unioned, you'll only ever see a datum from one input at a time. Overlapping the names of unioned inputs allows
you to write simpler code since you no longer need to consider which input directory a particular datum come from.  If an input's name is not specified, it defaults to the name of the repo.  Therefore, if you have two crossed inputs from the same repo, you'll be required to give at least one of them a unique name.

`input.pfs.repo` is the `repo` to be used for the input.

`input.pfs.branch` is the `branch` to watch for commits on, it may be left blank in
which case `"master"` will be used.

`input.pfs.glob` is a glob pattern that's used to determine how the input data
is partitioned.  It's explained in detail in the next section.

`input.pfs.lazy` controls how the data is exposed to jobs. The default is `false`
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

`input.pfs.empty_files` controls how files are exposed to jobs. If true, it will 
cause files from this PFS input to be presented as empty files. This is useful in shuffle 
pipelines where you want to read the names of files and reorganize them using symlinks.

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
`pfs` inputs, they can also be `union` and `cross` inputs. Although there's no
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
`pfs` inputs, they can also be `union` and `cross` inputs. Although there's no
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
those of `input.pfs.name`. Except that it's not optional.

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

#### Git Input (alpha feature)

Git inputs allow you to pull code from a public git URL and execute that code as part of your pipeline. A pipeline with a Git Input will get triggered (i.e. will see a new input commit and will spawn a job) whenever you commit to your git repository. 

**Note:** This only works on cloud deployments, not local clusters.

`input.git.URL` must be a URL of the form: `https://github.com/foo/bar.git`

`input.git.name` is the name for the input, its semantics are similar to
those of `input.pfs.name`. It is optional.

`input.git.branch` is the name of the git branch to use as input

Git inputs also require some additional configuration. In order for new commits on your git repository to correspond to new commits on the Pachyderm Git Input repo, we need to setup a git webhook. At the moment, only GitHub is supported. (Though if you ask nicely, we can add support for GitLab or BitBucket).

1. Create your Pachyderm pipeline with the Git Input.

2. To get the URL of the webhook to your cluster, do `pachctl inspect-pipeline` on your pipeline. You should see a `Githook URL` field with a URL set. Note - this will only work if you've deployed to a cloud provider (e.g. AWS, GKE). If you see `pending` as the value (and you've deployed on a cloud provider), it's possible that the service is still being provisioned. You can check `kubectl get svc` to make sure you see the `githook` service running.

3. To setup the GitHub webhook, navigate to:

```
https://github.com/<your_org>/<your_repo>/settings/hooks/new
```
Or navigate to webhooks under settings. Then you'll want to copy the `Githook URL` into the 'Payload URL' field.

### Output Branch (optional)

This is the branch where the pipeline outputs new commits.  By default,
it's "master".

### Egress (optional)

`egress` allows you to push the results of a Pipeline to an external data
store such as s3, Google Cloud Storage or Azure Storage. Data will be pushed
after the user code has finished running but before the job is marked as
successful.

### Standby (optional)

`standby` indicates that the pipeline should be put into "standby" when there's
no data for it to process.  A pipeline in standby will have no pods running and
thus will consume no resources, it's state will be displayed as "standby".

Standby replaces `scale_down_threshold` from releases prior to 1.7.1.

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

### Service (alpha feature, optional)

`service` specifies that the pipeline should be treated as a long running
service rather than a data transformation. This means that `transform.cmd` is
not expected to exit, if it does it will be restarted. Furthermore, the service
will be exposed outside the container using a kubernetes service.
`"internal_port"` should be a port that the user code binds to inside the
container, `"external_port"` is the port on which it is exposed, via the
NodePorts functionality of kubernetes services. After a service has been
created you should be able to access it at
`http://<kubernetes-host>:<external_port>`.

### Max Queue Size (optional)
`max_queue_size` specifies that maximum number of elements that a worker should
hold in its processing queue at a given time. The default value is `1` which
means workers will only hold onto the value that they're currently processing.
Increasing this value can improve pipeline performance as it allows workers to
simultaneously download, process and upload different datums at the same time.
Setting this value too high can cause problems if you have `lazy` inputs as
there's a cap of 10,000 `lazy` files per worker and multiple datums that are
running all count against this limit.

### Chunk Spec (optional)
`chunk_spec` specifies how a pipeline should chunk its datums.

`chunk_spec.number` if nonzero, specifies that each chunk should contain `number`
 datums. Chunks may contain fewer if the total number of datums don't
 divide evenly.

`chunk_spec.size_bytes` , if nonzero, specifies a target size for each chunk of datums.
 Chunks may be larger or smaller than `size_bytes`, but will usually be
 pretty close to `size_bytes` in size.

### Scheduling Spec (optional)
`scheduling_spec` specifies how the pods for a pipeline should be scheduled.

`scheduling_spec.node_selector` allows you to select which nodes your pipeline
will run on. Refer to the [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector)
on node selectors for more information about how this works.

`scheduling_spec.priority_class_name` allows you to select the prioriy class
for the pipeline, which will how Kubernetes chooses to schedule and deschedule
the pipeline. Refer to the [Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass)
on priority and preemption for more information about how this works.

### Pod Spec (optional)
`pod_spec` is an advanced option that allows you to set fields in the pod spec
that haven't been explicitly exposed in the rest of the pipeline spec. A good
way to figure out what JSON you should pass is to create a pod in Kubernetes
with the proper settings, then do:

```
kubectl get po/<pod-name> -o json | jq .spec
```

this will give you a correctly formated piece of JSON, you should then remove
the extraneous fields that Kubernetes injects or that can be set else where.

The JSON is applied after the other parameters for the `pod_spec` have already
been set. This means that you can modify things such as the storage and user
containers.

## The Input Glob Pattern

Each PFS input needs to specify a [glob pattern](../fundamentals/distributed_computing.html).

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

## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_name` which is where you would find the datum.
  - Each input will be found here by its name, which defaults to the repo
  name if not specified.
- `/pfs/out` which is where you write any output.

# Environment Variables

There are several environment variables that get injected into the user code
before it runs. They are:

- `PACH_JOB_ID` the id the currently run job.
- `PACH_OUTPUT_COMMIT_ID` the id of the commit being outputted to.
- For each input there will be an environment variable with the same name
    defined to the path of the file for that input. For example if you are
    accessing an input called `foo` from the path `/pfs/foo` which contains a
    file called `bar` then the environment variable `foo` will have the value
    `/pfs/foo/bar`. The path in the environment variable is the path which
    matched the glob pattern, even if the file is a directory, ie if your glob
    pattern is `/*` it would match a directory `/bar`, the value of `$foo`
    would then be `/pfs/foo/bar`. With a glob pattern of `/*/*` you would match
    the files contained in `/bar` and thus the value of `foo` would be
    `/pfs/foo/bar/quux`.
- For each input there will be an environment variable named `input_COMMIT`
    indicating the id of the commit being used for that input.

In addition to these environment variables Kubernetes also injects others for
Services that are running inside the cluster. These allow you to connect to
those outside services, which can be powerful but also can be hard to reason
about, as processing might be retried multiple times. For example if your code
writes a row to a database that row may be written multiple times due to
retries. Interaction with outside services should be idempotent to prevent
unexpected behavior. Furthermore, one of the running services that your code
can connect to is Pachyderm itself, this is generally not recommended as very
little of the Pachyderm API is idempotent, but in some specific cases it can be
a viable approach.
