# Glob Pattern

Defining how your data is spread among workers is one of
the most important aspects of distributed computation and is
the fundamental idea around concepts such as Map and Reduce.

Instead of confining users to data-distribution patterns,
such as Map, that splits everything as much as possible, and
Reduce, that groups all the data, Pachyderm
uses glob patterns to provide incredible flexibility to
define data distribution.

You can configure a glob pattern for each PFS input in
the input field of a pipeline specification. Pachyderm detects
this parameter and divides the input data into
individual *datums*.

You can think of each input repository as a filesystem where
the glob pattern is applied to the root of the
filesystem. The files and directories that match the
glob pattern are considered datums. The Pachyderm's
concept of glob patterns is similar to the Unix glob patterns.
For example, the `ls *.md` command matches all files with the
`.md` file extension.

In Pachyderm, the `/` and `*` indicators are most
commonly used globs.

The following are examples of glob patterns that you can define:

* `/` — Pachyderm denotes the whole repository as a
  single datum and sends all of the input data to a
  single worker node to be processed together.
* `/*` — Pachyderm defines each top-level filesystem
  object, that is a file or a directory, in the input
  repo as a separate datum. For example,
  if you have a repository with ten files in it and no
  directory structure, Pachyderm identifies each file as a
  single datum and processes them independently.
* `/*/*` — Pachyderm processes each filesystem object
  in each subdirectory as a separate datum.

<!-- Add the ohmyglob examples here-->

If you have more than one input repo in your pipeline,
you can define a different glob pattern for each input
repo. You can combine the datums from each input repo
by using the `cross`, `union`, `join`, or `group` operator to
create the final datums that your code processes.
For more information, see [Cross and Union](./cross-union.md), [Join](./join.md), [Group](./group.md).

## Example of Defining Datums

For example, you have the following directory:

!!! example
    ```shell
    /California
       /San-Francisco.json
       /Los-Angeles.json
       ...
    /Colorado
       /Denver.json
       /Boulder.json
       ...
    ...
    ```

Each top-level directory represents a US
state with a `json` file for each city in that state.

If you set glob pattern to `/`, every time
you change anything in any of the
files and directories or add a new file to the
repository, Pachyderm processes the contents
of the whole repository from scratch as a single datum.
For example, if you add `Sacramento.json` to the
`California/` directory, Pachyderm processes all files
and folders in the repo as a single datum.

If you set `/*` as a glob pattern, Pachyderm processes
the data for each state individually. It
defines one datum per state, which means that all the cities for
a given state are processed together by a single worker, but each
state is processed independently. For example, if you add a new file
`Sacramento.json` to the `California/` directory, Pachyderm
processes the `California/` datum only.

If you set `/*/*`, Pachyderm processes each city as a single
datum on a separate worker. For example, if you add
the `Sacramento.json` file, Pachyderm processes the
`Sacramento.json` file only.

Glob patterns also let you take only a particular directory or subset of
directories as an input instead of the whole repo. For example,
you can set `/California/*` to process only the data for the state of
California. Therefore, if you add a new city in the `Colorado/` directory,
Pachyderm ignore this change and does not start the pipeline.
However, if you add  `Sacramento.json` to the `California/` directory,
Pachyderm  processes the `California/` datum.

## Test a Glob pattern

You can use the `pachctl glob file` command to preview which filesystem
objects a pipeline defines as datums. This command helps
you to test various glob patterns before you use them in a pipeline.

* If you set the `glob` property to `/`, Pachyderm detects all
top-level filesystem objects in the `train` repository as one
datum:

!!! example
    ```shell
    pachctl glob file train@master:/
    ```

    **System Response:**

    ```shell
    NAME TYPE SIZE
    /    dir  15.11KiB
    ```

* If you set the `glob` property to `/*`, Pachyderm detects each
top-level filesystem object in the `train` repository as a separate
datum:

!!! example
    ```shell
    pachctl glob file train@master:/*
    ```

    **System Response:**

    ```shell
    NAME                   TYPE SIZE
    /IssueSummarization.py file 1.224KiB
    /requirements.txt      file 74B
    /seq2seq_utils.py      file 13.81KiB
    ```

## Test your Datums

The granularity of your datums defines how your data will be distributed across the available workers allocated to a job.
Pachyderm allows you to check those datums:

  - for a pipeline currently being developed  
  - for a past job 

### Testing your glob pattern before creating a pipeline
You can use the `pachctl list datum -f <my_pipeline_spec.json>` command to preview the datums defined by a pipeline given its specification file. 

!!! note "Note"  
    The pipeline does not need to have been created for the command to return the list of datums. This "dry run" helps you adjust your glob pattern when creating your pipeline.
 

!!! example
    ```shell
    pachctl list datum -f edges.json
    ```
    **System Response:**

    ```shell
        ID FILES                                                STATUS TIME
    -  images@8c958d1523f3428a98ac97fbfc367bae:/g2QnNqa.jpg -      -
    -  images@8c958d1523f3428a98ac97fbfc367bae:/8MN9Kg0.jpg -      -
    -  images@8c958d1523f3428a98ac97fbfc367bae:/46Q8nDz.jpg -      -
    ```

### Running list datum on a past job 
You can use the `pachctl list datum <job_number>` command to check the datums processed by a given job.

!!! example
    ```shell
    pachctl list datum d10979d9f9894610bb287fa5e0d734b5
    ```
    **System Response:**

    ```shell
        ID                                                                   FILES                                                STATUS TIME
    ebd35bb33c5f772f02d7dfc4735ad1dde8cc923474a1ee28a19b16b2990d29592e30 images@8c958d1523f3428a98ac97fbfc367bae:/g2QnNqa.jpg -      -
    ebd3ce3cdbab9b78cc58f40aa2019a5a6bce82d1f70441bd5d41a625b7769cce9bc4 images@8c958d1523f3428a98ac97fbfc367bae:/8MN9Kg0.jpg -      -
    ebd32cf84c73cfcc4237ac4afdfe6f27beee3cb039d38613421149122e1f9faff349 images@8c958d1523f3428a98ac97fbfc367bae:/46Q8nDz.jpg -      -
    ```

!!! note "Note"  
    Now that the 3 datums have been processed, their ID field is showing.
