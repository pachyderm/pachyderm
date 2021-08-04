## pachctl auth deactivate

Delete all ACLs, tokens, admins, IDP integrations and OIDC clients, and deactivate Pachyderm auth

### Synopsis

Deactivate Pachyderm's auth and identity systems, which will delete ALL auth tokens, ACLs and admins, IDP integrations and OIDC clients, and expose all data in the cluster to any user with cluster access. Use with caution.

```
pachctl auth deactivate [flags]
```

### Options

```
      --enterprise   Deactivate auth on the active enterprise context
  -h, --help         help for deactivate
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

