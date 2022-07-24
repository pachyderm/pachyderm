## pachctl flush commit

Wait for all commits caused by the specified commits to finish and return them.

### Synopsis

Wait for all commits caused by the specified commits to finish and return them.

```
pachctl flush commit <repo>@<branch-or-commit> ... [flags]
```

### Examples

```

# return commits caused by foo@XXX and bar@YYY
pachctl flush commit foo@XXX bar@YYY

# return commits caused by foo@XXX leading to repos bar and baz
pachctl flush commit foo@XXX -r bar -r baz
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for commit
      --raw               disable pretty printing, print raw json
  -r, --repos []string    Wait only for commits leading to a specific set of repos (default [])
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

