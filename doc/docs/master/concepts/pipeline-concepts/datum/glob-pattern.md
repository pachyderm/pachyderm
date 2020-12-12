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
For more information, see [Cross and Union](./cross-union.md), [Join](./join.md),[Group](./group.md).

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

Pachyderm allows you to check the list and content of datums defined by a pipeline for a past job or test your datums for a pipeline currently being developed. 

### Testing your glob pattern before creating a pipeline
You can use the `pachctl list datum -f <my_pipeline_spec.json>` command to preview the datums defined by a pipeline given its specification file. 

!!! note The pipeline does not need to have been created for the command to return the list of datums. This dry run helps you adjust your glob pattern when creating your pipeline.

!!! example
    ```shell
    pachctl list datum -f lab_group_by_hospital.json
    ```
    **System Response:**

    ```shell
        ID FILES                                                                                                                                                                                                                                                                                                                                                                  STATUS TIME
    -  labresults@b2095152ffc440c682aeb55045e30d7e:/T1606331395-LIPID-PATID2-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707557-LIPID-PATID1-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707597-LIPID-PATID4-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707579-LIPID-PATID3-CLIA24D9871327.txt -      -
    -  labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707613-LIPID-PATID1-CLIA24D9871328.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707635-LIPID-PATID3-CLIA24D9871328.txt
    ```

### Running list datum on a past job 
You can use the `pachctl list datum <job_number>` command to check the datums processed by a given job.

!!! example
    ```shell
    pachctl list datum 6186ee85db014cfcaf14ef2cb8395d6d
    ```
    **System Response:**

    ```shell
    ID                                                                   FILES                                                                                                                                                                                                                                                                                                                                                                  STATUS TIME
    42716c4ac21ab9dc093958f1decdb01890a373efcff2e33964b1d05502b0bcc87bd3 labresults@b2095152ffc440c682aeb55045e30d7e:/T1606331395-LIPID-PATID2-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707557-LIPID-PATID1-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707597-LIPID-PATID4-CLIA24D9871327.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707579-LIPID-PATID3-CLIA24D9871327.txt -      -
    42718eb32c45e1136076a19b271a1cf2bbc09602c779496f1db4e2b8625dc2d8200a labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707613-LIPID-PATID1-CLIA24D9871328.txt, labresults@b2095152ffc440c682aeb55045e30d7e:/T1606707635-LIPID-PATID3-CLIA24D9871328.txt                                                                                                                                                                                     -      -
    ```

