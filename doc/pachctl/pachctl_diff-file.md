## ./pachctl diff-file

Return a diff of two file trees.

### Synopsis


Return a diff of two file trees.

Examples:

```sh

# Return the diff between foo master path and its parent.
$ pachctl diff-file foo master path

# Return the diff between foo master path1 and bar master path2.
$ pachctl diff-file foo master path1 bar master path2

```

```
./pachctl diff-file new-repo-name new-commit-id new-path [old-repo-name old-commit-id old-path]
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -s, --shallow           Specifies whether or not to diff subdirectories
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

