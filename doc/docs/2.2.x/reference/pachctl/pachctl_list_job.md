## pachctl list job

Return info about jobs.

### Synopsis

Return info about jobs.

```
pachctl list job [<job-id>] [flags]
```

### Examples

```

# Return a summary list of all jobs
pachctl list job

# Return all sub-jobs in a job
pachctl list job <job-id>

# Return all sub-jobs split across all pipelines
pachctl list job --expand

# Return only the sub-jobs from the most recent version of pipeline "foo"
pachctl list job -p foo

# Return all sub-jobs from all versions of pipeline "foo"
pachctl list job -p foo --history all

# Return all sub-jobs whose input commits include foo@XXX and bar@YYY
pachctl list job -i foo@XXX -i bar@YYY

# Return all sub-jobs in pipeline foo and whose input commits include bar@YYY
pachctl list job -p foo -i bar@YYY
```

### Options

```
  -x, --expand              Show one line for each sub-job and include more columns
      --full-timestamps     Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help                help for job
      --history string      Return jobs from historical versions of pipelines. (default "none")
  -i, --input strings       List jobs with a specific set of input commits. format: <repo>@<branch-or-commit>
      --no-pager            Don't pipe output into a pager (i.e. less).
  -o, --output string       Output format when --raw is set: "json" or "yaml" (default "json")
  -p, --pipeline string     Limit to jobs made by pipeline.
      --raw                 Disable pretty printing; serialize data structures to an encoding such as json or yaml
      --state stringArray   Return only sub-jobs with the specified state. Can be repeated to include multiple states
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

