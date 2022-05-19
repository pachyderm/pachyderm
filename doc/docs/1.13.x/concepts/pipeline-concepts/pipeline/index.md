# Pipeline


A pipeline is a Pachyderm primitive that is responsible for reading data
from a specified source, such as a Pachyderm repo, transforming it
according to the pipeline configuration, and writing the result
to an output repo.
A pipeline subscribes to a branch in one or more input repositories.
Every time the branch has a new commit, the pipeline executes a job
that runs your code to completion and writes the results to a commit
in the output repository. Every pipeline automatically creates
an output repository by the same name as the pipeline. For example,
a pipeline named `model` writes all results to the
`model` output repo.

In Pachyderm, a Pipeline is an individual execution step. You can
chain multiple pipelines together to create a directed acyclic
graph (DAG).

A minimum pipeline specification must include the following parameters:

- `name` — The name of your data pipeline. Set a meaningful name for
  your pipeline, such as the name of the transformation that the
  pipeline performs. For example, `split` or `edges`. Pachyderm
  automatically creates an output repository with the same name.
  A pipeline name must be an alphanumeric string that is less than
  63 characters long and can include dashes and underscores.
  No other special characters allowed.

- `input` — A location of the data that you want to process, such as a
  Pachyderm repository. You can specify multiple input
  repositories and set up the data to be combined in various ways.
  For more information, see [Cross and Union](../datum/cross-union.md), 
  [Join](../datum/join.md), [Group](../datum/group.md).
  One very important property that is defined in the `input` field
  is the `glob` pattern that specifies how Pachyderm breaks the data into
  individual processing units, called Datums. For more information, see
  [Datum](../datum/index.md).

- `transform` — Specifies the code that you want to run against your
  data. The `transform` section must include an `image` field that
  defines the Docker image that you want to
  run, as well as a `cmd` field for the specific code within the
  container that you want to execute, such as a Python script.

!!! example
    ```json
    {
      "pipeline": {
        "name": "wordcount"
      },
      "transform": {
        "image": "wordcount-image",
        "cmd": ["python3", "/my_python_code.py"]
      },
      "input": {
            "pfs": {
                "repo": "data",
                "glob": "/*"
            }
        }
    }
    ```

Pachyderm has the following special types of pipelines whose behavior might slightly differ from the 'regular' pipelines described above:

**Cron**
:   A cron input enables you to trigger the pipeline code at
    a specific interval. This type of pipeline is useful for
    such tasks as web scraping, querying a database, and other
    similar operations where you do not want to wait for new
    data, but instead trigger the pipeline periodically.

**Service**
:   A service is a special type of pipeline that instead of
    executing jobs and then waiting, permanently runs a serving
    data through an endpoint. For example, you can be serving
    an ML model or a REST API that can be queried. A service
    reads data from Pachyderm but does not have an output repo.

**Spout**
:   A spout is a special type of pipeline for ingesting data from
    a data stream. A spout can subscribe to a message stream, such
    as Kafka or Amazon SQS, and ingest data when it receives a
    message. A spout does not have an input repo.

!!! note "See Also:"
    [Pipeline Specification](../../../reference/pipeline-spec.md)
