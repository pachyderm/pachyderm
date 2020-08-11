
# The build flag

The `--build` flag is another way to improve development speed when working with pipelines. While [build pipelines](build-pipelines.md) avoid rebuilding the docker image, the `--build` flag builds/re-builds, tags and pushes the new Docker image. This feature can be particularly useful while iterating on the Docker image itself, as it can be difficult to keep up with changing image tags and ensure the image is pushed before updating the pipeline (Steps 2-5 in the [pipeline workflow](working-with-pipelines.md)).

The `--build` flag performs the following steps:

1. Builds the Docker image specified in the pipeline
1. Gives the images a unique tag
1. Pushes the Docker image to the registry
1. Updates the image tag in the pipeline
1. Submits the updated pipeline Pachyderm cluster.

The usage of the flag is shown below:

   ```bash
   pachctl update pipeline -f <pipeline name> --build --registry <registry> --username <registry user>
   ```

!!! note
      For more details on the `--build` flag, see [Update a Pipeline](../../updating_pipelines/#update-the-code-in-a-pipeline).