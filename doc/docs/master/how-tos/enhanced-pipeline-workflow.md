# Enhanced Pipeline Workflow

The majority of development time is spent in the pipeline development stage. The manual build steps mentioned in the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md) can become cumbersome during repetitive development cycles. There are two ways to reduce the number of manual steps while iterating on pipelines. 

## Build Pipelines
Build pipelines are a useful way to improve the iteration speed if you are working on the code in your pipeline. They allow you to bypass the Docker build process and submit your code to the pipeline without having to re-build the Docker image. In essence, build pipelines automate Steps 1, 4, and 5 of the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md).

!!! note
      Build pipelines are not currently supported in Pachyderm Hub. 

We accomplish this by using a PFS repository for any code changes and a PPS pipeline, which builds any code requirements and updates the pipeline. Since we are not building a new Docker image each time, we rely on a base image to contain any system requirements. 

Using build pipelines requires a modification of the pipeline specification, for example: 
```json
{
  "pipeline": {
    "name": "map"
  },
  "description": "A pipeline that tokenizes scraped pages and appends counts of words to corresponding files.",
  "transform": {
    "build": {
      "language": "go",
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

The pipeline is submitted the same way as a pipeline update, via 

```bash
pachctl update pipeline -f <pipeline name> 
```

When the build pipeline is submitted, it follows the following steps. 
1. Copies all files (code, etc.) from the build path to a PFS repository, `<pipeline name>_build`. This can be thought of as the code source repository In the case above, everything in `./source` would be copied to `<pipeline name>_build`.

!!! note
      You can optionally specify a `.pachignore` file to prevent certain files from getting pushed to this repo. 

1. Starts a pipeline called `<pipeline name>_build` that reads the files uploaded and runs the `build.sh` script to pull dependencies and compile any requirements from the source code. 

1. Creates (or updates if it already exists) the running pipeline `<pipeline name>` with the the newly creates source files and built assets. 

The updated pipeline will contain the following PFS repos mapped in as inputs:

1. `/pfs/source` - any source code that is required for running the pipeline. There are language-specific requirements on the structure of this source code mentioned momentarily.

1. `/pfs/build` - TODO what goes here? any artifacts resulting from the build process. 

1. `/pfl/<input(s)>` - any inputs specified in the pipeline specification

For convenience, Go and Python "builders" have a standarized implementation, only requiring setting the `language` parameter in the pipeline specification.  

### Builders
Either `language` or `image` must be specified in the pipeline spec. If `transform.image` is not specified, it is set to be the same as the builder image, which should work in most cases. Similarly, if `transform.cmd` is not specified, it will default to `sh /pfs/build/run.sh`, which should be provided via the build pipeline.

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
The `build.sh` and `run.sh` files are optional, unless the base image is modified. 


#### Go Builder


#### Creating a Builder

For languages other than Python and Go, or customizations to the official builders, users can author their own builders. Builders are somewhat similar to buildpacks in design, and follow a convention over configuration approach. A builder needs 3 things:

- A Dockerfile to bake the image.
- A build.sh in the image workdir, which acts as the entry-point for the build pipeline.
- A `run.sh`, injected into `/pfs/out` via `build.sh`. This will act as the entry-point for the executing pipeline. By convention, `run.sh` should take an arbitrary number of arguments and forward them to whatever executes the actual user code.

A custom builder can be used by setting transform.build.image in a pipeline spec. The official builders can be used for reference; they're located in etc/pipeline-build.

## --build 
If the Docker image is undergoing consistient changes, it can be difficult to keep up with changing the image tag and ensuring the image is pushed before updating the pipeline (Steps 2-5) in the [general pipeline workflow](../how-tos/working-with-data-and-pipelines.md). The `--build` flag automates these steps by building the Docker image, giving it a unique tag, and updating the tag in the pipeline json before submitting it to the pachyderm cluster. 



   ```bash
   pachctl update pipeline -f <pipeline name> --build --registry <registry> --username <registry user>
   ```