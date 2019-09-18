## pachctl create pipeline

Create a new pipeline.

### Synopsis


Create a new pipeline from a pipeline specification. For details on the format, see http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html.

```
pachctl create pipeline
```

### Options

```
  -b, --build             If true, build and push local docker images into the docker registry.
  -f, --file string       The JSON file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
  -p, --push-images       If true, push local docker images into the docker registry.
  -r, --registry string   The registry to push images to. (default "docker.io")
  -u, --username string   The username to push images as, defaults to your docker username.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

