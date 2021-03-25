## pachctl port-forward

Forward a port on the local machine to pachd. This command blocks.

### Synopsis

Forward a port on the local machine to pachd. This command blocks.

```
pachctl port-forward [flags]
```

### Options

```
  -h, --help                    help for port-forward
      --namespace string        Kubernetes namespace Pachyderm is deployed in.
  -f, --pfs-port uint16         The local port to bind PFS over HTTP to. (default 30652)
  -p, --port uint16             The local port to bind pachd to. (default 30650)
  -x, --proxy-port uint16       The local port to bind Pachyderm's dash proxy service to. (default 30081)
      --remote-port uint16      The remote port that pachd is bound to in the cluster. (default 650)
  -s, --s3gateway-port uint16   The local port to bind the s3gateway to. (default 30600)
      --saml-port uint16        The local port to bind pachd's SAML ACS to. (default 30654)
  -u, --ui-port uint16          The local port to bind Pachyderm's dash service to. (default 30080)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

