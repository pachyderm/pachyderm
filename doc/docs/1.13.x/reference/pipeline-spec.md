# Pipeline Specification

This document discusses each of the fields present in a pipeline specification.
To see how to use a pipeline spec to create a pipeline, refer to the [pachctl
create pipeline](./pachctl/pachctl_create_pipeline.md) section.

!!! Info
    - Pachyderm's pipeline specifications can be written in JSON or YAML.
    - Pachyderm uses its json parser if the first character is `{`.
## Manifest Format

=== "JSON Full Specifications"
    ```json
    {
      "pipeline": {
        "name": string
      },
      "description": string,
      "metadata": {
        "annotations": {
            "annotation": string
        },
        "labels": {
            "label": string
        }
      },
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
      "sidecar_resource_limits": {
        "memory": string,
        "cpu": number
      },
      "datum_timeout": string,
      "datum_tries": int,
      "job_timeout": string,
      "input": {
        <"pfs", "cross", "union", "join", "group", "cron", or "git" see below>
      },
      "s3_out": bool,
      "reprocess_spec": string,
      "output_branch": string,
      "egress": {
        "URL": "s3://bucket/dir"
      },
      "standby": bool,
      "autoscaling": bool,
      "cache_size": string,
      "enable_stats": bool,
      "service": {
        "internal_port": int,
        "external_port": int
      },
      "spout": {
      \\ Optionally, you can combine a spout with a service:
      "service": {
            "internal_port": int,
            "external_port": int
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
      "empty_files": bool,
      "s3": bool
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
          "s3": bool
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
          "s3": bool
        }
      }
      ...
    ]


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
          "join_on": string,
          "outer_join": bool,
          "lazy": bool,
          "empty_files": bool,
          "s3": bool
        }
      },
      {
        "pfs": {
          "name": string,
          "repo": string,
          "branch": string,
          "glob": string,
          "join_on": string,
          "outer_join": bool,
          "lazy": bool,
          "empty_files": bool,
          "s3": bool
        }
      }
    ]


    ------------------------------------
    "group" input
    ------------------------------------

    "group": [
      {
        "pfs": {
          "name": string,
          "repo": string,
          "branch": string,
          "glob": string,
          "group_by": string,
          "lazy": bool,
          "empty_files": bool,
          "s3": bool
        }
      },
      {
        "pfs": {
          "name": string,
          "repo": string,
          "branch": string,
          "glob": string,
          "group_by": string,
          "lazy": bool,
          "empty_files": bool,
          "s3": bool
        }
      }
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
    "git" input
    ------------------------------------

    "git": {
      "URL": string,
      "name": string,
      "branch": string
    }

    ```

=== "YAML Sample"
    ```yaml
    pipeline:
      name: edges
    description: A pipeline that performs image edge detection by using the OpenCV library.
    input:
      pfs:
        glob: /*
        repo: images
    transform:
      cmd:
        - python3
        - /edges.py
      image: pachyderm/opencv
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

### Metadata

This parameter enables you to add metadata to your pipeline pods by using Kubernetes' `labels` and `annotations`. Labels help you to organize and keep track of your cluster objects by creating groups of pods based on the application they run, resources they use, or other parameters. Labels simplify the querying of Kubernetes objects and are handy in operations.

Similarly to labels, you can add metadata through annotations. The difference is that you can specify any arbitrary metadata through annotations.

Both parameters require a key-value pair.  Do not confuse this parameter with `pod_patch` which adds metadata to the user container of the pipeline pod. For more information, see [Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) and [Kubernetes Annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) in the Kubernetes documentation.

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
Pachyderm injects into the container. There are also environment variables
that are automatically injected into the container, such as:

* `PACH_JOB_ID` – the ID of the current job.
* `PACH_OUTPUT_COMMIT_ID` – the ID of the commit in the output repo for 
the current job.
* `<input>_COMMIT` - the ID of the input commit. For example, if your
input is the `images` repo, this will be `images_COMMIT`.

For a complete list of variables and
descriptions see: [Configure Environment Variables](../../deploy-manage/deploy/environment-variables/).

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
kubectl create secret docker-registry myregistrykey --docker-server=DOCKER_REGISTRY_SERVER --docker-username=DOCKER_USER --docker-password=DOCKER_PASSWORD --docker-email=DOCKER_EMAIL
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

The default value is "constant=1".

Because spouts and services are designed to be single instances, do not
modify the default `parallism_spec` value for these pipelines.

### Resource Requests (optional)

`resource_requests` describes the amount of resources that the pipeline
workers will consume. Knowing this in advance
enables Pachyderm to schedule big jobs on separate machines, so that they
do not conflict, slow down, or terminate.

This parameter is optional, and if you do not explicitly add it in
the pipeline spec, Pachyderm creates Kubernetes containers with the
following default resources: 

- The user container requests 0 CPU, 0 disk space, and 64MB of memory. 
- The init container requests the same amount of CPU, memory, and disk
space that is set for the user container.
- The storage container requests 0 CPU and the amount of memory set by the
[cache_size](#cache-size-optional) parameter.

The `resource_requests` parameter enables you to overwrite these default
values.

The `memory` field is a string that describes the amount of memory, in bytes,
that each worker needs. Allowed SI suffixes include M, K, G, Mi, Ki, Gi, and
other.

For example, a worker that needs to read a 1GB file into memory might set
`"memory": "1.2G"` with a little extra for the code to use in addition to the
file. Workers for this pipeline will be placed on machines with at least
1.2GB of free memory, and other large workers will be prevented from using it,
if they also set their `resource_requests`.

The `cpu` field is a number that describes the amount of CPU time in `cpu
seconds/real seconds` that each worker needs. Setting `"cpu": 0.5` indicates that
the worker should get 500ms of CPU time per second. Setting `"cpu": 2`
indicates that the worker gets 2000ms of CPU time per second. In other words,
it is using 2 CPUs, though worker threads might spend 500ms on four
physical CPUs instead of one second on two physical CPUs.

The `disk` field is a string that describes the amount of ephemeral disk space,
in bytes, that each worker needs. Allowed SI suffixes include M, K, G, Mi,
Ki, Gi, and other.

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

For more information about resource requests and limits see the
[Kubernetes docs](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
on the subject.

### Resource Limits (optional)

`resource_limits` describes the upper threshold of allowed resources a given
worker can consume. If a worker exceeds this value, it will be evicted.

The `gpu` field is a number that describes how many GPUs each worker needs.
Only whole number are supported, Kubernetes does not allow multiplexing of
GPUs. Unlike the other resource fields, GPUs only have meaning in Limits, by
requesting a GPU the worker will have sole access to that GPU while it is
running. It's recommended to enable `autoscaling` if you are using GPUs so other
processes in the cluster will have access to the GPUs while the pipeline has
nothing to process. For more information about scheduling GPUs see the
[Kubernetes docs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/)
on the subject.

### Sidecar Resource Limits (optional)

`sidecar_resource_limits` determines the upper threshold of resources
allocated to the sidecar containers.

This field can be useful in deployments where Kubernetes automatically
applies resource limits to containers, which might conflict with Pachyderm
pipelines' resource requests. Such a deployment might fail if Pachyderm
requests more than the default Kubernetes limit. The `sidecar_resource_limits`
enables you to explicitly specify these resources to fix the issue.

### Datum Timeout (optional)

`datum_timeout` determines the maximum execution time allowed for each
datum. The value must be a string that represents a time value, such as
`1s`, `5m`, or `15h`. This parameter takes precedence over the parallelism
or number of datums, therefore, no single datum is allowed to exceed
this value. By default, `datum_timeout` is not set, and the datum continues to
be processed as long as needed.

### Datum Tries (optional)

`datum_tries` is an integer, such as `1`, `2`, or `3`, that determines the
number of times a job attempts to run on a datum when a failure occurs. 
Setting `datum_tries` to `1` will attempt a job once with no retries. 
Only failed datums are retried in a retry attempt. If the operation succeeds
in retry attempts, then the job is marked as successful. Otherwise, the job
is marked as failed.


### Job Timeout (optional)

`job_timeout` determines the maximum execution time allowed for a job. It
differs from `datum_timeout` in that the limit is applied across all
workers and all datums. This is the *wall time*, which means that if
you set `job_timeout` to one hour and the job does not finish the work
in one hour, it will be interrupted.
When you set this value, you need to
consider the parallelism, total number of datums, and execution time per
datum. The value must be a string that represents a time value, such as
`1s`, `5m`, or `15h`. In addition, the number of datums might change over
jobs. Some new commits might have more files, and therefore, more datums.
Similarly, other commits might have fewer files and datums. If this
parameter is not set, the job will run indefinitely until it succeeds or fails.

### S3 Output Repository

`s3_out` allows your pipeline code to write results out to an S3 gateway
endpoint instead of the typical `pfs/out` directory. When this parameter
is set to `true`, Pachyderm includes a sidecar S3 gateway instance
container in the same pod as the pipeline container. The address of the
output repository will be `s3://<output_repo>`. If you enable `s3_out`,
verify that the `enable_stats` parameter is disabled.

If you want to expose an input repository through an S3 gateway, see
`input.pfs.s3` in [PFS Input](#pfs-input). 

!!! note "See Also:"
    [Environment Variables](../../deploy-manage/deploy/environment-variables/)

### Input

`input` specifies repos that will be visible to the jobs during runtime.
Commits to these repos will automatically trigger the pipeline to create new
jobs to process them. Input is a recursive type, there are multiple different
kinds of inputs which can be combined together. The `input` object is a
container for the different input types with a field for each, only one of
these fields be set for any instantiation of the object. While most types
of pipeline specifications require an `input` repository, there are
exceptions, such as a spout, which does not need an `input`.

```json
{
    "pfs": pfs_input,
    "union": union_input,
    "cross": cross_input,
    "join": join_input,
    "group": group_input,
    "cron": cron_input,
    "git": git_input,
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
    "empty_files": bool,
    "s3": bool,
    "trigger": {
        "branch": string,
        "all": bool,
        "cron_spec": string,
        "size": string,
        "commits": int
    }
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

!!! note
    `lazy` does not support datums that
    contain more than 10000 files.

`input.pfs.empty_files` controls how files are exposed to jobs. If
set to `true`, it causes files from this PFS to be presented as empty files.
This is useful in shuffle pipelines where you want to read the names of
files and reorganize them by using symlinks.

`input.pfs.s3` sets whether the sidecar in the pipeline worker pod
should include a sidecar S3 gateway instance. This option enables an S3 gateway
to serve on a pipeline-level basis and, therefore, ensure provenance tracking
for pipelines that integrate with external systems, such as Kubeflow. When
this option is set to `true`, Pachyderm deploys an S3 gateway instance
alongside the pipeline container and creates an S3 bucket for the pipeline
input repo. The address of the
input repository will be `s3://<input_repo>`. When you enable this
parameter, you cannot use glob patterns. All files will be processed
as one datum.

Another limitation for S3-enabled pipelines is that you can only use
either a single input or a cross input. Join and union inputs are not
supported.

If you want to expose an output repository through an S3
gateway, see [S3 Output Repository](#s3-output-repository).

`input.pfs.trigger`
Specifies a trigger that must be met for the pipeline to trigger on this input.
To learn more about triggers read the
[deferred process docs](../concepts/advanced-concepts/deferred-processing.md).

#### Union Input

Union inputs take the union of other inputs. In the example
below, each input includes individual datums, such as if  `foo` and `bar`
were in the same repository with the glob pattern set to `/*`.
Alternatively, each of these datums might have come from separate repositories
with the glob pattern set to `/` and being the only file system objects in these
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
with the glob pattern set to `/` and being the only file system objects in these
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
when `"overwrite"` is disabled, ticks accumulate in the cron input repo. When
`"overwrite"` is enabled, Pachyderm erases the old ticks and adds new ticks
with each commit. If you do not add any manual ticks or run
`pachctl run cron`, only one tick file per commit (for the latest tick)
is added to the input repo.

#### Join Input

A join input enables you to join files that are stored in separate
Pachyderm repositories and that match a configured glob
pattern. A join input must have the `glob` and `join_on` parameters configured
to work properly. A join can combine multiple PFS inputs.

You can optionally add `"outer_join": true` to your PFS input. 
In that case, you will alter the join's behavior from a default "inner-join" (creates a datum if there is a match only) to a "outer-join" (the repos marked as `"outer_join": true` will see a datum even if there is no match).
You can set 0 to many PFS input to `"outer_join": true` within your `join`.

You can specify the following parameters for the `join` input.

* `input.pfs.name` — the name of the PFS input that appears in the
`INPUT` field when you run the `pachctl list job` command.
If an input name is not specified, it defaults to the name of the repo.

* `input.pfs.repo` — see the description in [PFS Input](#pfs-input).
the name of the Pachyderm repository with the data that
you want to join with other data.

* `input.pfs.branch` — see the description in [PFS Input](#pfs-input).

* `input.pfs.glob` — a wildcard pattern that defines how a dataset is **broken
  up into datums** for further processing. When you use a glob pattern in joins,
  it creates a naming convention that Pachyderm uses to join files. In other
  words, Pachyderm joins the files that are named according to the glob
  pattern and skips those that are not.

    You can specify the glob pattern for joins in a parenthesis to create
    one or multiple capture groups. A capture group can include one or multiple
    characters. Use standard UNIX globbing characters to create capture,
    groups, including the following:

    - `?` — matches a single character in a filepath. For example, you
    have files named `file000.txt`, `file001.txt`, `file002.txt`, and so on.
    You can set the glob pattern to `/file(?)(?)(?)` and the `join_on` key to
    `$2`, so that Pachyderm matches only the files that have same second
    character.

    - `*` — any number of characters in the filepath. For example, if you set
    your capture group to `/(*)`, Pachyderm matches all files in the root
    directory.

    If you do not specify a correct `glob` pattern, Pachyderm performs the
    `cross` input operation instead of `join`.

* `input.pfs.outer_join`- Set to `true`, your PFS input will see datums even if there is no match. 
  Defaults to false.

* `input.pfs.lazy` — see the description in [PFS Input](#pfs-input).
* `input.pfs.empty_files` — see the description in [PFS Input](#pfs-input).

#### Group Input

A group input lets you group files that are stored in one or multiple
Pachyderm repositories by a configured glob
pattern. A group input must have the `glob` and `group_by` parameters configured
to work properly. A group can combine multiple inputs, as long as all the base inputs are PFS inputs.

You can specify the following parameters for the `group` input.

* `input.pfs.name` — the name of the PFS input that appears in the
`INPUT` field when you run the `pachctl list job` command.
If an input name is not specified, it defaults to the name of the repo.

* `input.pfs.repo` — see the description in [PFS Input](#pfs-input).
the name of the Pachyderm repository with the data that
you want to join with other data.

* `input.pfs.branch` — see the description in [PFS Input](#pfs-input).

* `input.pfs.glob` — a wildcard pattern that defines how a dataset is broken
  up into datums for further processing. When you use a glob pattern in a group input,
  it creates a naming convention that Pachyderm uses to group the files.
  

    You need specify in the glob pattern parenthesis to create
    one or multiple capture groups. A capture group can include one or multiple
    characters. Use standard UNIX globbing characters to create capture,
    groups, including the following:

    * `?` — matches a single character in a filepath. For example, you
    have files named `file000.txt`, `file001.txt`, `file002.txt`, and so on.
   You can set the glob pattern to `/file(?)(?)(?)` and the `group_by` key to
    `$2`, so that Pachyderm groups the files by just their second
    characters.

    * `*` — any number of characters in the filepath. For example, if you set
    your capture group to `/(*)`, Pachyderm will group the files by their filenames.

    If you do not specify a correct `glob` pattern, Pachyderm will place all of the files in a single group.

  Joins and groups can be used simultaneously; generally, you would use the join inside the group.
  The join will then essentially filter the files, and then the group will group by the remaining files (but will lose the join structure).

* `input.pfs.lazy` — see the description in [PFS Input](#pfs-input).
* `input.pfs.empty_files` — see the description in [PFS Input](#pfs-input).

#### Git Input

!!! Warning
    Git Inputs are an [experimental feature](../../reference/supported-releases/#experimental).

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

For more information, see [Exporting Data by using egress](../how-tos/basic-data-operations/export-data-out-pachyderm/export-data-egress.md)

### Standby (optional)

`standby` indicates that the pipeline should be put into "standby" when there's
no data for it to process.  A pipeline in standby will have no pods running and
thus will consume no resources, it's state will be displayed as "standby".

### Autoscaling (optional)
`autoscaling` indicates that the pipeline should automatically scale the worker
pool based on the datums it has to process. A pipeline with no outstanding jobs
will go into `standby` and have no workers.

`autoscaling` is a more sophisticated version of `standby` that was introduced
in 1.13.0.

### Cache Size (optional)

`cache_size` controls how much **in-memory cache** a pipeline's sidecar containers use. In
general, your pipeline's performance will increase with the cache size, but
only up to a certain point depending on your workload.

Every worker in every pipeline has a limited-functionality `pachd` server
running adjacent to it, which proxies PFS reads and writes (this prevents
thundering herds when jobs start and end, which is when all of a pipeline's
workers are reading from and writing to PFS simultaneously). 
Part of what these "sidecar" pachd servers do is *cache PFS reads:*
Any time the worker reads files from the input repo or writes files to the output repo,
the sidecar manages those requests, 
which will **cache the files in memory for better performance**.

If a pipeline has a cross input,
and a worker is downloading the same datum from one branch of the input
repeatedly, then the cache can speed up processing significantly.

If not explicitly specified, cache_size defaults to 64M.

!!! Note
    When setting `cache_size`, it is important to keep in mind that any file
    which is larger than 25% of the total `cache_size` will NOT be cached. 
    For example, if `cache_size` is set to 1G, 
    then only files which are 250M or smaller will be cached; 
    files larger than 250M will not be cached.

### Enable Stats (optional)

The `enable_stats` parameter turns on statistics tracking for the pipeline.
When you enable the statistics tracking, the pipeline automatically creates
and commits datum processing information to a special branch in its output
repo called `"stats"`. This branch stores information about each datum that
the pipeline processes, including timing information, size information, logs,
and `/pfs` snapshots. You can view this statistics by running the `pachctl
inspect datum` and `pachctl list datum` commands, as well as through the web UI.
Do not enable statistics tracking for S3-enabled pipelines.

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

### Reprocess Datums (optional)

Per default, Pachyderm avoids repeated processing of unchanged datums (i.e., it processes only the datums that have changed and skip the unchanged datums). This [**incremental behavior**](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/#example-1-one-file-in-the-input-datum-one-file-in-the-output-datum) ensures efficient resource utilization. However, you might need to alter this behavior for specific use cases and **force the reprocessing of all of your datums systematically**. This is especially useful when your pipeline makes an external call to other resources, such as a deployment or triggering an external pipeline system.  Set `"reprocess_spec": "every_job"` in order to enable this behavior. 

!!! Note "About the default behavior "
    `"reprocess_spec": "until_success"` is the default behavior.
    To mitigate datums failing for transient connection reasons,
    Pachyderm automatically [retries user code three (3) times before marking a datum as failed](https://docs.pachyderm.com/latest/troubleshooting/pipeline-troubleshooting/#introduction). Additionally, you can [set the  `datum_tries`](https://docs.pachyderm.com/latest/reference/pipeline-spec/#datum-tries-optional) field to determine the number of times a job attempts to run on a datum when a failure occurs.


Let's compare `"until_success"` and `"every_job"`:

  Say we have 2 identical pipelines (`reprocess_until_success.json` and `reprocess_at_every_job.json`) but for the `"reprocess_spec"` field set to `"every_job"` in reprocess_at_every_job.json. 
  Both use the same input repo and have a glob pattern set to `/*`. 

  - When adding 3 text files to the input repo (file1.txt, file2.txt, file3.txt), the 2 pipelines (reprocess_until_success and reprocess_at_every_job) will process the 3 datums (here, the glob pattern `/*` creates one datum per file).
  - Now, let's add a 4th file file4.txt to our input repo or modify the content of file2.txt for example.
      - **Case of our default `reprocess_until_success.json pipeline`**: A quick check at the [list datum on the job id](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/glob-pattern/#running-list-datum-on-a-past-job) shows 4 datums, of which 3 were skipped. (Only the changed file was processed)
      - **Case of `reprocess_at_every_job.json`**: A quick check at the list datum on the job id shows that all 4 datums were reprocessed, none were skipped.

!!! Warning
    `"reprocess_spec": "every_job` will not take advantage of Pachyderm's default de-duplication. In effect, this can lead to slower pipeline performance. Before using this setting, consider other options such as including metadata in your file, naming your files with a timestamp, UUID, or other unique identifiers in order to take advantage of de-duplication. Review how [datum processing](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/relationship-between-datums/) works to understand more.

### Service (optional)

!!! Warning
    Service Pipelines are an [experimental feature](../../reference/supported-releases/#experimental).

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

`spout` is a type of pipeline
that ingests streaming data.
Unlike a union or cross pipeline,
a spout pipeline does not have
a PFS input.
Instead, it consumes data from an outside source.

!!! Note
    A service pipeline cannot be configured as a spout,
    but **a spout can have a service added to it**
    by adding the `service` attribute to the `spout` field.
    In that case, Kubernetes creates
    a service endpoint that you can expose externally. 
    You can get the information
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
 A chunk is the unit of work that workers claim. Each worker claims 1 or more datums 
 and it commits a full chunk once it's done processing it.
 
`chunk_spec.number` if nonzero, specifies that each chunk should contain `number`
 datums. Chunks may contain fewer if the total number of datums don't
 divide evenly. If you lower the chunk number to 1 it'll update after every datum, 
 the cost is extra load on etcd which can slow other stuff down.
 The default value is 2.

`chunk_spec.size_bytes` , if nonzero, specifies a target size for each chunk of datums.
 Chunks may be larger or smaller than `size_bytes`, but will usually be
 pretty close to `size_bytes` in size.

### Scheduling Spec (optional)
`scheduling_spec` specifies how the pods for a pipeline should be scheduled.

`scheduling_spec.node_selector` allows you to select which nodes your pipeline
will run on. Refer to the [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector)
on node selectors for more information about how this works.

`scheduling_spec.priority_class_name` allows you to select the prioriy class
for the pipeline, which will how Kubernetes chooses to schedule and deschedule
the pipeline. Refer to the [Kubernetes docs](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass)
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

Each PFS input needs to **specify a [glob pattern](../../concepts/pipeline-concepts/datum/glob-pattern/)**.

Pachyderm uses the glob pattern to determine how many "datums" an input
consists of.  [Datums](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/#datum) are the *unit of parallelism* in Pachyderm.  
Per default,
Pachyderm auto-scales its workers to process datums in parallel. 
You can override this behaviour by setting your own parameter
(see [Distributed Computing](https://docs.pachyderm.com/latest/concepts/advanced-concepts/distributed-computing/)).


## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_name` which is where you would find the datum.
  - Each input will be found here by its name, which defaults to the repo
  name if not specified.
- `/pfs/out` which is where you write any output.
