# Split Data

Pachyderm provides functionality that enables you to split
your data while it is being loaded into a Pachyderm input
repository which helps to optimize pipeline processing
time and resource utilization.

Splitting data helps you to address the following use cases:

- **Optimize data processing**. If you have large size files,
  it might take Pachyderm a significant amount of time to process
  them. In addition, Pachyderm considers such a file as a single
  datum, and every time you apply even a minor change to that
  datum, Pachyderm processes the whole file from scratch.

- **Increase diff granularity**. Pachyderm does not create
  per-line diffs that display line-by-line changes. Instead,
  Pachyderm provides per-file diffing. Therefore, if all of your
  data is in one huge file, you might not be able to see what has
  changed in that file. Breaking up the file to smaller chunks addresses
  this issue.

This section provides examples of how you can use the `--split`
command that breaks up your data into smaller chunks, called
*split-files* and what happens when you update your data.
