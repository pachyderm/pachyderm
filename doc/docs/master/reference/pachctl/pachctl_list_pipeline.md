## pachctl list pipeline

Return info about all pipelines.

### Synopsis

Return info about all pipelines.

```
pachctl list pipeline [<pipeline>] [flags]
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for pipeline
      --history string    Return revision history for pipelines. (default "none")
  -o, --output string     Output format when --raw is set ("json" or "yaml")
      --raw               disable pretty printing, print raw json
  -s, --spec              Output 'create pipeline' compatibility specs.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

