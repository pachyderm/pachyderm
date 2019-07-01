# Datum

A datum is a unit of computation that Pachyderm processes and that
defines how data is written to output files. Depending on
the expected result, you can represent each dataset as a smaller
subset of data. You can configure a whole
input repository to be one datum, each filesystem object in
the root directory to be a separate datum,
filesystem objects in the subdirectories to be separate datums,
and so on. Datums affect how Pachyderm distributes processing workloads
among underlying compute nodes and are instrumental in optimizing
your configuration for best performance.

Pachyderm takes each datum and processes them in isolation on one of
the underlying worker nodes. Then, results from all the nodes are combined
and represented as one or multiple datums in the output repository.
All these properties can be configured through the user-defined variables in
the pipeline specification.

To understand how datum affects data processing in Pachyderm, you need to
understand the following subconcepts:

 * Glob pattern
 * Relationship between input and output datums
 * Parallelism
 * Incremental processing

## Relationship Between Datums

Pachyderm pipelines take a datum or datums from the source, processes
it or them, and places the result either as a single datum or multiple
datums to the output repository. You can configure how you want your
data to be processed, either as a single datum or as multiple datums.
You can also configure how you want the data to be uploaded to the output
repository - either as one datum, or file, or as multiple datums, or files.
This concept is called the relationship between datums.

Relationships between datums boil down to the following categories:

| Relationship       | Description                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------- |
| One-to-one (1-1)   | A unique datum in the input repository maps to a unique datum <br> in the output repository.  |
| One-to-many (1-M)  | A unique datum in the input repository maps to multiple datums <br> in the output repository. |
| Many-to-many (M-M) | Many unique datums in the input repository map to many datums  <br> in the output repository. |

### One-to-one Relationship

In a one-to-one relationship, Pachyderm takes one datum from the input
repository, transforms the data, and places the result as a single datum
in the output repository.

The [OpenCV example] illustrates the one-to-one relationship.

### One-to-many Relationship

In a one-to-many relationship, Pachyderm takes one datum from the input
repository and maps it to many datums in the output repository. For example,
the image in the OpenCV example can be split into many tiles for further
analysis. Each tile represents a unique datum.

### Many-to-many Relationship

Most of the pipelines that users create in Pachyderm have many-to-many
relationships. The pipeline takes many unique datums from the input
repository or repositories and writes to many output files in the
output repository. The output files might not be unique across
datums. Therefore, the results from processing many datums can
be represented in a single file. Pachyderm can merge all the
results into a single file. When you run `pachctl list jobs`,
you can sometimes see the `merging` status in the output of the
command, which means that Pachyderm optimizes the result by
presenting it in a single file.

The [Wordcount example](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count)
illustrates the many-to-many relationship

## Glob Pattern

A glob pattern defines datum in your Pachyderm repository and how
your data processing is spread among Pachyderm workers. The whole
repository, each filesystem object in the root folder,
each filesystem object in a subdirectory can be a standalone datum,
and so on.

Defining how your data is spread out among workers is one of
the most important aspects of distributed computation and is
the fundamental idea around concepts, such as Map and Reduce.

Instead of confining users to data-distribution patterns,
such as Map — split everything as much as possible, — and
Reduce — all the data must be grouped together — Pachyderm
uses glob patterns to provide incredible flexibility in
defining your data distribution.

You can define a glob pattern for each `pfs` within the input
of a pipeline, and they tell Pachyderm how to
divide the input data into individual *datums* that can
be processed independently.

Think of the PFS input repository as a filesystem where
the glob pattern is applied to the root of the
filesystem. The files and directories that match the
glob pattern are considered datums.
Pachyderm accepts the `/` and `*` indicators as
glob patterns.

Examples of glob patterns that you can define:

* `/` — Pachyderm processes the whole repository as a
  single datum and send the data to a single worker node.
* `/*` — Pachyderm processes each top-level filesystem
  object, such as a file or directory, in
  the root directory as a separate datum. For example,
  if you have a repository with ten files in it and no
  directory structure, Pachyderm handles each file as a
  single datum, similarly to a map operation.
* `/*/*` — Pachyderm processes each filesystem object
  in each subdirectory as a separate datum.

The following text is an extract from a pipeline specification.
It shows how you can define glob pattern for each individual PFS
input:

```
"input": {
    "cross" or "union": [
        {
            "pfs": {
                "repo": "string",
                "glob": "string",
            }
        },
        {
            "pfs": {
                "repo": "string",
                "glob": "string",
            }
        },
        etc...
    ]
}
```

For example, if you have two workers and define two datums,
Pachyderm sends one datum to each worker. If you have more
datums than workers, Pachyderm queues datums and sends
them to workers as they finish processing previous datums.

### Example of Defining Data

For example, you have the following directory:

```bash
/California
   /San-Francisco.json
   /Los-Angeles.json
   ...
/Colorado
   /Denver.json
   /Boulder.json
   ...
...
```

Each top-level directory represents a United States(US) state with
`json` files in it that represent cities of that state.

If you need to process all the data for each state together in one
batch, you need to define `/*` as a glob pattern. Pachyderm
defines one datum per state which means that all the cities for
a given state are processed together by a single worker, but each
state can be processed independently.

If you set `/*/*` for states, Pachyderm processes each city as a single
datum on a separate worker.

If you want to process a specific directory or a subset of directories
as a PFS input instead of the whole input repository. For example,
you can specify `/California/*` to process only the data in the
`California/` directory.

### Glob Patterns and Incremental Processing

A datum defines the granularity at which Pachyderm
decides what data is new and what data has already
been processed. When Pachyderm detects a new data
in any part of a datum, it reprocess the whole datum.
This section illustrates how glob patterns
affect incremental processing in Pachyderm.

For example, you have the same repository
structure as in the previous section with
the US states as directories and US cities
as `.json` files:

```bash
/California
   /San-Francisco.json
   /Los-Angeles.json
   ...
/Colorado
   /Denver.json
   /Boulder.json
   ...
...
```

Setting a different glob pattern affects your
repository as follows:

* If you set glob pattern to `/`, every time
you change anything in any of the
files and folders or add a new file to the
repository, Pachyderm processes the whole
repository fromn scratch.

* If you set the glob pattern to `/*`, Pachyderm treats each state
directory as a separate datum. For example, if you add a new file
`Sacramento.json` to the `/California` directory, Pachyderm
processes the `/California` datum only.

* If you set the glob pattern to `/*/*`, Pachyderm processes each
`<city>.json` file as its own datum. For example, if you add
the `Sacramento.json` file, Pachyderm processes the
`Sacramento.json` file only.
