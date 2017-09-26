# Triggering Pipelines Periodically (cron)

Pachyderm pipelines are triggered by changes to their input data repositories (as further discussed in [What Happens When You Create a Pipeline](../getting_started/beginner_tutorial.html#what-happens-when-you-create-a-pipeline)). However, there are many scenarios in which you might want to trigger pipelines periodically and/or have pipelines actively pull in data from outside sources.  For example, you might need to periodically: 

- Scrap one or more websites
- Make a series of API calls 
- Query a database
- Retrieve the latest version of a file from S3 or FTP

In cases like these, you can utilize Pachyderm's built in `cron` input to periodically trigger these sorts of actions, and, thus, periodically drive pipelines based on databases queries, scraped data, etc. 

- [Non-incremental `cron` input](#non-incremental-cron) - For when you want to update (overwrite) a single set of results periodically.
- [Incremental `cron` input](#incremental-cron) - For when you want to get a latest set of results periodically and store that set of results along with previous results.

A full example in which MongoDB is queried periodically can be found [here](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/cron).

## Non-Incremental Cron

Let's say that we want to query a database every 10 seconds for a result, and update our result every time the pipeline is triggered. We could do this with a non-incremental `cron` input to the pipeline, as follows:

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

The pipeline will run every 10 seconds, query our database and update the output.  

We have utilized the `@every 10s` `cron` scheduling format here, but you can, more generally, use any time formatted according to [RFC 3339](https://www.ietf.org/rfc/rfc3339.txt).  For example, `*/10 * * * *` would indicate that the pipeline should run every 10 minutes. These time formats should be familiar to those who have used cron in the past, and you can find more examples [here](https://en.wikipedia.org/wiki/Cron).

## Incremental Cron

In the [above example](#non-incremental-cron), Pachyderm will overwrite the output data from our `cron` triggered pipeline each time it runs. This happens because Pachyderm is updating the same input datum (the timestamp) after every period (see [our incremental processing docs](../fundamentals/incrementality.html) for more information on datums and incrementality). 

If we don't want our previous results to be necessarily replaced during every run, we should enable incrementality in the pipeline specification:

```
{

  ...

  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 10s"
    }  
  },
  "incremental": true
}
``` 

When we do this, Pachyderm won't update the same timestamp in the `cron` data repository, and, thus, we can accumulate results periodically over time:

![alt tag](cron2.png)

Note, even with `"incremental": true` you can still overwrite data in the output data repository (e.g., by replacing a file with a new file having the same name). The point is that the whole output wouldn't necessarily be replaced when you are running with `"incremental": true`.
