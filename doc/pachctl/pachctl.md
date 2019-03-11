## ./pachctl



### Synopsis


Access the Pachyderm API.

Environment variables:
  PACHD_ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).
  PACH_CONFIG=<path>, the path where pachctl will attempt to load your pach config.
  JAEGER_ENDPOINT=<host>:<port>, the Jaeger server to connect to, if PACH_ENABLE_TRACING is set
  PACH_ENABLE_TRACING={true,false}, If true, and JAEGER_ENDPOINT is set, attach a
    Jaeger trace to all outgoing RPCs


### Options

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

