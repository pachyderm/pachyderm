# The Build Flag

The `--build` flag is one way to improve development speed when working with pipelines. Unlike [build pipelines](build-pipelines.md), this method still uses docker images. This feature is particularly useful if you want to continue to work with docker images, e.g. because your team is accustomed to them, or because you need the added flexibility.

The `--build` flag performs the following steps:

1. Builds the Docker image specified in the pipeline
1. Gives the images a unique tag
1. Pushes the Docker image to the registry
1. Updates the image tag in the pipeline spec json to match the new image
1. Submits the updated pipeline to the Pachyderm cluster

The usage of the flag is shown below:

   ```shell
   pachctl update pipeline -f <pipeline name> --build --registry <registry> --username <registry user>
   ```

!!! note
      For more details on the `--build` flag, see [Update a Pipeline](../../pipeline-operations/updating_pipelines/#update-the-code-in-a-pipeline).
