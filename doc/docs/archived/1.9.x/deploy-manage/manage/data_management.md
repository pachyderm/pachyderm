# Storage Use Optimization

This section discusses best practices for minimizing the
space needed to store your Pachyderm data, increasing
the performance of your data processing as related to
data organization, and general good ideas when you
are using Pachyderm to version/process your data.

## Garbage collection

When a file, commit, repo is deleted, the data is not immediately removed
from the underlying storage system, such as S3, for performance and
architectural reasons. This is similar to how when you delete a file
on your computer, the file is not necessarily wiped from disk immediately.

To actually remove the data, you may need to manually invoke garbage
collection. The easiest way to do it is through `pachctl garbage-collect`.
You can start `pachctl garbage-collect` only when no active jobs are
running. Also, you need to ensure that all `pachctl put file` operations
have been completed. Garbage collection puts the cluster into a read-only
mode where no new jobs can be created and no data can be added.

## Setting a root volume size

When planning and configuring your Pachyderm deployment, you need to
make sure that each node's root volume is big enough to accommodate
your total processing bandwidth. Specifically, you should calculate
the bandwidth for your expected running jobs as follows:

```shell
(storage needed per datum) x (number of datums being processed simultaneously) / (number of nodes)
```

Here, the storage needed per datum must be the storage needed for
the largest datum you expect to process anywhere on your DAG plus
the size of the output files that will be written for that datum.
If your root volume size is not large enough, pipelines might fail
when downloading the input. The pod would get evicted and
rescheduled to a different node, where the same thing might happen
(assuming that node had a similar volume).

!!! note "See also:"

   [Troubleshoot a pipeline](../../../troubleshooting/pipeline_troubleshooting#all-your-pods-or-jobs-get-evicted)
