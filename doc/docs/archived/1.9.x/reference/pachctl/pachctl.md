## pachctl



### Synopsis

Access the Pachyderm API.

Environment variables:
  PACH_CONFIG=<path>, the path where pachctl will attempt to load your pach config.
  JAEGER_ENDPOINT=<host>:<port>, the Jaeger server to connect to, if PACH_TRACE is set
  PACH_TRACE={true,false}, If true, and JAEGER_ENDPOINT is set, attach a
    Jaeger trace to any outgoing RPCs


### Options

```
  -h, --help       help for pachctl
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

