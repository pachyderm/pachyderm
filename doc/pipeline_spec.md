# Pipeline Specification

## Format

```json
{
  "pipeline": {
    "name": string
  },
  "transform": {
    "image": string,
    "cmd": [ string ],
    "stdin": [ string ]
  },
  "parallelism": int,
  "inputs": [
    {
      "repo": {
        "name": string
      },
      "method": "map"/"reduce"/"incremental_reduce"/"global"
      // alternatively, method can be specified as an object.
      // this is only for advanced use cases; most of the time, one of the four
      // strategies above should suffice.
      "method": {
        "partition": "block"/"file"/"repo",
        "incrementality": bool
      }
    }
  ]
}
```

`pipeline.name` is the name of the pipeline that you are creating.  Each pipeline needs to have a unique name.

`transform.image` is the name of the Docker image that your jobs run in.  Currently, this image needs to [inherit from a Pachyderm-provided image known as `job-shim`](https://github.com/pachyderm/pachyderm/blob/fae98e54af0d6932e258e4b0df4ea784414c921e/examples/fruit_stand/Dockerfile#L1).

`transform.cmd` is the command passed to the Docker run invocation.  Note that as with Docker, cmd is not run inside a shell which means that things like wildcard globbing (`*`), pipes (`|`) and file redirects (`>` and `>>`) will not work.  To get that behavior, you can set `cmd` to be a shell of your choice (e.g. `sh`) and pass a shell script to stdin.

`transform.stdin` is an array of lines that are sent to your command on stdin.  Lines need not end in newline characters.

`parallelism` is how many copies of your container should run in parallel.  If you'd like Pachyderm to automatically scale the parallelism based on available cluster resources, you can set this to 0.

`inputs` specifies a set of Repos that will be visible to the jobs during runtime. Commits to these repos will automatically trigger the pipeline to create new jobs to process them.

`inputs.method` specifies how a repo will be partitioned among parallel containers, and whether the entire repo or just the new commit is used as the input.

You may specify a method using either an alias or a JSON object.  We support four aliases that represent the four commonly used strategies:

* map: each job sees a part of the new commit; files may be partitioned.
* reduce: each job sees a part of the entire repo; files are not partitioned
* incremental_reduce: each job sees a part of the new commit; files are not partitioned.
* global: each job sees the entire repo

If a method is not specified, the "map" method is used by default.


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

This pipeline runs when the repo `my-input` gets a new commit.  The pipeline will spawn 4 parallel jobs, each of which runs the command `my-binary` in the Docker image `my-image`, with `arg1` and `arg2` as arguments to the command and `my-std-input` as the standard input.  Each job will get a part of the new commit as input because `method` is set to `map`.

## Accessing the output of a job's parent

Sometimes in a job, you might want to use the output of the job's parent.  See the "sum" part of the [fruit stand demo](../examples/fruit_stand/README.md) as an example.  If the job does have a parent, the output of its parent will be available under `/pfs/prev`. 
