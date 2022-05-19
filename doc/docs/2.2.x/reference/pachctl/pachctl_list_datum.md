## pachctl list datum

Return the datums in a job.

### Synopsis

Return the datums in a job.

```
pachctl list datum <pipeline>@<job> [flags]
```

### Options

```
  -f, --file string     The JSON file containing the pipeline to list datums from, the pipeline need not exist
  -h, --help            help for datum
  -o, --output string   Output format when --raw is set: "json" or "yaml" (default "json")
      --raw             Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

