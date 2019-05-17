## pachctl restore

Restore Pachyderm state from stdin or an object store.

### Synopsis


Restore Pachyderm state from stdin or an object store.

```
pachctl restore
```

### Examples

```
```sh

# Restore from a local file:
$ pachctl restore < backup

# Restore from s3:
$ pachctl restore -u s3://bucket/backup
```
```

### Options

```
  -u, --url string   An object storage url (i.e. s3://...) to restore from.
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

