# Enhanced Pipeline Workflow

The manual build steps mentioned in the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md) can become cumbersome during development cycles. Here we introduce two ways to automate these manual steps, namely build pipelines and the `--build` flag. 

## Build Pipelines

A build pipeline is a useful feature when iterating on the code in your pipeline. They allow you to bypass the Docker build process and submit your code directly to the pipeline (without having to re-build the Docker image). In essence, build pipelines automate Steps 1, 4, and 5 of the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md).

!!! note
      Build pipelines are not currently supported in Pachyderm Hub. 

Functionally, the build pipeline uses a base Docker image that remains unchanged during the development process. The code is copied into this pipeline at runtime from a PFS repository. The build pipeline installs requirements and updates the code in the pipeline before running running it.

Build pipelines require a modification to the pipeline spec's `transform` object with options for the following fields:

- `path`: An optional string specifying where the source code is relative to the pipeline spec path (or the cwd if the pipeline is fed into `pachctl` via stdin.)
- `language`: An optional string specifying what language builder to use (see below.) Only works with official builders. If unspecified, `image` will be used instead.
- image: An optional string specifying what builder image to use, if a non-official builder is desired. If unspecified, the `transform` object's `image` will be used instead.

Below is a Python example of a build pipline.

```json
{
  "pipeline": {
    "name": "map"
  },
  "description": "A pipeline that tokenizes scraped pages and appends counts of words to corresponding files.",
  "transform": {
    "build": {
      "language": "python",
      "path": "./source"
    }
  },
  "input": {
    "pfs": {
      "repo": "scraper",
      "glob": "/*"
    }
  }
}
```

A build pipeline can be submitted the same way as any other pipeline, for example:

```bash
pachctl update pipeline -f <pipeline name> 
```

When submitted, the following actions occur: 

1. All files (code, etc.) are copied from the build path to a PFS repository, `<pipeline name>_build`, which we can think of as the source code repository. In the case above, everything in `./source` would be copied to to the PFS `<pipeline name>_build` repository.

1. Starts a pipeline called `<pipeline name>_build` that reads the files uploaded and runs the `build.sh` script to pull dependencies and compile any requirements from the source code. 

1. Creates (or updates if it already exists) the running pipeline `<pipeline name>` with the the newly creates source files and built assets. 

!!! note
      You can optionally specify a `.pachignore` file to prevent certain files from getting pushed to this repo. 

The updated pipeline contains the following PFS repos mapped in as inputs:

1. `/pfs/source` - source code that is required for running the pipeline. There are language-specific requirements on the structure of this source code mentioned in the [Python Builder](#python-builder) section.

1. `/pfs/build` - any artifacts resulting from the build process. 

1. `/pfs/<input(s)>` - any inputs specified in the pipeline specification.

### Builders
Either `language` or `image` must be specified in the pipeline spec. The langauge parameter informs the pipeline of what base image to use and already contain implementations for `build.sh` and `run.sh`. Go and Python "builders" have a standarized implementation, only requiring setting the `language` parameter in the pipeline specification.  

If `language` is not specified, then the build pipeline uses the `transform.image` container (if `transform.image` is not specified the builder image is used). Similarly, if `transform.cmd` is not specified, it will default to running `sh /pfs/build/run.sh`, which should be provided via the build pipeline.

#### Python Builder

The Python builder requires a file structure similar to the following:

```tree
./map
├── source
│   ├── requirements.txt
│   ├── main.py
│   ├── build.sh (optional)
│   └── run.sh (optional)
```
There must exist a `main.py` which acts as the entrypoint for the pipeline. Optionally, a `requirements.txt` can be used to specify pip packages that will be installed during the build process. 

The `build.sh` and `run.sh` files are optional, as the Python Builder already contains these scripts. However, if the base image is modified, they must be provided.

#### Go Builder

The Go Builder follows the same format as the [Python Builder](#python-builder). There must be a main source file in the source root that imports and invokes the intended code.

#### Creating a Builder

For languages other than Python and Go, or customizations to the official builders, users can author their own builders. Builders are somewhat similar to buildpacks in design, and follow a convention over configuration approach. A builder needs 3 things:

- A Dockerfile to bake the image.
- A `build.sh` in the image workdir, which acts as the entry-point for the build pipeline.
- A `run.sh`, injected into `/pfs/out` via `build.sh`. This will act as the entry-point for the executing pipeline. By convention, `run.sh` should take an arbitrary number of arguments and forward them to whatever executes the actual user code.

A custom builder can be used by setting `transform.build.image` in a pipeline spec. The official builders can be used for reference; they're located in `etc/pipeline-build`.

## --build 
In addition to Build Pipelines, the `--build` flag is another way to improve development speed when working with pipelines. While build pipelines avoid rebuilding the docker image, the `--build` flag builds, tags and pushes the new Docker image. This feature can be particularly useful while iterating on the Docker image itself, as it can be difficult to keep up with changing image tags and ensure the image is pushed before updating the pipeline (Steps 2-5 in the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md)).

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
      For more details on the `--build` flag, see [Update a Pipeline](../updating_pipelines/#update-the-code-in-a-pipeline).
