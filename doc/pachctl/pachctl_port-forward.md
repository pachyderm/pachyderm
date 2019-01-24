## ./pachctl port-forward

Forward a port on the local machine to pachd. This command blocks.

### Synopsis


Forward a port on the local machine to pachd. This command blocks.

```
./pachctl port-forward
```

### Options

```
      --namespace string   Kubernetes namespace Pachyderm is deployed in. (default "default")
  -f, --pfs-port int       The local port to bind PFS over HTTP to. (default 30652)
  -p, --port int           The local port to bind pachd to. (default 30650)
  -x, --proxy-port int     The local port to bind Pachyderm's dash proxy service to. (default 30081)
      --saml-port int      The local port to bind pachd's SAML ACS to. (default 30654)
  -u, --ui-port int        The local port to bind Pachyderm's dash service to. (default 30080)
```

### Options inherited from parent commands

```
      --no-metrics   Don't report user metrics for this command
  -v, --verbose      Output verbose logs
```

