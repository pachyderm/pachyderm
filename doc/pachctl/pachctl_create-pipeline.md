## ./pachctl create-pipeline

Create a new pipeline.

### Synopsis


Create a new pipeline from a spec

# Pipeline Specification

This document discusses each of the fields present in a pipeline specification. To see how to use a pipeline spec, refer to the [pachctl create-pipeline](./pachctl/pachctl_create-pipeline.html) doc.

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
    } ]
  },
  "parallelism": int,
  "inputs": [
    {
      "repo": {
        "name": string
      },
      "runEmpty": false,
      "method": "map"/"reduce"/"global"
      // alternatively, method can be specified as an object.
      // this is only for advanced use cases; most of the time, one of the three
      // strategies above should suffice.
      "method": {
        "partition": "block"/"file"/"repo",
        "incremental": bool
      }
    }
  ]
}
```

### Name

`pipeline.name` is the name of the pipeline that you are creating.  Each pipeline needs to have a unique name.

### Transform

`transform.image` is the name of the Docker image that your jobs run in.  Currently, this image needs to [inherit from a Pachyderm-provided image known as `job-shim`](https://github.com/pachyderm/pachyderm/blob/fae98e54af0d6932e258e4b0df4ea784414c921e/examples/fruit_stand/Dockerfile#L1).

`transform.cmd` is the command passed to the Docker run invocation.  Note that as with Docker, cmd is not run inside a shell which means that things like wildcard globbing (`*`), pipes (`|`) and file redirects (`>` and `>>`) will not work.  To get that behavior, you can set `cmd` to be a shell of your choice (e.g. `sh`) and pass a shell script to stdin.

`transform.stdin` is an array of lines that are sent to your command on stdin.  Lines need not end in newline characters.

`transform.env is a map from key to value of environment variables that will be injected into the container

`transform.secrets` is an array of secrets, secrets reference Kubernetes secrets by name and specify a path that the secrets should be mounted to. Secrets are useful for embedding sensitive data such as credentials. Read more about secrets in Kubernetes [here](http://kubernetes.io/docs/user-guide/secrets/).

### Parallelism

`parallelism` is how many copies of your container should run in parallel.  If you'd like Pachyderm to automatically scale the parallelism based on available cluster resources, you can set this to 0.

### Inputs

`inputs` specifies a set of Repos that will be visible to the jobs during runtime. Commits to these repos will automatically trigger the pipeline to create new jobs to process them.

`inputs.runEmpty` specifies what happens when an empty commit comes into the input repo.  If this flag is set to false (the default), then the empty commit won't trigger a job.  If set to true, the empty commit will trigger a job. 

`inputs.method` specifies two different properties:
- Partition unit: How input data  will be partitioned across parallel containers.
- Incrementality: Whether the entire all of the data or just the new data (diff) is processed. 
 
The next section explains input methods in detail.

### Pipeline Input Methods

For each pipeline input, you may specify a "method".  A method dictates exactly what happens in the pipeline when a commit comes into the input repo.

A method consists of two properties: partition unit and incrementality.

#### Partition Unit
Partition unit specifies the granularity at which input data is parallelized across containers.  It can be of three values: 

* `block`: different blocks of the same file may be parelleized across containers.
* `file`: the files and/or directories residing under the root directory (/) must be grouped together.  For instance, if you have four files in a directory structure like: 

```
/foo 
/bar
/buzz
   /a
   /b
```
then there are only three top-level objects, `/foo`, `/bar`, and `/buzz`, each of which will remain grouped in the same container. 
* `repo`: the entire repo.  In this case, the input won't be partitioned at all. 

#### Incrementality

Incrementality is a boolean flag that describes what data needs to be available when a new commit is made on an input repo. Namely, do you want to process only the new data in that commmit (the diff) or does all of the data need to be reprocessed?

For instance, if you have a repo with the file `/foo` in commit 1 and file `/bar` in commit 2, then:

* If the input is incremental, the first job sees file `/foo` and the second job sees file `/bar`.
* If the input is nonincremental, the first job sees file `/foo` and the second job sees file `/foo` and file `/bar`.

For convenience, we have defined aliases for the four most commonly used input methods: map, reduce, incremental-reduce, and global.  They are defined below:

|                | Block |  Top-level Objects |  Repo  |
|----------------|-------|--------------------|--------|
|   Incremental  |  map  |                    |        |
| Nonincremental |       |       reduce       | global |

#### Defaults
If no method is specified, the `map` method (Block + Incremental) is used by default.


### Multiple Inputs

A pipeline is allowed to have multiple inputs.  The important thing to understand is what happens when a new commit comes into one of the input repos.  In short, a pipeline processes the **cross product** of its inputs.  We will use an example to illustrate.

Consider a pipeline that has two input repos: `foo` and `bar`.  `foo` uses the `file/incremental` method and `bar` uses the `reduce` method.  Now let's say that the following events occur:

```
1. PUT /file-1 in commit1 in foo -- no jobs triggered
2. PUT /file-a in commit1 in bar -- triggers job1
3. PUT /file-2 in commit2 in foo -- triggers job2
4. PUT /file-b in commit2 in bar -- triggers job3
```

The first time the pipeline is triggered will be when the second event completes.  This is because we need data in both repos before we can run the pipeline.

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

`job1` sees `/pfs/foo/file-1` and `/pfs/bar/file-a` because those are the only files available.

`job2` sees `/pfs/foo/file-2` and `/pfs/bar/file-a` because it's triggered by commit2 in `foo`, and `foo` uses an incremental input method (`file/incremental`).

`job3` sees all the files because it's triggered by commit2 in `bar`, and `bar` uses a non-incremental input method (`reduce`).

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

This pipeline runs when the repo `my-input` gets a new commit.  The pipeline will spawn 4 parallel jobs, each of which runs the command `my-binary` in the Docker image `my-image`, with `arg1` and `arg2` as arguments to the command and `my-std-input` as the standard input.  Each job will get a set of blocks from the new commit as its input because `method` is set to `map`.

## PPS Mounts and File Access

### Mount Paths

The root mount point is at `/pfs`, which contains:

- `/pfs/input_repo_a` which is where you would find the latest commit from the `input_repo_a` input `Repo` you specified.
  - Each input repo will be found here by name
  - Note: Unlike when mounting locally for debugging, there is no `Commit` ID in the path. This is because the commit will always change, and the ID isn't relevant to the processing. The commit that is exposed is configured based on the `incrementality` flag above
- `/pfs/out` which is where you write any output
- `/pfs/prev` which is this `Job` or `Pipeline`'s previous output, if it exists. (You can think of it as this job's output commit's parent). 

### Output Formats

PFS supports considers data to be delimited by line, JSON, or binary blobs. [Refer here for more information on delimiters](./pachyderm_file_system.html#block-delimiters)

## Environment Variables

When the pipeline runs, the input and output commit IDs are exposed via environment variables:

- `$PACH_OUTPUT_COMMIT_ID` contains the output commit of the job itself
- For each of the job's input repositories, there will be a corresponding environment variable w the input commid ID:
  - e.g. if there are two input repos `foo` and `bar, the following will be populated:
    - `$PACH_FOO_COMMIT_ID`
    - `$PACH_BAR_COMMIT_ID`


## Flash-crowd behavior

In distributed systems, a flash-crowd behavior occurs when a large number of nodes send traffic to a particular node in an uncoordinated fashion, causing the node to become a hotspot, resulting in performance degradation.

To understand how such a behavior can occur in Pachyderm, it's important to understand the way requests are sharded in a Pachyderm cluster.  Pachyderm currently employs a simple sharding scheme that shards based on file names.  That is, requests pertaining to a certain file will be sent to a specific node.  As a result, if you have a number of nodes processing a large dataset in parallel, it's advantageous for them to process files in a random order.

For instance, imagine that you have a dataset that contains `file_A`, `file_B`, and `file_C`, each of of which is 1TB in size.  Now, each of your nodes will get a portion of each of these files.  If your nodes independently start processing files in alphanumeric order, they will all start with `file_A`, causing all traffic to be sent to the node that handles `file_A`.  In contrast, if your nodes process files in a random order, traffic will be distributed between three nodes.



```
./pachctl create-pipeline -f pipeline.json
```

### Options

```
  -f, --file string   The file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
```

### Options inherited from parent commands

```
      --log-flush-frequency duration   Maximum number of seconds between log flushes (default 5s)
  -v, --verbose                        Output verbose logs
```

### SEE ALSO
* [./pachctl](./pachctl.md)	 - 

###### Auto generated by spf13/cobra on 12-Sep-2016
