# Combine, Merge, and Join Data

!!! info Before you read this section, make sure that you understand the
concepts described in [Distributed Processing](distributed_computing.md).

In some of your projects, you might need to match datums from multiple data
repositories to process, join, or aggregate data. For example, you might need to
process together multiple records that correspond to a certain user, experiment,
or device.

In these cases, you can create two pipelines that perform the following steps:

![Steps](../assets/images/d_steps_combine_pipelines.svg)

More specifically, you need to create the following pipelines:

1. Create a pipeline that groups all of the records for a specific key and
   index.

2. Create another pipeline that takes that grouped output and performs the
   merging, joining, or other processing for the group.

You can use these two data-combining pipelines for merging or grouped processing
of data from various experiments, devices, and so on. You can also apply the
same pattern to perform distributed joins of tabular data or data from database
tables. For example, you can join user email records together with user IP
records on the key and index of a user ID.

You can parallelize each of the stages across workers to scale with the size of
your data and the number of data sources that you want to merge.

!!! tip If your data is not split into separate files for each record, you can
split it automatically as described in
[Splitting Data for Distributed Processing](splitting-data/splitting.md).

## Group Matching Records

The first pipeline that you create groups the records that need to be processed
together.

In this example, you have two repositories `A` and `B` with JSON records. These
repositories might correspond to two experiments, two geographic regions, two
different devices that generate data, or other.

The following diagram displays the first pipeline:

![alt tag](../assets/images/d_join1.svg)

The repository `A` has the following structure:

```bash
$ pachctl list file A@master
NAME                TYPE                SIZE
1.json              file                39 B
2.json              file                39 B
3.json              file                39 B
```

The repository `B` has the following structure:

```bash
$ pachctl list file B@master
NAME                TYPE                SIZE
1.json              file                39 B
2.json              file                39 B
3.json              file                39 B
```

If you want to process `A/1.json` with `B/1.json` to merge their contents or
otherwise process them together, you need to group each set of JSON records into
respective datums that the pipelines that you create in

[Process Grouped Records](#process-grouped-records) can process together.

The grouping pipeline takes a union of `A` and `B` as inputs, each with glob
pattern `/*`. While the pipeline processes a JSON file, the data is copied to a
folder in the output that corresponds to the key and index for that record. In
this example, it is just the number in the file name. Pachyderm also renames the
files to unique names that correspond to the source:

```bash
/1
  A.json
  B.json
/2
  A.json
  B.json
/3
  A.json
  B.json
```

When you group your data, set the following parameters in the pipeline
specification:

-   In the `pfs` section, set `"empty_files": true` to avoid unnecessary
    downloads of data.

-   Use symlinks to avoid unnecessary uploads of data and unnecessary data
    duplication.

## Process Grouped Records

After you group the records together by using the grouping pipeline, you can use
a merging pipeline on the `group` repository with a glob pattern of `/*`. By
using the glob pattern of `/*` the pipeline can process each grouping of records
in parallel.

THe following diagram displays the second pipeline:

![alt tag](../assets/images/d_join2.svg)

The second pipeline performs merging, aggregation, or other processing on the
respective grouping of records. It can also output each respective result to the
root of the output directory:

```bash
$ pachctl list file merge@master
NAME                TYPE          SIZE
result_1.json       file          39 B
result_2.json       file          39 B
result_3.json       file          39 B
```
