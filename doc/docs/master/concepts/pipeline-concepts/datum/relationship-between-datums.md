# Datum Processing

A [datum](../index.md) is a Pachyderm abstraction that helps in optimizing
pipeline processing. Datums exist only as a pipeline
processing property and are not filesystem objects. You can never
copy a datum. They are a representation of a unit
of work.
## Job Processing and Datums
When new data comes in (in the form of commit(s) in its input repo(s)), a Pachyderm pipeline automatically starts a new job. Each Pachyderm job consists of the
following stages:

### 1- **Creation of input datums** 
In this stage, Pachyderm **breaks
input files into datums according to the glob pattern** setting
in the pipeline specification.

When you create a pipeline specification, one of the most important
fields that you need to configure is **`pfs/`**, or PFS input.
The PFS input field is where you **define a data source** from which
the pipeline pulls data for further processing. The **`glob`**
parameter defines the number of datums in the `pfs/` source
repository. You can define everything in the source repository
to be processed as a single datum or break it down into multiple
datums. The way you break your source repository into datums
directly affects incremental processing and your pipeline
processing speed. You know your data better and can decide
how to optimize your pipeline based on the repository structure
and data generation workflows.
For more information about glob patterns, see
[Glob Pattern](glob-pattern.md).

### 2- **Transformation**
The pipeline **uses your code to process the datums**.

!!! Note "Reminder"
        The output produced by a pipeline's job is written to an output repo of the same name. However, you do not need to specify this name explicitly; you write your output to `/pfs/out`. 
### 3- **Creation of output datums**    

Pachyderm creates file(s) from the
processed data and combines them into output datums.

Disregarding how many datums you define and how many
filesystem objects a datum has, Pachyderm always matches the
number of input datums with the number of output datums. For
example, if you have three input datums in `pfs/`, you will
have three output datums in `pfs/out`. `pfs/out` is the
output repository that Pachyderm creates automatically for
each pipeline.    

### 4- **Final commit in the pipeline's output repo**

The output from multiple datums (each `/pfs/out`) is combined in a commit to the pipeline's output repo. 
This generally means unioning all the files together.

!!! Important "The Single Datum Provenance Rule"
     If two outputted files have the same name (i.e. two datums wrote to the same output file, creating a conflict), then an error is raised resulting in your pipeline failure. 
     An easy way to avoid this anti-pattern from the start is to **have each datum write in separate files.
### 5. **Next: Add a `Reduce` pipeline**

If you need files from different datums merged into single files in a particular way, you will need to **add a pipeline that groups the files into single datums using the appropriate glob pattern and merges them as intended using your code**. The example that follows illustrates this 2 steps approach. 

### Example: 2 Steps pattern and Single Datum Provenance Rule

In this example, we highlight a 2 pipelines pattern where a first pipeline's glob pattern splits an incoming commit into 3 datums (called "Red", "Blue", "Purple"), each producing 2 files each in their output datum.
The files in the output datums can be further
appended or overwritten with other files to create your final result. In the example below, the following pipeline appends the content of all files in each directory into one final document.


!!! Note "Worth Noting"
    - The files are named after the datum itself. Depending on your use case, there might be more logical ways to name the files produced by a datum. However, in any case, make sure that this **name is unique for each datum** to avoid duplicate
    files with the same file path.
    - Each file is put in specific directories. This directory structure has been thought to facilitate the aggregation of the content in the following pipeline. Think about your directory structure so that the next glob pattern will aggregate your data as needed.


![Map Reduce](../../../images/parallel_data_processing.png)


Let's now create a new commit and overwrite a file in `datum 2`,
Pachyderm detects three datums. However, because `datum 1` and `datum 3` are
unchanged, it skips processing these datums. Then, Pachyderm detects
that something has changed in `datum 2`. Pachyderm is unaware of any
details of the change; therefore, it processes the whole `datum 2`
and outputs 3 files. Then, the following pipeline aggregates
these data to create the following final result:

![Map Reduce](../../../images/parallel_data_processing_following_commit.png)

## Incrementality 
In Pachyderm, glob patterns are applied to the entire input, 
however, unchanged datums are never re-processed. 
For example, if you have multiple datums, 
and only one datum was modified; Pachyderm processes that datum only
and `skips` processing other datums. 
This incremental behavior ensures efficient resource utilization.

## Overwriting files
Another aspect of Pachyderm data processing is
appending and overwriting files. By default, Pachyderm
overwriting new data. For example, if you
have a file `foo` that is 100 KB in size in the repository `A`
and add the same file `foo` to that repository again by
using the `pachctl put file` command, Pachyderm will
overwrite that file in the repo. 
You can check that the size of the file `foo` still is 100KB. 

Pachyderm enables you to appends files as
well by using the `--appends` flag. The order of processing
is not guaranteed, and all datums are processed randomly.
For more information, see [File](../../data-concepts/file.md).
 
## Note: Data persistence between datums
Pachyderm only controls and wipes the `/pfs` directories between datums. If scratch/temp space is used during execution, the user needs to be careful to clean that up. Not cleaning temporary directories may cause unexpected bugs where one datum accesses temporary files that were previously used by another datum!
