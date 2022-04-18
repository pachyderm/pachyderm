## pachctl update pipeline

Update an existing Pachyderm pipeline.

### Synopsis

Update a Pachyderm pipeline with a new pipeline specification. For details on the format, see https://docs.pachyderm.com/latest/reference/pipeline_spec/.

```
pachctl update pipeline [flags]
```

### Options

```
      --arg stringArray   Top-level argument passed to the Jsonnet template in --jsonnet (which must be set if any --arg arguments are passed). Value must be of the form 'param=value'. For multiple args, --arg may be set more than once.
  -f, --file string       A JSON file (url or filepath) containing one or more pipelines. "-" reads from stdin (the default behavior). Exactly one of --file and --jsonnet must be set.
  -h, --help              help for pipeline
      --jsonnet string    BETA: A Jsonnet template file (url or filepath) for one or more pipelines. "-" reads from stdin. Exactly one of --file and --jsonnet must be set. Jsonnet templates must contain a top-level function; strings can be passed to this function with --arg (below)
  -p, --push-images       If true, push local docker images into the docker registry.
  -r, --registry string   The registry to push images to. (default "index.docker.io")
      --reprocess         If true, reprocess datums that were already processed by previous version of the pipeline.
  -u, --username string   The username to push images as.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

