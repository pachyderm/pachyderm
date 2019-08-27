# Cron Pipeline

Pachyderm triggers pipelines when new changes appear in the input repository.
However, if you want to trigger a pipeline based on time instead of upon
arrival of input data, you can schedule such pipelines to run periodically
by using the built-in `cron` input type.

Cron inputs are well suited for a variety of use cases, including
the following:

- Scraping websites
- Making API calls
- Querying a database
- Retrieving a file from a location accessible through an S3 protocol
or a File Transfer Protocol (FTP).

A minimum cron pipeline must include the following parameters:

| Parameter  | Description  |
| ---------- | ------------ |
| `"name"`   | A descriptive name of the cron pipeline. |
| `"spec"`   | The `spec` parameter defines when to trigger a cron job. You can specify any value that is <br> formatted according to [RFC 3339](https://www.ietf.org/rfc/rfc3339.txt). <br> For example, if you set `*/10 * * * *`, the pipeline runs every ten minutes. |

## Example of a Cron Pipeline

For example, you want to query a database every ten seconds and update your
dataset with the new data every time the pipeline is triggered. The following
pipeline extract illustrates how you can specify this configuration:

```
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 10s"
    }
  }
```

When you create this pipeline, Pachyderm creates a new input data repository
that corresponds to the `cron` input. Then, Pachyderm automatically commits
a timestamp file to the `cron` input repository every ten seconds, which
triggers the pipeline.

![alt tag](../../../images/cron1.png)

The pipeline runs every ten seconds, queries the database, and updates its
output. By default, each cron trigger adds a new tick file to the cron input
repository, accumulating more datums over time. This behavior works for some
pipelines. For others, you can set the `--overwrite` flag to `true` to
overwrite the timestamp file on each tick. To learn more about overwriting files
within a commit as opposed to adding a new file, and how that
affects the datums in the subsequent jobs, see [Datum processing](../datum/index.rst)

**Example:**

```
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 10s",
      "overwrite": true
    }
  }
```

**See Also:**

- [Periodic Ingress from MongoDB](https://github.com/pachyderm/pachyderm/tree/master/examples/db)
