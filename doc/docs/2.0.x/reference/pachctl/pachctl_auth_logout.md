## pachctl auth logout

Log out of Pachyderm by deleting your local credential

### Synopsis

Log out of Pachyderm by deleting your local credential. Note that it's not necessary to log out before logging in with another account (simply run 'pachctl auth login' twice) but 'logout' can be useful on shared workstations.

```
pachctl auth logout [flags]
```

### Options

```
      --enterprise   Log out of the active enterprise context
  -h, --help         help for logout
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

