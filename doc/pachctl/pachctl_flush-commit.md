## ./pachctl flush-commit

Wait for all commits caused by the specified commits to finish and return them.

### Synopsis


Wait for all commits caused by the specified commits to finish and return them.

Examples:

```sh

# return commits caused by foo/XXX and bar/YYY
$ pachctl flush-commit foo/XXX bar/YYY

# return commits caused by foo/XXX leading to repos bar and baz
$ pachctl flush-commit foo/XXX -r bar -r baz

```

```
./pachctl flush-commit commit [commit ...]
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
      --raw               disable pretty printing, print raw json
  -r, --repos []string    Wait only for commits leading to a specific set of repos (default [])
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

