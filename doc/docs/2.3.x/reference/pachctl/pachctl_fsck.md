## pachctl fsck

Run a file system consistency check on pfs.

### Synopsis

Run a file system consistency check on the pachyderm file system, ensuring the correct provenance relationships are satisfied.

```
pachctl fsck [flags]
```

### Options

```
  -f, --fix             Attempt to fix as many issues as possible.
  -h, --help            help for fsck
      --zombie string   A single commit to check for zombie files
      --zombie-all      Check all pipelines for zombie files: files corresponding to old inputs that were not properly deleted
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

