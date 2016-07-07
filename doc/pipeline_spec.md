# Pipeline Specification

## Format

```
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
      "method": "map"/"reduce"/"global"
      // alternatively, method can be specified as an object.
      // this is only for advanced use cases; most of the time, one of the four
      // strategies above should suffice.
      "method": {
        "partition": "block"/"file"/"repo",
        "incremental": bool
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

## Accessing the output of a job's parent

Sometimes in a job, you might want to use the output of the job's parent.  See the "sum" part of the [fruit stand demo](../examples/fruit_stand/README.md) as an example.  If the job does have a parent, the output of its parent will be available under `/pfs/prev`. 

## Environment Variables

When the pipeline runs, the input and output commit IDs are exposed via environment variables:

- `$PACH_OUTPUT_COMMIT_ID` contains the output commit of the job itself
- For each of the job's input repositories, there will be a corresponding environment variable w the input commid ID:
  - e.g. if there are two input repos `foo` and `bar, the following will be populated:
    - `$PACH_FOO_COMMIT_ID`
    - `$PACH_BAR_COMMIT_ID`

