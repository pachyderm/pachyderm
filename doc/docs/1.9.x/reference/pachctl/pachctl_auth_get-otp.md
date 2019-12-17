## pachctl auth get-otp

Get a one-time password that authenticates the holder as "username", or the currently signed in user if no 'username' is specified

### Synopsis

Get a one-time password that authenticates the holder as "username", or the currently signed in user if no 'username' is specified. Only cluster admins may obtain a one-time password on behalf of another user.

```
pachctl auth get-otp <username> [flags]
```

### Options

```
  -h, --help         help for get-otp
      --ttl string   if set, the resulting one-time password will have the given lifetime (or the lifetime of the caller's current session, whichever is shorter). This flag should be a golang duration (e.g. "30s" or "1h2m3s"). If unset, one-time passwords will have a lifetime of 5 minutes
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

