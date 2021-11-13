## pachctl auth deactivate

Delete all ACLs, tokens, and admins, and deactivate Pachyderm auth

### Synopsis

Deactivate Pachyderm's auth system, which will delete ALL auth tokens, ACLs and admins, and expose all data in the cluster to any user with cluster access. Use with caution.

```
pachctl auth deactivate [flags]
```

### Options

```
  -h, --help   help for deactivate
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

