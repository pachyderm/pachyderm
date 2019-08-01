# Glob Pattern

Defining how your data is spread among workers is one of
the most important aspects of distributed computation and is
the fundamental idea around concepts such as Map and Reduce.

Instead of confining users to data-distribution patterns,
such as Map, that splits everything as much as possible, and
Reduce, that groups all the data, Pachyderm
uses glob patterns to provide incredible flexibility to
define data distribution.

You can configure a glob pattern for each PFS input in
the input field of a pipeline specification. Pachyderm detects
this parameter and divides the input data into
individual *datums*.

You can think of each input repository as a filesystem where
the glob pattern is applied to the root of the
filesystem. The files and directories that match the
glob pattern are considered datums. The Pachyderm's
concept of glob patterns is similar to the Unix glob patterns.
For example, the `ls *.md` command matches all files with the
`.md` file extension.

Pachyderm accepts the `/` and `*` indicators as
glob patterns.

The following are examples of glob patterns that you can define:

* `/` — Pachyderm processes the whole repository as a
  single datum and sends the data to a single worker node.
* `/*` — Pachyderm processes each top-level filesystem
  object, such as a file or directory, in
  the root directory as a separate datum. For example,
  if you have a repository with ten files in it and no
  directory structure, Pachyderm handles each file as a
  single datum, similar to a map operation.
* `/*/*` — Pachyderm processes each filesystem object
  in each subdirectory as a separate datum.

<!-- Add the ohmyglob examples here-->

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

## Example of Defining Datums

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

Each top-level directory represents a United State (US) with
`json` files in it that represent cities of that state.

* If you set glob pattern to `/`, every time
you change anything in any of the
files and folders or add a new file to the
repository, Pachyderm processes the whole
repository from scratch.

If you need to process all the data for each state together in one
batch, you need to define `/*` as a glob pattern. Pachyderm
defines one datum per state, which means that all the cities for
a given state are processed together by a single worker, but each
state can be processed independently. For example, if you add a new file
`Sacramento.json` to the `California/` directory, Pachyderm
processes the `California/` datum only.

If you set `/*/*` for states, Pachyderm processes each city as a single
datum on a separate worker. For example, if you add
the `Sacramento.json` file, Pachyderm processes the
`Sacramento.json` file only.

If you want to process a specific directory or a subset of directories
as a PFS input instead of the whole input repository,
you can specify `/<state>/*` directory to process only the data in the
`<state>/` directory.

