## pachctl logs

Return logs from a job.

### Synopsis

Return logs from a job.

```
pachctl logs [--pipeline=<pipeline>|--job=<job>] [--datum=<datum>] [flags]
```

### Examples

```

# Return logs emitted by recent jobs in the "filter" pipeline
$ pachctl logs --pipeline=filter

# Return logs emitted by the job aedfa12aedf
$ pachctl logs --job=aedfa12aedf

# Return logs emitted by the pipeline \"filter\" while processing /apple.txt and a file with the hash 123aef
$ pachctl logs --pipeline=filter --inputs=/apple.txt,123aef
```

### Options

```
      --datum string      Filter for log lines for this datum (accepts datum ID)
  -f, --follow            Follow logs as more are created.
  -h, --help              help for logs
      --inputs string     Filter for log lines generated while processing these files (accepts PFS paths or file hashes)
  -j, --job string        Filter for log lines from this job (accepts job ID)
      --master            Return log messages from the master process (pipeline must be set).
  -p, --pipeline string   Filter the log for lines from this pipeline (accepts pipeline name)
      --raw               Return log messages verbatim from server.
  -t, --tail int          Lines of recent logs to display.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

