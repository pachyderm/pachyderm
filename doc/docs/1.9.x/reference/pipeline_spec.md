# Pipeline Specification

This document discusses each of the fields present in a pipeline specification.
To see how to use a pipeline spec to create a pipeline, refer to the [pachctl
create pipeline](pachctl/pachctl_create_pipeline.md) section.

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
    "stdin": [ string ],
    "err_cmd": [ string ],
    "err_stdin": [ string ],
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
  "hashtree_spec": {
   "constant": int,
  },
  "resource_requests": {
    "memory": string,
    "cpu": number,
    "disk": string,
  },
  "resource_limits": {
    "memory": string,
    "cpu": number,
    "gpu": {
      "type": string,
      "number": int
    }
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
  "spout": {
  "overwrite": bool
  \\ Optionally, you can combine a spout with a service:
  "service": {
        "internal_port": int,
        "external_port": int,
        "annotations": {
            "foo": "bar"
        }
    }
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
  "pod_spec": string,
  "pod_patch": string,
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
  ...
]



------------------------------------
"cron" input
------------------------------------

"cron": {
    "name": string,
    "spec": string,
    "repo": string,
    "start": time,
    "overwrite": bool
}

------------------------------------
"join" input
------------------------------------

"join": [
  {
    "pfs": {
      "name": string,
      "repo": string,
      "branch": string,
      "glob": string,
      "join_on": string
      "lazy": bool
      "empty_files": bool
    }
  },
  {
    "pfs": {
       "name": string,
       "repo": string,
       "branch": string,
       "glob": string,
       "join_on": string
       "lazy": bool
       "empty_files": bool
    }
  }
]

------------------------------------
"git" input
------------------------------------

"git": {
  "URL": string,
  "name": string,
  "branch": string
}

```

In practice, you rarely need to specify all the fields.
Most fields either come with sensible defaults or can be empty.
The following text is an example of a minimum spec:

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

### Name (required)

`pipeline.name` is the name of the pipeline that you are creating. Each
pipeline needs to have a unique name. Pipeline names must meet the following
requirements:

- Include only alphanumeric characters, `_` and `-`.
- Begin or end with only alphanumeric characters (not `_` or `-`).
- Not exceed 63 characters in length.

### Description (optional)

`description` is an optional text field where you can add information
about the pipeline.

### Transform (required)

`transform.image` is the name of the Docker image that your jobs use.

`transform.cmd` is the command passed to the Docker run invocation. Similarly
to Docker, `cmd` is not run inside a shell which means that
wildcard globbing (`*`), pipes (`|`), and file redirects (`>` and `>>`) do
not work. To specify these settings, you can set `cmd` to be a shell of your
choice, such as `sh` and pass a shell script to `stdin`.

`transform.stdin` is an array of lines that are sent to your command on
`stdin`.
Lines do not have to end in newline characters.

`transform.err_cmd` is an optional command that is executed on failed datums.
If the `err_cmd` is successful and returns 0 error code, it does not prevent
the job from succeeding.
This behavior means that `transform.err_cmd` can be used to ignore
failed datums while still writing successful datums to the output repo,
instead of failing the whole job when some datums fail. The `transform.err_cmd`
command has the same limitations as `transform.cmd`.

`transform.err_stdin` is an array of lines that are sent to your error command
on `stdin`.
Lines do not have to end in newline characters.

`transform.env` is a key-value map of environment variables that
Pachyderm injects into the container.

**Note:** There are environment variables that are automatically injected
into the container, for a comprehensive list of them see the [Environment
Variables](#environment-variables) section below.

`transform.secrets` is an array of secrets. You can use the secrets to
embed sensitive data, such as credentials. The secrets reference
Kubernetes secrets by name and specify a path to map the secrets or
an environment variable (`env_var`) that the value should be bound to. Secrets
must set `name` which should be the name of a secret in Kubernetes. Secrets
must also specify either `mount_path` or `env_var` and `key`. See more
information about Kubernetes secrets [here](https://kubernetes.io/docs/concepts/configuration/secret/).

`transform.image_pull_secrets` is an array of image pull secrets, image pull
secrets are similar to secrets except that they are mounted before the
containers are created so they can be used to provide credentials for image
pulling. For example, if you are using a private Docker registry for your
images, you can specify it by running the following command:

```shell
$ kubectl create secret docker-registry myregistrykey --docker-server=DOCKER_REGISTRY_SERVER --docker-username=DOCKER_USER --docker-password=DOCKER_PASSWORD --docker-email=DOCKER_EMAIL
```

And then, notify your pipeline about it by using
`"image_pull_secrets": [ "myregistrykey" ]`. Read more about image pull secrets
[here](https://kubernetes.io/docs/concepts/containers/images/#specifying-imagepullsecrets-on-a-pod).

`transform.accept_return_code` is an array of return codes, such as exit codes
from your Docker command that are considered acceptable.
If your Docker command exits with one of the codes in this array, it is
considered a successful run to set job status. `0`
is always considered a successful exit code.

`transform.debug` turns on added debug logging for the pipeline.

`transform.user` sets the user that your code runs as, this can also be
accomplished with a `USER` directive in your `Dockerfile`.

`transform.working_dir` sets the directory that your command runs from. You
can also specify the `WORKDIR` directive in your `Dockerfile`.

`transform.dockerfile` is the path to the `Dockerfile` used with the `--build`
flag. This defaults to `./Dockerfile`.

### Parallelism Spec (optional)

`parallelism_spec` describes how Pachyderm parallelizes your pipeline.
Currently, Pachyderm has two parallelism strategies: `constant` and
`coefficient`.

If you set the `constant` field, Pachyderm starts the number of workers
that you specify. For example, set `"constant":10` to use 10 workers.

If you set the `coefficient` field, Pachyderm starts a number of workers
that is a multiple of your Kubernetes cluster’s size. For example, if your
Kubernetes cluster has 10 nodes, and you set `"coefficient": 0.5`, Pachyderm
starts five workers. If you set it to 2.0, Pachyderm starts 20 workers
(two per Kubernetes node).

The default value is "constant=1" .

Because spouts and services are designed to be single instances, do not
modify the default `parallism_spec` value for these pipelines.

### Resource Requests (optional)

`resource_requests` describes the amount of resources you expect the
workers for a given pipeline to consume. Knowing this in advance
lets Pachyderm schedule big jobs on separate machines, so that they do not
conflict and either slow down or die.

The `memory` field is a string that describes the amount of memory, in bytes,
each worker needs (with allowed SI suffixes (M, K, G, Mi, Ki, Gi, and so on).
For example, a worker that needs to read a 1GB file into memory might set
`"memory": "1.2G"` with a little extra for the code to use in addition to the
file. Workers for this pipeline will be placed on machines with at least
1.2GB of free memory, and other large workers will be prevented from using it
(if they also set their `resource_requests`).

The `cpu` field is a number that describes the amount of CPU time in `cpu
seconds/real seconds` that each worker needs. Setting `"cpu": 0.5` indicates that
the worker should get 500ms of CPU time per second. Setting `"cpu": 2`
indicates that the worker gets 2000ms of CPU time per second. In other words,
it is using 2 CPUs, though worker threads might spend 500ms on four
physical CPUs instead of one second on two physical CPUs.

The `disk` field is a string that describes the amount of ephemeral disk space,
in bytes, each worker needs with allowed SI suffixes (M, K, G, Mi, Ki, Gi,
and so on).

In both cases, the resource requests are not upper bounds. If the worker uses
more memory than it is requested, it does not mean that it will be shut down.
However, if the whole node runs out of memory, Kubernetes starts deleting
pods that have been placed on it and exceeded their memory request,
to reclaim memory.
To prevent deletion of your worker node, you must set your `memory` request to
a sufficiently large value. However, if the total memory requested by all
workers in the system is too large, Kubernetes cannot schedule new
workers because no machine has enough unclaimed memory. `cpu` works
similarly, but for CPU time.

By default, workers are scheduled with an effective resource request of 0 (to
avoid scheduling problems that prevent users from being unable to run
pipelines). This means that if a node runs out of memory, any such worker
might be terminated.

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

`datum_tries` is an integer, such as `1`, `2`, or `3`, that determines the
number of times a job attempts to run on a datum when a failure occurs. 
Setting `datum_tries` to `1` will attempt a job once with no retries. 
Only failed datums are retried in a retry attempt. If the operation succeeds
in retry attempts, then the job is marked as successful. Otherwise, the job
is marked as failed.


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
    "union": union_input,
    "cross": cross_input,
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

`input.pfs.name` is the name of the input. An input with the name `XXX` is
visible under the path `/pfs/XXX` when a job runs. Input names must be unique
if the inputs are crossed, but they may be duplicated between `PFSInput`s that
are combined by using the `union` operator. This is because when
`PFSInput`s are combined, you only ever see a datum from one input
at a time. Overlapping the names of combined inputs allows
you to write simpler code since you no longer need to consider which
input directory a particular datum comes from. If an input's name is not
specified, it defaults to the name of the repo. Therefore, if you have two
crossed inputs from the same repo, you must give at least one of them a
unique name.

`input.pfs.repo` is the name of the Pachyderm repository with the data that
you want to join with other data.

`input.pfs.branch` is the `branch` to watch for commits. If left blank,
Pachyderm sets this value to `master`.

`input.pfs.glob` is a glob pattern that is used to determine how the
input data is partitioned.

`input.pfs.lazy` controls how the data is exposed to jobs. The default is
`false` which means the job eagerly downloads the data it needs to process and
exposes it as normal files on disk. If lazy is set to `true`, data is
exposed as named pipes instead, and no data is downloaded until the job
opens the pipe and reads it. If the pipe is never opened, then no data is
downloaded.

Some applications do not work with pipes. For example, pipes do not support
applications that makes `syscalls` such as `Seek`. Applications that can work
with pipes must use them since they are more performant. The difference will
be especially notable if the job only reads a subset of the files that are
available to it.

**Note:** `lazy` currently does not support datums that
contain more than 10000 files.

`input.pfs.empty_files` controls how files are exposed to jobs. If
set to `true`, it causes files from this PFS to be presented as empty files.
This is useful in shuffle pipelines where you want to read the names of
files and reorganize them by using symlinks.

#### Union Input

Union inputs take the union of other inputs. In the example
below, each input includes individual datums, such as if  `foo` and `bar`
were in the same repository with the glob pattern set to `/*`.
Alternatively, each of these datums might have come from separate repositories
with the glob pattern set to `/` and being the only filesystm objects in these
repositories.

```
| inputA | inputB | inputA ∪ inputB |
| ------ | ------ | --------------- |
| foo    | fizz   | foo             |
| bar    | buzz   | fizz            |
|        |        | bar             |
|        |        | buzz            |
```

The union inputs do not take a name and maintain the names of the
sub-inputs. In the example above, you would see files under
`/pfs/inputA/...` or `/pfs/inputB/...`, but never both at the same time.
When you write code to address this behavior, make sure that
your code first determines which input directory is present. Starting
with Pachyderm 1.5.3, we recommend that you give your inputs the
same `Name`. That way your code only needs to handle data being present
in that directory. This only works if your code does not need to be
aware of which of the underlying inputs the data comes from.

`input.union` is an array of inputs to combine. The inputs do not have to be
`pfs` inputs. They can also be `union` and `cross` inputs. Although, there is
no reason to take a union of unions because union is associative.

#### Cross Input

Cross inputs create the cross product of other inputs. In other words,
a cross input creates tuples of the datums in the inputs. In the example
below, each input includes individual datums, such as if  `foo` and `bar`
were in the same repository with the glob pattern set to `/*`.
Alternatively, each of these datums might have come from separate repositories
with the glob pattern set to `/` and being the only filesystm objects in these
repositories.

```
| inputA | inputB | inputA ⨯ inputB |
| ------ | ------ | --------------- |
| foo    | fizz   | (foo, fizz)     |
| bar    | buzz   | (foo, buzz)     |
|        |        | (bar, fizz)     |
|        |        | (bar, buzz)     |
```

The cross inputs above do not take a name and maintain
the names of the sub-inputs.
In the example above, you would see files under `/pfs/inputA/...`
and `/pfs/inputB/...`.

`input.cross` is an array of inputs to cross.
The inputs do not have to be `pfs` inputs. They can also be
`union` and `cross` inputs. Although, there is
 no reason to take a union of unions because union is associative.

#### Cron Input

Cron inputs allow you to trigger pipelines based on time. A Cron input is
based on the Unix utility called `cron`. When you create a pipeline with
one or more Cron inputs, `pachd` creates a repo for each of them. The start
time for Cron input is specified in its spec.
When a Cron input triggers,
`pachd` commits a single file, named by the current [RFC
3339 timestamp](https://www.ietf.org/rfc/rfc3339.txt) to the repo which
contains the time which satisfied the spec.

```
{
    "name": string,
    "spec": string,
    "repo": string,
    "start": time,
    "overwrite": bool
}
```

`input.cron.name` is the name for the input. Its semantics is similar to
those of `input.pfs.name`. Except that it is not optional.

`input.cron.spec` is a cron expression which specifies the schedule on
which to trigger the pipeline. To learn more about how to write schedules,
see the [Wikipedia page on cron](https://en.wikipedia.org/wiki/Cron).
Pachyderm supports non-standard schedules, such as `"@daily"`.

`input.cron.repo` is the repo which Pachyderm creates for the input. This
parameter is optional. If you do not specify this parameter, then
`"<pipeline-name>_<input-name>"` is used by default.

`input.cron.start` is the time to start counting from for the input. This
parameter is optional. If you do not specify this parameter, then the
time when the pipeline was created is used by default. Specifying a
time enables you to run on matching times from the past or skip times
from the present and only start running
on matching times in the future. Format the time value according to [RFC
3339](https://www.ietf.org/rfc/rfc3339.txt).

`input.cron.overwrite` is a flag to specify whether you want the timestamp file
to be overwritten on each tick. This parameter is optional, and if you do not
specify it, it defaults to simply writing new files on each tick. By default,
`pachd` expects only the new information to be written out on each tick
and combines that data with the data from the previous ticks. If `"overwrite"`
is set to `true`, it expects the full dataset to be written out for each tick and
replaces previous outputs with the new data written out.

#### Join Input

A join input enables you to join files that are stored in separate
Pachyderm repositories and that match a configured glob
pattern. A join input must have the `glob` and `join_on` parameters configured
to work properly. A join can combine multiple PFS inputs.

You can specify the following parameters for the `join` input.

* `input.pfs.name` — the name of the PFS input that appears in the
`INPUT` field when you run the `pachctl list job` command.
If an input name is not specified, it defaults to the name of the repo.

* `input.pfs.repo` — see the description in [PFS Input](#pfs-input).
the name of the Pachyderm repository with the data that
you want to join with other data.

* `input.pfs.branch` — see the description in [PFS Input](#pfs-input).

* `input.pfs.glob` — a wildcard pattern that defines how a dataset is broken
  up into datums for further processing. When you use a glob pattern in joins,
  it creates a naming convention that Pachyderm uses to join files. In other
  words, Pachyderm joins the files that are named according to the glob
  pattern and skips those that are not.

  You can specify the glob pattern for joins in a parenthesis to create
  one or multiple capture groups. A capture group can include one or multiple
  characters. Use standard UNIX globbing characters to create capture,
  groups, including the following:

  * `?` — matches a single character in a filepath. For example, you
  have files named `file000.txt`, `file001.txt`, `file002.txt`, and so on.
  You can set the glob pattern to `/file(?)(?)(?)` and the `join_on` key to
  `$2`, so that Pachyderm matches only the files that have same second
  character.

* `*` — any number of characters in the filepath. For example, if you set
  your capture group to `/(*)`, Pachyderm matches all files in the root
  directory.

  If you do not specify a correct `glob` pattern, Pachyderm performs the
  `cross` input operation instead of `join`.

* `input.pfs.lazy` — see the description in [PFS Input](#pfs-input).
* `input.pfs.empty_files` — see the description in [PFS Input](#pfs-input).

#### Git Input (alpha feature)

Git inputs allow you to pull code from a public git URL and execute that code as part of your pipeline. A pipeline with a Git Input will get triggered (i.e. will see a new input commit and will spawn a job) whenever you commit to your git repository.

**Note:** This only works on cloud deployments, not local clusters.

`input.git.URL` must be a URL of the form: `https://github.com/foo/bar.git`

`input.git.name` is the name for the input, its semantics are similar to
those of `input.pfs.name`. It is optional.

`input.git.branch` is the name of the git branch to use as input.

Git inputs also require some additional configuration. In order for new commits on your git repository to correspond to new commits on the Pachyderm Git Input repo, we need to setup a git webhook. At the moment, only GitHub is supported. (Though if you ask nicely, we can add support for GitLab or BitBucket).

1. Create your Pachyderm pipeline with the Git Input.

2. To get the URL of the webhook to your cluster, do `pachctl inspect pipeline` on your pipeline. You should see a `Githook URL` field with a URL set. Note - this will only work if you've deployed to a cloud provider (e.g. AWS, GKE). If you see `pending` as the value (and you've deployed on a cloud provider), it's possible that the service is still being provisioned. You can check `kubectl get svc` to make sure you see the `githook` service running.

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

For more information, see [Exporting Data by using egress](../../how-tos/export-data-out-pachyderm/#export-your-data-with-egress)

### Standby (optional)

`standby` indicates that the pipeline should be put into "standby" when there's
no data for it to process.  A pipeline in standby will have no pods running and
thus will consume no resources, it's state will be displayed as "standby".

Standby replaces `scale_down_threshold` from releases prior to 1.7.1.

### Cache Size (optional)

`cache_size` controls how much cache a pipeline's sidecar containers use. In
general, your pipeline's performance will increase with the cache size, but
only up to a certain point depending on your workload.

Every worker in every pipeline has a limited-functionality `pachd` server
running adjacent to it, which proxies PFS reads and writes (this prevents
thundering herds when jobs start and end, which is when all of a pipeline's
workers are reading from and writing to PFS simultaneously). Part of what these
"sidecar" pachd servers do is cache PFS reads. If a pipeline has a cross input,
and a worker is downloading the same datum from one branch of the input
repeatedly, then the cache can speed up processing significantly.

### Enable Stats (optional)

The `enable_stats` parameter turns on statistics tracking for the pipeline.
When you enable the statistics tracking, the pipeline automatically creates
and commits datum processing information to a special branch in its output
repo called `"stats"`. This branch stores information about each datum that
the pipeline processes, including timing information, size information, logs,
and `/pfs` snapshots. You can view this statistics by running the `pachctl
inspect datum` and `pachctl list datum` commands, as well as through the web UI.

Once turned on, statistics tracking cannot be disabled for the pipeline. You can
turn it off by deleting the pipeline, setting `enable_stats` to `false` or
completely removing it from your pipeline spec, and recreating the pipeline from
that updated spec file. While the pipeline that collects the stats
exists, the storage space used by the stats cannot be released.

!!! note
    Enabling stats results in slight storage use increase for logs and timing
    information.
    However, stats do not use as much extra storage as it might appear because
    snapshots of the `/pfs` directory that are the largest stored assets
    do not require extra space.

### Service (alpha feature, optional)

`service` specifies that the pipeline should be treated as a long running
service rather than a data transformation. This means that `transform.cmd` is
not expected to exit, if it does it will be restarted. Furthermore, the service
is exposed outside the container using a Kubernetes service.
`"internal_port"` should be a port that the user code binds to inside the
container, `"external_port"` is the port on which it is exposed through the
`NodePorts` functionality of Kubernetes services. After a service has been
created, you should be able to access it at
`http://<kubernetes-host>:<external_port>`.

### Spout (optional)

`spout` is a type of pipeline that processes streaming data.
Unlike a union or cross pipeline, a spout pipeline does not have
a PFS input. Instead, it opens a Linux *named pipe* into the source of the
streaming data. Your pipeline
can be either a spout or a service and not both. Therefore, if you added
the `service` as a top-level object in your pipeline, you cannot add `spout`.
However, you can expose a service from inside of a spout pipeline by
specifying it as a field in the `spout` spec. Then, Kubernetes creates
a service endpoint that you can expose externally. You can get the information
about the service by running `kubectl get services`.

For more information, see [Spouts](../concepts/pipeline-concepts/pipeline/spout.md).

### Max Queue Size (optional)
`max_queue_size` specifies that maximum number of datums that a worker should
hold in its processing queue at a given time (after processing its entire
queue, a worker "checkpoints" its progress by writing to persistent storage).
The default value is `1` which means workers will only hold onto the value that
they're currently processing.

Increasing this value can improve pipeline performance, as that allows workers
to simultaneously download, process and upload different datums at the same
time (and reduces the total time spent on checkpointing). Decreasing this value
can make jobs more robust to failed workers, as work gets checkpointed more
often, and a failing worker will not lose as much progress. Setting this value
too high can also cause problems if you have `lazy` inputs, as there's a cap of
10,000 `lazy` files per worker and multiple datums that are running all count
against this limit.

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
been set as a [JSON Merge Patch](https://tools.ietf.org/html/rfc7386). This
means that you can modify things such as the storage and user containers.

### Pod Patch (optional)
`pod_patch` is similar to `pod_spec` above but is applied as a [JSON
Patch](https://tools.ietf.org/html/rfc6902). Note, this means that the
process outlined above of modifying an existing pod spec and then manually
blanking unchanged fields won't work, you'll need to create a correctly
formatted patch by diffing the two pod specs.

## The Input Glob Pattern

Each PFS input needs to specify a [glob pattern](../concepts/pipeline-concepts/distributed_computing.md).

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
retries. Interaction with outside services should be [idempotent](https://en.wikipedia.org/wiki/Idempotence) to prevent
unexpected behavior. Furthermore, one of the running services that your code
can connect to is Pachyderm itself, this is generally not recommended as very
little of the Pachyderm API is idempotent, but in some specific cases it can be
a viable approach.
