## pachctl



### Synopsis

Access the Pachyderm API.

Environment variables:
  PACH_CONFIG=<path>, the path where pachctl will attempt to load your config.
  JAEGER_ENDPOINT=<host>:<port>, the Jaeger server to connect to, if PACH_TRACE
    is set
  PACH_TRACE={true,false}, if true, and JAEGER_ENDPOINT is set, attach a Jaeger
    trace to any outgoing RPCs.
  PACH_TRACE_DURATION=<duration>, the amount of time for which PPS should trace
    a pipeline after 'pachctl create-pipeline' (PACH_TRACE must also be set).


### Options

```
  -h, --help       help for pachctl
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

