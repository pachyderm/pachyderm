# Updating Pipelines

During development, it's very common to update pipelines, whether it's changing
your code or just cranking up parallelism.  For example, when developing a
machine learning model you will likely need to try out a bunch of different
versions of your model while your training data stays relatively constant.
This is where `update-pipeline` comes in.

## Updating your pipeline specification

In cases in which you are updating parallelism, adding another input repo, or
otherwise modifying your [pipeline
specification](../reference/pipeline_spec.html), you just need to update your
JSON file and call `update-pipeline`:

```sh
$ pachctl update-pipeline -f pipeline.json
```

Similar to `create-pipeline`, `update-pipeline` with the `-f` flag can also
take a URL if your JSON manifest is hosted on GitHub or elsewhere.

## Updating the code used in a pipeline

You can also use `update-pipeline` to update the code you are using in one or
more of your pipelines.  To update the code in your pipeline:

1. Make the code changes.
2. Build, tag, and push the image in docker to the place specified in the pipeline spec.
3. Call `pachctl update-pipeline` again.

## Building pipeline images within pachyderm

Building, tagging and pushing the image in docker requires a bit of ceremony,
so there's a shortcut: the `--build` flag for `pachctl update-pipeline`. When
used, Pachyderm will do the following:

1. Rebuild the docker image.
2. Tag your image with a new unique name.
3. Push that tagged image to your registry (e.g., DockerHub).
4. Update the pipeline specification that you previously gave to Pachyderm to
   use the new unique tag.

For example, you could update the Python code used in the [OpenCV
pipeline](../getting_started/beginner_tutorial.html) via:

```sh
pachctl update-pipeline -f edges.json --build --username <registry user>
```

You'll then be prompted for the password associated with the registry user.

### Private registries

`--build` supports private registries as well. Make sure the private registry
is specified as part of the pipeline spec, and use the `--registry` flag when
calling `pachctl update-pipeline --build`.

For example, if you wanted to push the image `pachyderm/opencv` to a registry
located at `localhost:5000`, you'd have this in your pipeline spec:

```
"image": "localhost:5000/pachyderm/opencv"
```

And would run this to update the pipeline:

```sh
pachctl update-pipeline -f edges.json --build --registry localhost:5000 --username <registry user>
```

## Re-processing data

As of 1.5.1, updating a pipeline will NOT reprocess previously
processed data by default. New data that's committed to the inputs will be processed with
the new code and "mixed" with the results of processing data with the previous
code. Furthermore, data that Pachyderm tried and failed to process with the
previous code due to code erroring will be processed with the new code.

`update-pipeline` (without flags) is designed for the situation where your code needs to be
fixed because it encountered an unexpected new form of data.

If you'd like to update your pipeline and have that updated pipeline reprocess all the data 
that is currently in the HEAD commit of your input repos, you
should use the `--reprocess` flag. This type of update will automatically trigger a job that reprocesses all of the input data in its current state (i.e., the HEAD commits)
with the updated pipeline. Then from that point on, the updated pipeline will continue to be used to process any new input data. Previous results will still be
available in via their corresponding commit IDs.
