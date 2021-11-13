## pachctl squash commit

Squash the sub-commits of a commit.

### Synopsis

Squash the sub-commits of a commit.  The data in the sub-commits will remain in their child commits.
The squash will fail if it includes a commit with no children

```
pachctl squash commit <commit-id> [flags]
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

