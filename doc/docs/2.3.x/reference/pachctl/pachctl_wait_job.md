## pachctl wait job

Wait for a job to finish then return info about the job.

### Synopsis

Wait for a job to finish then return info about the job.

```
pachctl wait job <job>|<pipeline>@<job> [flags]
```

### Options

```
      --full-timestamps   Return absolute timestamps (as opposed to the default, relative timestamps).
  -h, --help              help for job
  -o, --output string     Output format when --raw is set: "json" or "yaml" (default "json")
      --raw               Disable pretty printing; serialize data structures to an encoding such as json or yaml
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

