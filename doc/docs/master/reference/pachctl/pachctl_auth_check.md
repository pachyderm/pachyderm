## pachctl auth check

Check whether you have reader/writer/etc-level access to 'repo'

### Synopsis

Check whether you have reader/writer/etc-level access to 'repo'. For example, 'pachctl auth check reader private-data' prints "true" if the you have at least "reader" access to the repo "private-data" (you could be a reader, writer, or owner). Unlike `pachctl auth get`, you do not need to have access to 'repo' to discover your own access level.

```
pachctl auth check (none|reader|writer|owner) <repo> [flags]
```

### Options

```
  -h, --help   help for check
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

