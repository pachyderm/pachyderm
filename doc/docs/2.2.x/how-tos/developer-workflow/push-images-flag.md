# The Push Images Flag

The `--push-images` flag is one way to improve development speed when working with pipelines. 

The `--push-images` flag performs the following steps after you have built your image:

1. In your local registry, generates a unique tag for the image named after the `transform.image` field of your pipeline spec. 

    !!! Important
        You must build your image with your username as a prefix  (example: `pachyderm/example-joins-inner-outer`) - This name  must match the one declared in the `transform.image` field of your pipeline spec. 

1. Pushes the Docker image, with the tag, to your registry 
1. Updates the image tag in the pipeline spec json (on the fly) to match the new image
1. Submits the updated pipeline to the Pachyderm cluster

The usage of the flag is shown below:

   ```shell
   pachctl update pipeline -f <pipeline name> --push-images --registry <registry> --username <registry user>
   ```

!!! note
      For more details on the `--push-images` flag, see [Update a Pipeline](../../pipeline-operations/updating-pipelines/#update-the-code-in-a-pipeline).
