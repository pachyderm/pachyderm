# Triggering Pipelines Periodically (cron)

Pachyderm pipelines are triggered by changes to their input data repositories (as further discussed in [What Happens When You Create a Pipeline](../getting_started/beginner_tutorial.html#what-happens-when-you-create-a-pipeline)). However, if a pipeline consumes data from sources outside of Pachyderm, it can't use Pachyderm's triggering mechanism to process updates from those sources. For example, you might need to:

- Scrape websites
- Make API calls
- Query a database
- Retrieve a file from S3 or FTP

You can schedule pipelines like these to run regularly with Pachyderm's built-in `cron` input type. You can find an example pipeline that queries MongoDB periodically [here](https://github.com/pachyderm/pachyderm/tree/master/examples/db).

## Cron Example

Let's say that we want to query a database every 10 seconds and update our dataset every time the pipeline is triggered. We could do this with `cron` input as follows:

```
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 10s"
    }
  }
```

When we create this pipeline, Pachyderm will create a new input data repository corresponding to the `cron` input. It will then automatically commit an updated timestamp file every 10 seconds to the `cron` input repository, which will automatically trigger our pipeline.

![alt tag](cron1.png)

The pipeline will run every 10 seconds, querying our database and updating its output.

We have used the `@every 10s` cron spec here, but you can use any cron spec formatted according to [RFC 3339](https://www.ietf.org/rfc/rfc3339.txt). For example, `*/10 * * * *` would indicate that the pipeline should run every 10 minutes (these time formats should be familiar to those who have used cron in the past, and you can find more examples [here](https://en.wikipedia.org/wiki/Cron))
