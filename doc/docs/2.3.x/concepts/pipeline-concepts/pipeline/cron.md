# Cron Pipeline

A Cron pipeline is triggered by a set time interval *instead* of whenever new changes appear in the input repository.

## About Cron Pipelines

###  Use Cases

Cron pipelines are great for tasks like:

- Scraping websites
- Making API calls
- Querying a database
- Retrieving a file from a location accessible through an S3 protocol or a File Transfer Protocol (FTP).

### Behavior 

When you create a Cron pipeline, Pachyderm creates a new input data repository that corresponds to the `cron` input Then, Pachyderm automatically commits a timestamp file to the `cron` input repository at your determined interval, which triggers the pipeline.  

By default, each cron trigger adds a new tick file to the cron input repository, accumulating more datums over time. 
Optionally, you can set the overwrite flag to `true` to overwrite the timestamp file on each tick. To learn more about overwriting commits in Pachyderm, see [Datum processing](../datum/index.md).

### Required Parameters 

At minimum, a Cron pipeline must include all of the following parameters:

| Parameters  | Description  |
| ---------- | ------------ |
| `"name"`   | A descriptive name of the cron pipeline. |
| `"spec"`   | The interval between scheduled cron jobs; accepts [RFC 3339](https://www.ietf.org/rfc/rfc3339.txt) inputs, [Predefined Schedules](https://pkg.go.dev/github.com/robfig/cron#hdr-Predefined_schedules) (`@daily`), and [Intervals](https://pkg.go.dev/github.com/robfig/cron#hdr-Intervals) (`@every 1h30m20s`)|


### Callouts

!!! Warning 
    Avoid using intervals faster than 1-5 minutes

- You can use `never` during development and manually trigger the pipeline
- If using jsonnet, you can pass arguments like:  `--arg cronSpec="@every 5m"`

---

## Examples

### Every 60 Seconds

```json
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@every 60s"
    }
  }
```

### Daily with Overwrites

```json
  "input": {
    "cron": {
      "name": "tick",
      "spec": "@daily",
      "overwrite": true
    }
  }
```

### SQL Ingest with Jsonnet

```json
pachctl update pipeline --jsonnet https://raw.githubusercontent.com/pachyderm/pachyderm/2.3.x/src/templates/sql_ingest_cron.jsonnet \
  --arg name=myingest \
  --arg url="mysql://root@mysql:3306/test_db" \
  --arg query="SELECT * FROM test_data" \
  --arg hasHeader=false \
  --arg cronSpec="@every 60s" \
  --arg secretName="mysql-creds" \
  --arg format=json 
```

!!! note "See Also:"
    [Periodic Ingress from MongoDB](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/db){target=_blank}
