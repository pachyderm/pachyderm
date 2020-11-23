# Update a Pipeline

While working with your data, you often need to modify an existing
pipeline with new transformation code or pipeline
parameters.
For example, when you develop a machine learning model, you might need
to try different versions of your model while your training data
stays relatively constant. To make changes to a pipeline, you need to use
the `pachctl update pipeline` command.

## Update Your Pipeline Specification

If you need to update your
[pipeline specification](../reference/pipeline_spec.md), such as change the
parallelism settings, add an input repository, or other, you need to update your
JSON file and then run the `pachctl update pipeline`command.
By default, when you update your code, the new pipeline specification
does not reprocess the data that has already been processed. Instead,
it processes only the new data that you submit to the input repo.
If you want to run the changes in your pipeline against the data in
the `HEAD` commit of your input repo, use the `--reprocess` flag.
After that, the updated pipeline continues to process new input data.
Previous results remain accessible through the corresponding commit IDs.

To update a pipeline specification, complete the following steps:

1. Make the changes in your pipeline specification JSON file.

1. Update the pipeline with the new configuration:

   ```shell
   $ pachctl update pipeline -f pipeline.json
   ```

Similar to `create pipeline`, `update pipeline` with the `-f` flag can also
take a URL if your JSON manifest is hosted on GitHub or other
remote location.

## Update the Code in a Pipeline

The `pachctl update pipeline` updates the code that you use in one or
more of your pipelines. To apply your code changes, you need to
build a new Docker image and push it to your Docker image registry.

You can either use your registry instructions to build and push your
new image or push the new image by using the flags built into
the `pachctl update pipeline` command. Both procedures achieve the same goal,
and it is entirely a matter of a personal preference which one of them
to follow. If you do not have a build-push process that you
already follow, you might prefer to use Pachyderm's built-in functionality.

To create a new image by using the Pachyderm commands, you need
to use the `--build` flag with the `pachctl update pipeline`
command. By default, if you do not specify a registry with the `--registry`
flag, Pachyderm uses [DockerHub](https://hub.docker.com).
When you build your image with Pachyderm, it assigns a random
tag to your new image.

If you use a private registry or any other registry that is different
from the default value, use the `--registry` flag to specify it.
Make sure that you specify the private registry in the [pipeline
specification](../reference/pipeline_spec.md).

For example, if you want to push a `pachyderm/opencv` image to a
registry located at `localhost:5000`, you need to add this in
your pipeline spec:

 ```shell
 "image": "localhost:5000/pachyderm/opencv"
 ```

Also, to be able to build and push images, you need to make sure that
the Docker daemon is running. Depending on your operating system and
the Docker distribution that you use, steps for enabling it might
vary.

To update the code in your pipeline, complete the following steps:

1. Make the code changes.
1. Verify that the Docker daemon is running:

   ```shell
   $ docker ps
   ```

   * If you get an error message similar to the following:

     ```shell
     Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
     ```

     enable the Docker daemon. To enable the Docker daemon,
     see the Docker documentation for your operating system and platform.
     For example, if you use `minikube` on  macOS, run the following
     command:

     ```shell
     $ eval $(minikube docker-env)
     ```

1. Build, tag, and push the new image to your image registry:

   * If you prefer to use Pachyderm commands:

     1. Run the following command:

        ```shell
        $ pachctl update pipeline -f <pipeline name> --build --registry <registry> --username <registry user>
        ```

        If you use DockerHub, omit the `--registry` flag.

        **Example:**

        ```shell
        $ pachctl update pipeline -f edges.json --build --username testuser
        ```

     1. When prompted, type your image registry password:

        **Example:**

        ```
        Password for docker.io/testuser: Building pachyderm/opencv:f1e0239fce5441c483b09de425f06b40, this may take a while.
        ```

   * If you prefer to use instructions for your image registry:

     1. Build, tag, and push a new image as described in the
     image registry documentation. For example, if you use
     DockerHub, see [Docker Documentation](https://docs.docker.com/docker-hub/).

     1. Update the pipeline:

        ```shell
        $ pachctl update pipeline -f <pipeline.json>
        ```
