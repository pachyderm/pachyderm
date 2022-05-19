# Storage Use Optimization

This section discusses best practices for minimizing the
space needed to store your Pachyderm data, increasing
the performance of your data processing as related to
data organization, and general good ideas when you
are using Pachyderm to version/process your data.

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

!!! note "See Also:"

   [Troubleshoot a pipeline](../../../troubleshooting/pipeline-troubleshooting#all-your-pods-or-jobs-get-evicted)
