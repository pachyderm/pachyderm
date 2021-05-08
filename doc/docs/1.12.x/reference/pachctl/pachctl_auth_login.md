## pachctl auth login

Log in to Pachyderm

### Synopsis

Login to Pachyderm. Any resources that have been restricted to the account you have with your ID provider (e.g. GitHub, Okta) account will subsequently be accessible.

```
pachctl auth login [flags]
```

### Options

```
  -h, --help                help for login
  -o, --one-time-password   If set, authenticate with a Dash-provided One-Time Password, rather than via GitHub
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

