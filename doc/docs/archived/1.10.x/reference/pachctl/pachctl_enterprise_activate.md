## pachctl enterprise activate

Activate the enterprise features of Pachyderm with an activation code

### Synopsis

Activate the enterprise features of Pachyderm with an activation code

```
echo <your-activation-token> | pachctl enterprise activate
```

### Options

```
      --expires string   A timestamp indicating when the token provided above should expire (formatted as an RFC 3339/ISO 8601 datetime). This is only applied if it's earlier than the signed expiration time encoded in 'activation-code', and therefore is only useful for testing.
  -h, --help             help for activate
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

