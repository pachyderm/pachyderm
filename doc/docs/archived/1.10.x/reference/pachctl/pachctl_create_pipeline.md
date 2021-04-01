## pachctl create pipeline

Create a new pipeline.

### Synopsis

Create a new pipeline from a pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.

```
pachctl create pipeline [flags]
```

### Options

```
  -b, --build             If true, build and push local docker images into the docker registry.
  -f, --file string       The JSON file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
  -h, --help              help for pipeline
  -p, --push-images       If true, push local docker images into the docker registry.
  -r, --registry string   The registry to push images to. (default "index.docker.io")
  -u, --username string   The username to push images as.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

