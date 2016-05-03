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
      "reduce": bool
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

`inputs.reduce` specifies how a repo will be partitioned among parallel containers.  If set to true, the data will be partitioned by files.  If set to false, the data will be partitioned by blocks.

## Examples

Here is a modified example from the [fruit stand example](../examples/fruit_stand/GUIDE.md).

```json
{
  "pipeline": {
    "name": "filter"
  },
  "transform": {
    "image": "fruit_stand",
    "cmd": [ "sh" ],
    "stdin": [
        "grep apple  /pfs/data/* >/pfs/out/apple",
        "grep banana /pfs/data/* >/pfs/out/banana",
        "grep orange /pfs/data/* >/pfs/out/orange"
    ]
  },
  "parallelism": "4",
  "inputs": [
    {
      "repo": {
        "name": "data"
      },
      "reduce": true
    }
  ]
}
```

This pipeline runs when the repo `data` gets a new commit.  The pipeline will spawn 4 parallel jobs, each of which runs a bash script.  Each job will get a unique set of files as input because `reduce` is set to true.

