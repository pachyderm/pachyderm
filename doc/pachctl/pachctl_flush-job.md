## ./pachctl flush-job

Wait for all jobs caused by the specified commits to finish and return them.

### Synopsis


Wait for all jobs caused by the specified commits to finish and return them.

Examples:

```sh# return jobs caused by foo/XXX and bar/YYY
$ pachctl flush-job foo/XXX bar/YYY

# return jobs caused by foo/XXX leading to pipelines bar and baz
$ pachctl flush-job foo/XXX -p bar -p baz
```

```
./pachctl flush-job commit [commit ...]
```

### Options

```
      --full-timestamps     Return absolute timestamps (as opposed to the default, relative timestamps).
  -p, --pipeline []string   Wait only for jobs leading to a specific set of pipelines (default [])
      --raw                 disable pretty printing, print raw json
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

