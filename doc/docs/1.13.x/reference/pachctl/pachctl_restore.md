## pachctl restore

Restore Pachyderm state from stdin or an object store.

### Synopsis

Restore Pachyderm state from stdin or an object store.

```
pachctl restore [flags]
```

### Examples

```

# Restore from a local file:
pachctl restore < backup

# Restore from s3:
pachctl restore -u s3://bucket/backup
```

### Options

```
  -h, --help         help for restore
      --no-token     Don't generate a new auth token at the beginning of the restore.
  -u, --url string   An object storage url (i.e. s3://...) to restore from.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

