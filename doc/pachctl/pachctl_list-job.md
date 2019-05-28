## ./pachctl list-job

Return info about jobs.

### Synopsis


Return info about jobs.

Examples:

```sh# return all jobs
$ pachctl list-job

# return all jobs in pipeline foo
$ pachctl list-job -p foo

# return all jobs whose input commits include foo/XXX and bar/YYY
$ pachctl list-job foo/XXX bar/YYY

# return all jobs in pipeline foo and whose input commits include bar/YYY
$ pachctl list-job -p foo bar/YYY
```

```
./pachctl list-job [commits]
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -i, --input strings     List jobs with a specific set of input commits.
  -o, --output string     List jobs with a specific output commit.
  -p, --pipeline string   Limit to jobs made by pipeline.
      --raw               disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

