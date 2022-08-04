## pachctl port-forward

Forward a port on the local machine to pachd. This command blocks.

### Synopsis

Forward a port on the local machine to pachd. This command blocks.

```
pachctl port-forward [flags]
```

### Options

```
      --console-port uint16            The local port to bind the console service to. (default 4000)
      --dex-port uint16                The local port to bind the identity service to. (default 30658)
  -h, --help                           help for port-forward
      --namespace string               Kubernetes namespace Pachyderm is deployed in.
      --oidc-port uint16               The local port to bind pachd's OIDC callback to. (default 30657)
  -p, --port uint16                    The local port to bind pachd to. (default 30650)
      --remote-console-port uint16     The remote port to bind the console  service to. (default 4000)
      --remote-dex-port uint16         The local port to bind the identity service to. (default 1658)
      --remote-oidc-port uint16        The remote port that OIDC callback is bound to in the cluster. (default 1657)
      --remote-port uint16             The remote port that pachd is bound to in the cluster. (default 1650)
      --remote-s3gateway-port uint16   The remote port that the s3 gateway is bound to. (default 1600)
  -s, --s3gateway-port uint16          The local port to bind the s3gateway to. (default 30600)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

