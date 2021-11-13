# Individual Developer Workflow

A typical Pachyderm workflow involves multiple iterations of
experimenting with your code and pipeline specs.

!!! info
    Before you read this section, make sure that you
    understand basic Pachyderm pipeline concepts described in
    [Concepts](../concepts/pipeline-concepts/index.md).

## How it works

Working with Pachyderm includes multiple iterations of the
following steps:

![Developer workflow](../assets/images/d_steps_analysis_pipeline.svg)

## Step 1: Write Your Analysis Code

Because Pachyderm is completely language-agnostic, the code
that is used to process data in Pachyderm can
be written in any language and can use any libraries of choice. Whether
your code is as simple as a bash command or as complicated as a
TensorFlow neural network, it needs to be built with all the required
dependencies into a container that can run anywhere, including inside
of Pachyderm. See [Examples](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples).

Your code does not have to import any special Pachyderm
functionality or libraries. However, it must meet the
following requirements:

* **Read files from a local file system**. Pachyderm automatically
  mounts each input data repository as `/pfs/<repo_name>` in the running
  containers of your Docker image. Therefore, the code that you write needs
  to read input data from this directory, similar to any other
  file system.

  Because Pachyderm automatically spreads data across parallel
  containers, your analysis code does not have to deal with data
  sharding or parallelization. For example, if you have four
  containers that run your Python code, Pachyderm automatically
  supplies 1/4 of the input data to `/pfs/<repo_name>` in
  each running container. These workload balancing settings
  can be adjusted as needed through Pachyderm tunable parameters
  in the pipeline specification.

* **Write files into a local file system**, such as saving results.
  Your code must write to the `/pfs/out` directory that Pachyderm
  mounts in all of your running containers. Similar to reading data,
  your code does not have to manage parallelization or sharding.

## Step 2: Build Your Docker Image

When you create a Pachyderm pipeline, you need
to specify a Docker image that includes the code or binary that
you want to run. Therefore, every time you modify your code,
you need to build a new Docker image, push it to your image registry,
and update the image tag in the pipeline spec. This section
describes one way of building Docker images, but
if you have your own routine, feel free to apply it.

To build an image, you need to create a `Dockerfile`. However, do not
use the `CMD` field in your `Dockerfile` to specify the commands that
you want to run. Instead, you add them in the `cmd` field in your pipeline
specification. Pachyderm runs these commands inside the
container during the job execution rather than relying on Docker
to run them.
The reason is that Pachyderm cannot execute your code immediately when
your container starts, so it runs a shim process in your container
instead, and then, it calls your pipeline specification's `cmd` from there.

After building your image, you need to upload the image into
a public or private image registry, such as
[DockerHub](https://hub.docker.com) or other.

Alternatively, you can use the Pachyderm's built-in functionality to
tag, build, and push images by running the `pachctl update pipeline` command
with the `--build` and `--push-images` flags. For more information, see
[Update a pipelines](updating_pipelines.md).

!!! note
    The `Dockerfile` example below is provided for your reference
    only. Your `Dockerfile` might look completely different.

To build a Docker image, complete the following steps:

1. If you do not have a registry, create one with a preferred provider.
If you decide to use DockerHub, follow the [Docker Hub Quickstart](https://docs.docker.com/docker-hub/) to
create a repository for your project.
1. Create a `Dockerfile` for your project. See the [OpenCV example](https://github.com/pachyderm/pachyderm/blob/1.13.x/examples/opencv/Dockerfile).
1. Log in to an image registry.

   * If you use DockerHub, run:

     ```shell
     docker login --username=<dockerhub-username> --password=<dockerhub-password> <dockerhub-fqdn>
     ```

1. Build a new image from the `Dockerfile` by specifying a tag:

   ```shell
   docker build -t <IMAGE>:<TAG> .
   ```

1. Push your image to your image registry.

   * If you use DockerHub, run:

     ```shell
     docker push <image>:tag
     ```

For more information about building Docker images, see
[Docker documentation](https://docs.docker.com/engine/tutorials/dockerimages/).


## Step 3: Create a Pipeline

Pachyderm's pipeline specifications store the configuration information
about the Docker image and code that Pachyderm should run. Pipeline
specifications are stored in JSON format. As soon as you create a pipeline,
Pachyderm immediately spins a pod or pods on a Kubernetes worker node
in which pipeline code runs. By default, after the pipeline finishes
running, the pods continue to run while waiting for the new data to be
committed into the Pachyderm input repository. You can configure this
parameter, as well as many others, in the pipeline specification.

A minimum pipeline specification must include the following
parameters:

- `name`
- `transform`
- `parallelism`
- `input`

You can store your pipeline locally or in a remote location, such
as a GitHub repository.

To create a Pipeline, complete the following steps:

1. Create a pipeline specification. Here is an example of a pipeline
   spec:

   ```shell
   # my-pipeline.json
   {
     "pipeline": {
       "name": "my-pipeline"
     },
     "transform": {
       "image": "my-pipeline-image",
       "cmd": ["/binary", "/pfs/data", "/pfs/out"]
     },
     "input": {
         "pfs": {
           "repo": "data",
           "glob": "/*"
         }
     }
   }
   ```

1. Create a Pachyderm pipeline from the spec:

   ```shell
   $ pachctl create pipeline -f my-pipeline.json
   ```

   You can specify a local file or a file stored in a remote
   location, such as a GitHub repository. For example,
   `https://raw.githubusercontent.com/pachyderm/pachyderm/1.13.x/examples/opencv/edges.json`.

!!! note "See Also:"

- [Pipeline Specification](../reference/pipeline_spec.md)
<!-- - [Running Pachyderm in Production](TBA)-->
