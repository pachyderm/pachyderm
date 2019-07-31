.. _datum:

Datum
=====

A datum is a unit of computation that Pachyderm processes and that
defines how data is written to output files. Depending on
the results that you want to achieve, you can represent each
dataset as a smaller subset of data. You can configure a whole
input repository to be one datum, each filesystem object in
the root directory to be a separate datum,
filesystem objects in the subdirectories to be separate datums,
and so on. Datums affect how Pachyderm distributes processing workloads
among underlying Pachyderm workers and are instrumental in optimizing
your configuration for best performance.

Pachyderm takes each datum and processes them in isolation on one of
the underlying worker nodes. Then, results from all the nodes are combined
and represented as one or multiple datums in the output repository.
All these properties can be configured through the user-defined variables in
the pipeline specification.

To understand how a datum affects data processing in Pachyderm, you need to
understand the following subconcepts:

.. toctree::
   :maxdepth: 1

   glob-pattern.md
   relationship-between-datums.md
