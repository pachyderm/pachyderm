## pachctl auth deactivate

Delete all ACLs, tokens, and admins, and deactivate Pachyderm auth

### Synopsis


Deactivate Pachyderm's auth system, which will delete ALL auth tokens, ACLs and admins, and expose all data in the cluster to any user with cluster access. Use with caution.

```
pachctl auth deactivate
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

