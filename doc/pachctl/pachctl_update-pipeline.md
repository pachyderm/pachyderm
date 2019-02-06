## ./pachctl update-pipeline

Update an existing Pachyderm pipeline.

### Synopsis


Update a Pachyderm pipeline with a new [Pipeline Specification](../reference/pipeline_spec.html)

```
./pachctl update-pipeline -f pipeline.json
```

### Options

```
  -b, --build             If true, build and push local docker images into the docker registry.
  -f, --file string       The file containing the pipeline, it can be a url or local file. - reads from stdin. (default "-")
  -p, --push-images       If true, push local docker images into the docker registry.
  -r, --registry string   The registry to push images to. (default "docker.io")
      --reprocess         If true, reprocess datums that were already processed by previous version of the pipeline.
  -u, --username string   The username to push images as, defaults to your OS username.
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

