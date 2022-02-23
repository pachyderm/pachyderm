## pachctl debug binary

Collect a set of binaries.

### Synopsis

Collect a set of binaries.

```
pachctl debug binary <file> [flags]
```

### Options

```
  -h, --help              help for binary
      --pachd             Only collect the binary from pachd.
  -p, --pipeline string   Only collect the binary from the worker pods for the given pipeline.
  -w, --worker string     Only collect the binary from the given worker pod.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

