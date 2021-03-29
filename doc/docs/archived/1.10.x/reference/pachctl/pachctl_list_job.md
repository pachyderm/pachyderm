## pachctl list job

Return info about jobs.

### Synopsis

Return info about jobs.

```
pachctl list job [flags]
```

### Examples

```

# Return all jobs
$ pachctl list job

# Return all jobs from the most recent version of pipeline "foo"
$ pachctl list job -p foo

# Return all jobs from all versions of pipeline "foo"
$ pachctl list job -p foo --history all

# Return all jobs whose input commits include foo@XXX and bar@YYY
$ pachctl list job -i foo@XXX -i bar@YYY

# Return all jobs in pipeline foo and whose input commits include bar@YYY
$ pachctl list job -p foo -i bar@YYY
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for job
      --history string    Return jobs from historical versions of pipelines. (default "none")
  -i, --input strings     List jobs with a specific set of input commits. format: <repo>@<branch-or-commit>
      --no-pager          Don't pipe output into a pager (i.e. less).
  -o, --output string     List jobs with a specific output commit. format: <repo>@<branch-or-commit>
  -p, --pipeline string   Limit to jobs made by pipeline.
      --raw               Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

