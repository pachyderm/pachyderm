# Datum

!!! note "TL;DR"
    Datums define what input data is seen by your code. It can be
    all data at once, each directory independently, individual
    files one by one, or combined data from multiple inputs together.

## Definition
A datum is **the smallest indivisible unit of computation within a job**.
A job can have one, many or no datums. Each datum is processed
independently with a single execution of the user code on one of
the pipeline worker pods.
The files output by all of the datums are then combined together to
create the final output commit.

## Data distribution
Think of datums as a way to **divide your input data** 
and **distribute processing workloads**.
They are instrumental in optimizing your pipeline performance.

You define how your data is spread among workers by
**specifying [pipeline inputs](#pipeline-inputs)** for your pipeline 
in its pipeline specification file. 

Based on this specification file, the data in the `input` 
of your pipeline is turned in datums 
each of which can contain 1 to many files.
Pachyderm provides a wide variety of ways to define the granularity of each datum. 

For example, you can configure a whole branch of an
input repository to be one datum, each top-level filesystem object of a given branch
to be a separate datum, specific paths on a given branch can be datums, etc...
You can also create a combination of the above by aggregating multiple input.

## Pipeline Inputs
This section details the tools at your disposal to
"break down" your data and fit your specific use case.
### PFS Input and Glob Pattern
The most primitive input of a pipeline is a [**PFS input**](../../../reference/pipeline-spec/#pfs-input), defined, at a minimum, by:

- a repo containing the data you want your pipeline to consider
- a branch to watch for commits
- and **a [glob pattern](glob-pattern.md) to determine how the input data is partitioned**.

A pipeline input can have one or multiple PFS inputs.
In the latter case, Pachyderm provides a variety of options to aggregate several PFS inputs together.
### Aggregating Datums Further
Your initial PFS inputs can be further combined by using the following
aggregation methods, each configurable in the "input" section of the pipeline specification:

[**Cross**](./cross-union/#cross-input)
:    A cross input creates a cross-product of multiple repositories.
     Therefore, each datum from one repository is combined with each
     datum from the other repository.

[**Group**](./group/)
:    A group input can take one or multiple repositories and processes
     all the data that match a particular
     file path pattern at once. Pachyderm's group is similar to the database *group-by*, but it matches on file paths only, not the content of the files.

[**Union**](./cross-union/#union-input)
:    A union input can take multiple repositories and processes
     all the data in each input independently. The pipeline
     processes the datums in no defined order and the output
     repository includes results from all input sources.

[**Joins**](./join/)
:    A join input enables you to join files that are stored
     in different Pachyderm repositories and match a particular
     file path pattern. Pachyderm supports joins similar to the
     database’s *inner join* and *outer join* operations, although they only match
     on file paths, not the actual file content.


!!! Important
     Pipeline inputs, number of workers, parallelism parameters, and other
     performance knobs can all be configured through their
     corresponding fields in the [pipeline specification](../../../reference/pipeline-spec.md).

To understand how datums affect data processing in Pachyderm, you need to
understand the following subconcepts:

* [Glob Pattern](glob-pattern.md)
* [Datum Processing](relationship-between-datums.md)
