## pachctl wait commit

Wait for the specified commit to finish and return it.

### Synopsis

Wait for the specified commit to finish and return it.

```
pachctl wait commit <repo>@<branch-or-commit> [flags]
```

### Examples

```

# wait for the commit foo@XXX to finish and return it
pachctl wait commit foo@XXX -b bar@baz
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for commit
  -o, --output string     Output format when --raw is set: "json" or "yaml" (default "json")
      --raw               Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

