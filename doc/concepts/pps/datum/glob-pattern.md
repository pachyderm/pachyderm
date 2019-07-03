# Glob Pattern

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
Reduce — all the data must be grouped — Pachyderm
uses glob patterns to provide incredible flexibility in
defining your data distribution.

You can define a glob pattern for each PFS input within the input
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
It shows how you can define glob pattern for individual PFS
inputs:

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

## Example of Defining Data

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
defines one datum per state, which means that all the cities for
a given state are processed together by a single worker, but each
state can be processed independently.

If you set `/*/*` for states, Pachyderm processes each city as a single
datum on a separate worker.

If you want to process a specific directory or a subset of directories
as a PFS input instead of the whole input repository,
you can, for example, specify `/California/*` to process only the data in the
`California/` directory.

