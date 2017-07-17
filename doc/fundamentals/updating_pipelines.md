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
more of your piplines.  To update the code in your pipeline:

1. Make the code changes.
2. Re-build your Docker image.
3. Call `update-pipeline` with the `--push-images` flag.

You need to call `update-pipeline` with the `--push-images` flag because, if
you have already run your pipeline, Pachyderm has already pulled the specified
images.  It won't re-pull new versions of the images, unless we tell it to
(which ensures that we don't waste time pulling images when we don't need to).
When `--push-images` is specified, Pachyderm will do the following:

1. Tag your image with a new unique tag.
2. Push that tagged image to your registry (e.g., DockerHub).
3. Update the pipeline specification that you previously gave to Pachyderm with
   the new unique tag.

For example, you could update the Python code used in the [OpenCV
pipeline](../getting_started/beginner_tutorial.html) via:

```sh
pachctl update-pipeline -f edges.json --push-images --password <registry password> -u <registry user>
```

## Re-processing data

As of 1.5.1, updating a pipeline will NOT reprocess previously
processed data by default. New data that's committed to the inputs will be processed with
the new code and "mixed" with the results of processing data with the previous
code. Furthermore, data that Pachyderm tried and failed to process with the
previous code due to code erroring will be processed with the new code.

`update-pipeline` (without flags) is designed for the situation where your code needs to be
fixed because it encountered an unexpected new form of data.

If you'd like to update your pipeline and reprocess everything from scratch, you
should use the `--reprocess` flag. This will reprocess all previously processed
data and all new data with the new code. Previous results will still be
available in pfs.
