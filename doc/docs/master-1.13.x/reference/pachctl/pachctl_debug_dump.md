## pachctl debug dump

Collect a standard set of debugging information.

### Synopsis

Collect a standard set of debugging information.

```
pachctl debug dump <file> [flags]
```

### Options

```
  -h, --help              help for dump
  -l, --limit int         Limit sets the limit for the number of commits / jobs that are returned for each repo / pipeline in the dump.
      --pachd             Only collect the dump from pachd.
  -p, --pipeline string   Only collect the dump from the worker pods for the given pipeline.
  -w, --worker string     Only collect the dump from the given worker pod.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

