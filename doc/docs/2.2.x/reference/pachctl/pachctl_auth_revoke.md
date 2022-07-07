## pachctl auth revoke

Revoke a Pachyderm auth token

### Synopsis

Revoke a Pachyderm auth token.

```
pachctl auth revoke [flags]
```

### Options

```
      --enterprise     Revoke an auth token (or all auth tokens minted for one user) on the enterprise server
  -h, --help           help for revoke
      --token string   Pachyderm auth token that should be revoked (one of --token or --user must be set)
      --user string    User whose Pachyderm auth tokens should be revoked (one of --token or --user must be set)
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

