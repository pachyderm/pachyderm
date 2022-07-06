## pachctl flush job

Wait for all jobs caused by the specified commits to finish and return them.

### Synopsis

Wait for all jobs caused by the specified commits to finish and return them.

```
pachctl flush job <repo>@<branch-or-commit> ... [flags]
```

### Examples

```

# Return jobs caused by foo@XXX and bar@YYY.
pachctl flush job foo@XXX bar@YYY

# Return jobs caused by foo@XXX leading to pipelines bar and baz.
pachctl flush job foo@XXX -p bar -p baz
```

### Options

```
      --full-timestamps     Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help                help for job
  -o, --output string       Output format when --raw is set: "json" or "yaml" (default "json")
  -p, --pipeline []string   Wait only for jobs leading to a specific set of pipelines (default [])
      --raw                 Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

