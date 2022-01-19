# Update a Pipeline

While working with your data, you often need to modify an existing
pipeline with new transformation code or pipeline parameters.
For example, you might want to try different model versions
while your training data stays constant. 

You need to use the `pachctl update pipeline` command to make changes to a pipeline,
whether you have re-built a docker image after a code change and/or
need to update pipeline parameters in the pipeline specification file. 

Alternatively, you can update a pipeline using [pipeline templates](#using-pipeline-templates).

## Update Your Pipeline Specification

### After You Changed Your Specification File

Run the `pachctl update pipeline` command to apply any change to your
[pipeline specification](../../reference/pipeline_spec) JSON file, such as change to the
parallelism settings, change of an image tag, change of an input repository, etc...

By default, a pipeline update does not trigger the reprocessing of the data
that has already been processed. Instead,
it processes only the new data you submit to the input repo.
If you want to run the changes in your pipeline against the data in
your input repo's `HEAD` commit, use the `--reprocess` flag.
The updated pipeline will then continue to process new input data only.
Previous results remain accessible through the corresponding commit IDs.

To update a pipeline specification, run the following command after
you have updated your pipeline specification JSON file.

```shell
pachctl update pipeline -f pipeline.json
```

!!! Note
    Similar to `create pipeline`, `update pipeline` with the `-f` flag can also
    take a URL if your JSON manifest is hosted on GitHub or other remote location.

### Using Pipeline Templates
[Pipeline templates](../pipeline-template) allow you to bypass the "update you specification file" step and 
apply your changes at once by running:

```shell
pachctl update pipeline --jsonnet <your template path or URL> --arg <param 1>=<value 1> --arg <param 2>=<value 2>
```
!!! Example
      ```shell
      pachctl update pipeline --jsonnet templates/edges.jsonnet --arg suffix=1 --arg tag=1.0.2
      ```

## Update the Code in a Pipeline

To apply your code changes, you need to
build a new Docker image and push it to your Docker image registry.

You can either use your registry instructions to build and push your
new image or push the new image by using the flags built into
the `pachctl update pipeline` command. Both procedures achieve the same goal,
and it is entirely a matter of a personal preference which one of them
to follow. 

To push a new image by using the Pachyderm commands, use the `--push-images` flag
with the `pachctl update pipeline` command. 
By default, if you do not specify a registry with the `--registry`
flag, Pachyderm uses [DockerHub](https://hub.docker.com){target=_blank}.
When you build your image with Pachyderm, it assigns a random
tag to your new image.

If you use a private registry or any other registry that is different
from the default value, use the `--registry` flag to specify it.
Check the required fields in the [pipeline
specification file](../../../reference/pipeline_spec/#transform-required).

For example, if you want to push a `pachyderm/opencv` image to a
registry located at `localhost:5000`, add this in
your pipeline spec:

 ```shell
 "image": "localhost:5000/pachyderm/opencv"
 ```

Make sure that the Docker daemon is running. 
Depending on your operating system and
the Docker distribution that you use, steps for enabling it might
vary.

To update the code in your pipeline, complete the following steps:

1. Make the code changes.
1. Verify that the Docker daemon is running:

     ```shell
     docker ps
     ```
     If you get an error message similar to the following:

     ```shell
     Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?
     ```
     enable the Docker daemon. To enable the Docker daemon,
     see the Docker documentation for your operating system and platform.
     For example, if you use `minikube` on  macOS, run the following
     command:

     ```shell
     eval $(minikube docker-env)
     ```

1. Build, tag, and push the new image to your image registry:

      * **If you use Pachyderm commands**

         1. [Build your new image](../../developer-workflow/working-with-pipelines/#step-2-build-your-docker-image) using `docker build` (for example, in a makefile: `@docker build --platform linux/amd64 -t $(DOCKER_ACCOUNT)/$(CONTAINER_NAME) .`). No tag needed, the folllowing [`--push-images` flag](../../developer-workflow/push-images-flag/) flag will take care of it.

      
         1. Run the following command:

            ```shell
            pachctl update pipeline -f <pipeline name> --push-images --registry <registry> --username <registry user>
            ```

            If you use DockerHub, omit the `--registry` flag.

            **Example:**

            ```shell
            pachctl update pipeline -f edges.json --push-images --username testuser
            ```

         1. When prompted, type your image registry password:

            **Example:**

            ```
            Password for docker.io/testuser: Building pachyderm/opencv:f1e0239fce5441c483b09de425f06b40, this may take a while.
            ```

      * **If you prefer to use instructions for your image registry**

         1. Build, tag, and push a new image as described in your
          image registry documentation. For example, if you use
          DockerHub, see [Docker Documentation](https://docs.docker.com/docker-hub/){target=_blank}.

         1. Update the `transform.image` field of your pipeline spec with your new tag.
         
            !!! Important
                  Make sure to update your tag every time you re-build. Our pull policy is `IfNotPresent` (Only pull the image if it does not already exist on the node.). Failing to update your tag will result in your pipeline running on a previous version of your code.

         1. Update the pipeline:

            ```shell
            pachctl update pipeline -f <pipeline.json>
            ```

      * **If you choose to use a [templated version of your pipeline](./pipeline-template.md)**

         1. Pass the tag of your image to your template.

            As an example, see the `tag` parameter in this templated version of opencv's edges pipeline (`edges.jsonnet`):
            
            ```shell
            function(suffix, tag)
               {
               pipeline: { name: "edges-"+suffix },
               description: "OpenCV edge detection on "+src,
               input: {
                  pfs: {
                     name: "images",
                     glob: "/*",
                     repo: "images",
                  }
               },
               transform: {
                  cmd: [ "python3", "/edges.py" ],
                  image: "pachyderm/opencv:"+tag
               }
               }
            ```

         1. Once your pipeline code is updated and your image is built, tagged, and pushed, update your pipeline using this command line. In this case, there is no need to edit the pipeline specification file to update the value of your new tag. This command will take care of it:

            ```shell
            pachctl update pipeline --jsonnet templates/edges.jsonnet --arg suffix=1 --arg tag=1.0.2
            ```

