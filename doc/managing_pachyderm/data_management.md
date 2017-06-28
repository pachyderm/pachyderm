# Data Management Best Practices

This document discusses best practices for minimizing the space needed to store your Pachyderm data, increasing the performance of your data processing as related to data organization, and general good ideas when you are using Pachyderm to version/process your data.

- [Shuffling files](#shuffling-files)
- [Garbage collection](#garbage-collection)
- [Setting a root volume size](#setting-a-root-volume-size)

## Shuffling files

Certain pipelines simply shuffle files around (e.g., organizing files into buckets).  If you find yourself writing a pipeline that does a lot of copying, such as [Time Windowing](http://docs.pachyderm.io/en/latest/cookbook/time_windows.html), it probably falls into this category.

The best way to shuffle files, especially large files, is to create **symlinks** in the output directory that point to files in the input directory.

For instance, to move a file `log.txt` to `logs/log.txt`, you might be tempted to write a [`transform`](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#transform-required) like this:

```sh
cp /pfs/input/log.txt /pfs/out/logs/log.txt
```

However, it's more efficient to create a symlink:

```sh
ln -s /pfs/input/log.txt /pfs/out/logs/log.txt
```

Under the hood, Pachyderm is smart enough to recognize that the output file simply symlinks to a file that already exists in Pachyderm, and therefore skips the upload altogether.

Note that if your shuffling pipeline only needs the names of the input files but not their content, you can use [`lazy input`](http://pachyderm.readthedocs.io/en/latest/reference/pipeline_spec.html#atom-input).  That way, your shuffling pipeline can skip both the download and the upload.

## Garbage collection

When a file/commit/repo is deleted, the data is not immediately removed from the underlying storage system (e.g. S3) for performance and architectural reasons.  This is similar to how when you delete a file on your computer, the file is not necessarily wiped from disk immediately.

To actually remove the data, you may need to manually invoke garbage collection.  The easiest way to do it is through `pachctl garbage-collect`.  Currently `pachctl garbage-collect` can only be started when there are no active jobs running.  You also need to ensure that there's no ongoing `put-file`.  Garbage collection puts the cluster into a readonly mode where no new jobs can be created and no data can be added.

## Setting a root volume size

When planning and configuring your Pachyderm deploy, you need to make sure that each node's root volume is big enough to accomodate your total processing bandwidth. Specifically, you should calculate the bandwidth for your expected running jobs as follows:

```
(storage needed per datum) x (number of datums being processed simultaneously) / (number of nodes)
```

Here, the storage needed per datum should be the storage needed for the largest "datum" you expect to process anywhere on your DAG plus the size of the output files that will be written for that datum.  If your root volume size is not large enough, pipelines might fail when downloading the input. The pod would get evicted and rescheduled to a different node, where the same thing will happen (assuming that node had a similar volume). This scenario is further discussed [here](general_troubleshooting.html#all-your-pods--jobs-get-evicted).
