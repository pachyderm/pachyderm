# Datum Processing

A [datum](../) is a Pachyderm abstraction that helps in optimizing
pipeline processing. Datums exist only as a pipeline
processing property and are not filesystem objects. You can never
copy a datum. They are a representation of a unit
of work.
## Job Processing and Datums
When new data comes in (in the form of commit(s) in its input repo(s)), a Pachyderm pipeline automatically starts a new job. Each Pachyderm job consists of the
following stages:

### 1- **Creation of input datums** 
In this stage, Pachyderm **creates datums based on the input data according to the
[pipeline input(s)](../#pipeline-inputs)** set
in the [pipeline specification file](../../../../reference/pipeline-spec/#pipeline-specification).

### 2- **Transformation**
The pipeline **uses your code to process the datums**.

### 3- **Creation of output files**    
Your code writes output file(s) in the
`pfs/out` output directory that Pachyderm 
creates automatically for
each pipeline's job.    

### 4- **Final commit in the pipeline's output repo**

!!! Note "Reminder"
        The output produced by a pipeline's job is written to an output repo of the same name (i.e., output repo name = pipeline name).

The content of all `/pfs/out` is combined in a commit to the pipeline's output repo. 
This generally means unioning all the files together.

!!! Important "The Single Datum Provenance Rule"
     If two outputted files have the same name (i.e., two datums wrote to the same output file, creating a conflict), then an error is raised, resulting in your pipeline failure. 

     Avoid this anti-pattern from the start by **having each datum write in separate files**. Pachyderm provides an **environment variable `PACH_DATUM_ID`** that stores the datum ID. This variable is available in the pipeline's user code. To ensure that each datum outputs distinct file paths, you can use this variable in the name of your outputted files.
### 5. **Next: Add a `Reduce` (Merge) pipeline**

If you need files from different datums merged into single files in a particular way:

- **add a pipeline that groups the files** from the previous output repo into single datums using the appropriate glob pattern.
- then **merge them as intended using your code**. 

The example that follows illustrates this two steps approach. 

### Example: Two Steps Map/Reduce Pattern and Single Datum Provenance Rule

In this example, we highlight a two pipelines pattern where a first pipeline's glob pattern splits an incoming commit into three datums (called "Datum1" (Red), "Datum2" (Blue), "Datum3" (Purple)), each producing two files each.
The files can then be further
appended or overwritten with other files to create the final result. Below, a second pipeline appends the content of all files in each directory into one final document.


!!! Note "Worth Noting"
    - In the example, the files are named after the datum itself. Depending on your use case, there might be more logical ways to name the files produced by a datum. However, in any case, make sure that this **name is unique for each datum** to avoid duplicate
    files with the same file path.
    - Each file is put in specific directories. This directory structure has been thought to facilitate the aggregation of the content in the following pipeline. Think about your directory structure so that the next glob pattern will aggregate your data as needed.


![Map Reduce](../../../images/parallel_data_processing.png)


Let's now create a new commit and overwrite a file in `datum 2`,
Pachyderm detects three datums. However, because `datum 1` and `datum 3` are
unchanged, it skips processing these datums. Pachyderm detects
that something has changed in `datum 2`. It is unaware of any
details of the change; therefore, it processes the whole `datum 2(')` (here in yellow)
and outputs 3 files. Then, the following pipeline aggregates
these data to create the final result.

![Map Reduce](../../../images/parallel_data_processing_following_commit.png)

## Incrementality 
In Pachyderm, unchanged datums are never re-processed. 
For example, if you have multiple datums, 
and only one datum was modified; Pachyderm processes that datum only
and `skips` processing other datums. 
This incremental behavior ensures efficient resource utilization.

## Overwriting files

By default, Pachyderm
overwrites new data. For example, if you
have a file `foo` in the repository `A`
and add the same file `foo` to that repository again by
using the `pachctl put file` command, Pachyderm will
overwrite that file in the repo. 

For more information, and learn how to change this behavior, see [File](../../../data-concepts/file/).
 
## Note: Data persistence between datums
Pachyderm only controls and wipes the `/pfs` directories between datums. If scratch/temp space is used during execution, the user needs to be careful to clean that up. Not cleaning temporary directories may cause unexpected bugs where one datum accesses temporary files that were previously used by another datum!
