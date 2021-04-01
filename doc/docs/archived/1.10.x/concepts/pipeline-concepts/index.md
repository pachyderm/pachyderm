# Pipeline Concepts

Pachyderm Pipeline System (PPS) is the computational
component of the Pachyderm platform that enables you to
perform various transformations on your data. Pachyderm
pipelines have the following main concepts:

**Pipeline**
:   A pipeline is a job-spawner that waits for certain
    conditions to be met. Most commonly, this means
    watching one or more Pachyderm repositories for new
    data. When a new data arrives, a pipeline executes
    a user-defined piece of code to perform an operation
    and process the data. Each of these executions is
    called a job.

**Job**
:   A job is an individual execution of a pipeline. A job
    can succeed or fail. Within a job, data and processing
    can be broken up into individual units of work called datums.

**Datum**
:   A datum is the smallest indivisible unit of work within
    a job. Different datums can be processed in parallel
    within a job.
