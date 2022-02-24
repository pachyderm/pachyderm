## pachctl delete commit

Delete the sub-commits of a commit.

### Synopsis

Delete the sub-commits of a commit.  The data in the sub-commits will be lost.
This operation is only supported if none of the sub-commits have children.

```
pachctl delete commit <commit-id> [flags]
```

### Options

```
  -h, --help   help for commit
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

