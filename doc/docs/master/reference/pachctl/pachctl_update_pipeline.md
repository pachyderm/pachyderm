## pachctl update pipeline

Update an existing Pachyderm pipeline.

### Synopsis

Update a Pachyderm pipeline with a new pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline-spec.html.

```
pachctl update pipeline [flags]
```

### Options

```
  -f, --file string       The JSON file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
  -h, --help              help for pipeline
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

