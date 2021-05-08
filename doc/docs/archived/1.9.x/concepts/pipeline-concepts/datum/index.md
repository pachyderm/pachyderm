# Datum

!!! note "TL;DR"
    Datums define what input data is seen by your code. It can be
    all data at once, each directory independently, individual
    files one by one, or combined data from multiple inputs together.

A datum is the smallest indivisible unit of computation within a job.
A job can have one, many or no datums. Each datum is processed
independently with a single execution of the user code and
then the results of all the datums are merged together to
create the final output commit.

Datums define what input data is seen by your code. An input can take one or multiple
repositories. Pachyderm has the following types of inputs that
combine multiple repositories:

**Cross**
:    A cross input creates a cross-product of multiple repositories.
     Therefore, each datum from one repository is combined with each
     datum from the other repository.

**Join**
:    A join input enables you to join files that are stored
     in different Pachyderm repositories and match a particular
     file path pattern. Joins are similar to `cross`, except
     instead of matching every pair of datums from each input,
     it only matches specific ones based on file paths.
     Conceptually, joins are similar to the
     database’s inner join operations, although they only match
     on file paths, not the actual file content.

**Union**
:    A union input can take multiple repositories and processes
     all the data in each input independently. The pipeline
     processes the datums in no defined order and the output
     repository includes results from all input sources.

The number of datums for a job is defined by the
[glob pattern](glob-pattern.md) which you specify for each input. Think of
datums as if you were telling Pachyderm how to divide your
input data to efficiently distribute computation and
only process the *new* data. You can configure a whole
input repository to be one datum, each top-level filesystem object
to be a separate datum, specific paths can be datums,
and so on. Datums affect how Pachyderm distributes processing workloads
and are instrumental in optimizing your configuration for best performance.

Pachyderm takes each datum and processes it in isolation on one of
the pipeline worker nodes. You can define datums, workers, and other
performance parameters can all be configured through the
corresponding fields in the [pipeline specification](../../../reference/pipeline_spec.md).

To understand how datums affect data processing in Pachyderm, you need to
understand the following subconcepts:

* [Glob Pattern](glob-pattern.md)
* [Datum Processing](relationship-between-datums.md)
