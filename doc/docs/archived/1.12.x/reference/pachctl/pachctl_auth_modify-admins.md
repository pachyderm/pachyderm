## pachctl auth modify-admins

Modify the current cluster admins

### Synopsis

Modify the current cluster admins. --add accepts a comma-separated list of users to grant admin status, and --remove accepts a comma-separated list of users to revoke admin status

```
pachctl auth modify-admins [flags]
```

### Options

```
      --add strings      Comma-separated list of users to grant admin status
  -h, --help             help for modify-admins
      --remove strings   Comma-separated list of users revoke admin status
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

