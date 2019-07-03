# Glob Patterns and Incremental Processing

<!--- Add more information from ../../../fundamentals/incrementality.html-->

A datum defines the granularity at which Pachyderm
decides what data is new and what data has already
been processed. When Pachyderm detects a new data
in any part of a datum, it reprocesses the whole datum.
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
repository from scratch.

* If you set the glob pattern to `/*`, Pachyderm treats each state
directory as a separate datum. For example, if you add a new file
`Sacramento.json` to the `/California` directory, Pachyderm
processes the `/California` datum only.

* If you set the glob pattern to `/*/*`, Pachyderm processes each
`<city>.json` file as its own datum. For example, if you add
the `Sacramento.json` file, Pachyderm processes the
`Sacramento.json` file only.
