## pachctl auth get-auth-token

Get an auth token that authenticates the holder as "username", or the currently signed-in user, if no 'username' is provided

### Synopsis

Get an auth token that authenticates the holder as "username"; or the currently signed-in user, if no 'username' is provided. Only cluster admins can obtain an auth token on behalf of another user.

```
pachctl auth get-auth-token [username] [flags]
```

### Options

```
  -h, --help         help for get-auth-token
  -q, --quiet        if set, only print the resulting token (if successful). This is useful for scripting, as the output can be piped to use-auth-token
      --ttl string   if set, the resulting auth token will have the given lifetime (or the lifetimeof the caller's current session, whichever is shorter). This flag should be a golang duration (e.g. "30s" or "1h2m3s"). If unset, tokens will have a lifetime of 30 days.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

