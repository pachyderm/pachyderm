## pachctl run pps-load-test

Run a PPS load test.

### Synopsis

Run a PPS load test.

```
pachctl run pps-load-test <spec-file>  [flags]
```

### Options

```
  -d, --dag string         The DAG specification file to use for the load test
  -h, --help               help for pps-load-test
  -p, --parallelism int    The parallelism to use for the pipelines.
      --pod-patch string   The pod patch file to use for the pipelines.
  -s, --seed int           The seed to use for generating the load.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

