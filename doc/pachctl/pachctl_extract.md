## pachctl extract

Extract Pachyderm state to stdout or an object store bucket.

### Synopsis


Extract Pachyderm state to stdout or an object store bucket.

```
pachctl extract
```

### Examples

```
```sh

# Extract into a local file:
$ pachctl extract > backup

# Extract to s3:
$ pachctl extract -u s3://bucket/backup
```
```

### Options

```
      --no-objects   don't extract from object storage, only extract data from etcd
  -u, --url string   An object storage url (i.e. s3://...) to extract to.
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

