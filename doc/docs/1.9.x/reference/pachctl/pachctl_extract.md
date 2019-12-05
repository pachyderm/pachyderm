## pachctl extract

Extract Pachyderm state to stdout or an object store bucket.

### Synopsis

Extract Pachyderm state to stdout or an object store bucket.

```
pachctl extract [flags]
```

### Examples

```

# Extract into a local file:
$ pachctl extract > backup

# Extract to s3:
$ pachctl extract -u s3://bucket/backup
```

### Options

```
  -h, --help         help for extract
      --no-objects   don't extract from object storage, only extract data from etcd
  -u, --url string   An object storage url (i.e. s3://...) to extract to.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

