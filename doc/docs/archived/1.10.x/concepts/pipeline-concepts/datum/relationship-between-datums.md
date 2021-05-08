# Datum Processing

This section helps you to understand the following
concepts:

* Pachyderm job stages
* Multiple datums processing
* Incremental processing
* Data persistence between datums

A datum is a Pachyderm abstraction that helps in optimizing
pipeline processing. Because datums exist only as a pipeline
processing property and are not filesystem objects, you can never
list or copy a datum. Instead, a datum, as a representation of a unit
of work, helps you to run your pipelines much faster by avoiding
repeated processing of unchanged datums. For example, if you have
multiple datums, and only one datum was modified, Pachyderm processes
only that datum and skips processing other datums. This incremental
behavior ensures efficient resource utilization.

Each Pachyderm job can process multiple datums, which can consist
of one or multiple files. While each input datum results in one output
datum, the number of files in the output datum might differ from
the number of files in the input datum.

When you create a pipeline specification, one of the most important
fields that you need to configure is `pfs/`, or PFS input.
The PFS input field is where you define a data source from which
the pipeline pulls data for further processing. The `glob`
parameter defines the number of datums in the `pfs/` source
repository. Thus, you can define everything in the source repository
to be processed as a single datum or break it down to multiple
datums. The way you break your source repository into datums
directly affects incremental processing and your pipeline
processing speed. You know your data better and can decide
how to optimize your pipeline based on the repository structure
and data generation workflows.
For more information about glob patterns, see
[Glob Pattern](glob-pattern.md).

Disregarding of how many datums you define and how many
filesystem objects a datum has, Pachyderm always matches the
number of input datums with the number of output datums. For
example, if you have three input datums in `pfs/`, you will
have three output datums in `pfs/out`. `pfs/out` is the
output repository that Pachyderm creates automatically for
each pipeline. You can add your changes in any order and
submit them in one or multiple commits, the result of your
pipeline processing remains the same.

Another aspect of Pachyderm data processing is
appending and overwriting files. By default, Pachyderm
appends new data to the existing data. For example, if you
have a file `foo` that is 100 KB in size in the repository `A`
and add the same file `foo` to that repository again by
using the `pachctl put file` command, Pachyderm does not
overwrite that file but appends it to the file `foo` in the
repo. Therefore, the size of the file `foo` doubles and
becomes 200 KB. Pachyderm enables you to overwrite files as
well by using the `--overwrite` flag. The order of processing
is not guaranteed, and all datums are processed randomly.
For more information, see [File](../../data-concepts/file.md).

When new data comes in, a Pachyderm pipeline automatically
starts a new job. Each Pachyderm job consists of the
following stages:

1. Creation of input datums. In this stage, Pachyderm breaks
input files into datums according to the glob pattern setting
in the pipeline specification.
1. Transformation. The pipeline uses your code to processes the
datums.
1. Creation of output datums. Pachyderm creates file or files from the
processed data and combines them into output datums.
1. Merge. Pachyderm combines all files with the same file path
by appending, overwriting, or deleting them to create the final commit.

If you think about this process in terms of filesystem objects and
processing abstractions, the following transformation happens:

!!! note ""
    **input files = > input datums => output datums => output files**

This section provides examples that help you understand such fundamental
Pachyderm concepts as the datum, incremental processing, and phases of
data processing.

## Example 1: One file in the input datum, one file in the output datum

The simplest example of datum processing is when you have one file in
the input datum that results in one file in the output datum.
In the diagram below, you can see three input datums, each of which
includes one file, that result in three output datums. Whether you have
submitted all these datums in a single or multiple commits, the final
result remains the same—three datums, each of which has one file.

In the diagram below, you can see the following datums:

 - `datum 1` has one file and results in one file in one output datum.
 - `datum 2` has one file and results in one file in one output datum.
 - `datum 3` has one file and results in one file in one output datum.

![One to one](../../../assets/images/d_datum_processing_one_to_one.svg)

If you decide to overwrite a single line in the file in `datum 3` and
add `datum 4`, Pachyderm sees the four datums and checks them for changes
one-by-one. Pachyderm verifies that there are no changes in `datum 1` and
`datum 2` and skips these datums. Pachyderm detects changes in the
`datum 3` and the `--overwrite` flag and replaces the `datum 3` with the
new `datum 3'`. When it detects `datum 4` as a completely new datum,
it processes the whole datum as new. Although only two datums were
processed, the output commit of this change contains all four files.

![One to one overwrite](../../../assets/images/d_datum_processing_one_to_one_overwrite.svg)

## Example 2: One file in the input datum, multiple files in the output datum

Some pipelines ingest one file in one input datum and create multiple
files in the output datum. The files in the output datums might need to
be appended or overwritten with other files to create the final commit.

If you apply changes to that datum, Pachyderm does not detect which
particular part of the datum has changed and processes the whole datum.
In the diagram below, you have the following datums:

- `datum 1` has one file and results in files `1` and `3`.
- `datum 2` has one file and results in files `2` and `3`.
- `datum 3` has one file and results in files `2` and `1`.

![One to many](../../../assets/images/d_datum_processing_one_to_many.svg)

Pachyderm processes all these datums independently, and in the end, it needs
to create a commit by combining the results of processing these datums.
A commit is a filesystem that has specific constraints, such as duplicate
files with the same file path. Pachyderm merges results from
different output datums with the same file path into single files. For
example, `datum 1` produces `pfs/out/1` and `datum 3` produces `pfs/out/1`.
Pachyderm merges these two files by appending them one to another
without any particular order. Therefore, the file `1` in the final
commit has parts from `datum1` and `datum3`.

If you decide to create a new commit and overwrite the file in `datum 2`,
Pachyderm detects three datums. Because `datum 1` and `datum 3` are
unchanged, it skips processing these datums. Then, Pachyderm detects
that something has changed in `datum 2`. Pachyderm is unaware of any
details of the change. Therefore, it processes the whole `datum 2`
and outputs the files `1`, `3`, and `4`. Then, Pachyderm merges
these datums to create the following final result:

![One to many](../../../assets/images/d_datum_processing_one_to_many_overwrite.svg)


In the diagram above, Pachyderm appends the file `1` from the `datum 2`
to the file `1` in the final commit, deletes the file `2` from `datum 2`,
overwrites the old part from `datum 2` in file `3`  with a new version,
and creates a new output file `4`.

Similarly, if you have multiple files in your input datum, Pachyderm might
write them into multiple files in output datums that are later merged into
files with the same file path.

## Note: Data persistence between datums

Pachyderm only controls and wipes the /pfs directories between datums. If scratch/temp space is used during execution, the user needs to be careful to clean that up. Not cleaning temporary directories may cause unexpected bugs where one datum accesses temporary files that were previously used by another datum!
