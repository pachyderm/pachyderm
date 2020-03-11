## pachctl delete pipeline

Delete a pipeline.

### Synopsis

Delete a pipeline.

```
pachctl delete pipeline (<pipeline>|--all) [flags]
```

### Options

```
      --all         delete all pipelines
  -f, --force       delete the pipeline regardless of errors; use with care
  -h, --help        help for pipeline
      --keep-repo   delete the pipeline, but keep the output repo around (the pipeline can be recreated later and use the same repo)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

