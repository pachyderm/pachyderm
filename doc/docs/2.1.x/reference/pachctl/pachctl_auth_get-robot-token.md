## pachctl auth get-robot-token

Get an auth token for a robot user with the specified name.

### Synopsis

Get an auth token for a robot user with the specified name.

```
pachctl auth get-robot-token [username] [flags]
```

### Options

```
      --enterprise   Get a robot token for the enterprise context
  -h, --help         help for get-robot-token
  -q, --quiet        if set, only print the resulting token (if successful). This is useful for scripting, as the output can be piped to use-auth-token
      --ttl string   if set, the resulting auth token will have the given lifetime. If not set, the token does not expire. This flag should be a golang duration (e.g. "30s" or "1h2m3s").
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

