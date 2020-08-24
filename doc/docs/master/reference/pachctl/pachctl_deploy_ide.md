## pachctl deploy ide

Deploy the Pachyderm IDE.

### Synopsis

Deploy a JupyterHub-based IDE alongside the Pachyderm cluster.

```
pachctl deploy ide [flags]
```

### Options

```
      --dry-run               Don't actually deploy, instead just print the Helm config.
  -h, --help                  help for ide
      --hub-image string      Image for IDE hub (default "pachyderm/ide-hub:1.0.0")
      --lb-tls-email string   Contact email for minting a Let's Encrypt TLS cert on the load balancer
      --lb-tls-host string    Hostname for minting a Let's Encrypt TLS cert on the load balancer
  -o, --output string         Output format. One of: json|yaml (default "json")
      --user-image string     Image for IDE user environments (default "pachyderm/ide-user:1.0.0")
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

