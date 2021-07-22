# The Push Images Flag

The `--push-images` flag is one way to improve development speed when working with pipelines. 

The `--push-images` flag performs the following steps:

1. Gives the images a unique tag
1. Pushes the Docker image to the registry
1. Updates the image tag in the pipeline spec json to match the new image
1. Submits the updated pipeline to the Pachyderm cluster

The usage of the flag is shown below:

   ```shell
   pachctl update pipeline -f <pipeline name> --push-images --registry <registry> --username <registry user>
   ```

!!! note
      For more details on the `--push-images` flag, see [Update a Pipeline](../../pipeline-operations/updating_pipelines/#update-the-code-in-a-pipeline).
