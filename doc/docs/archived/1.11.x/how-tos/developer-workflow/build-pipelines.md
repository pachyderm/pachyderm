# Build Pipelines

A build pipeline is a useful feature when iterating on the code in your pipeline. They allow you to bypass the Docker build process and submit your code directly to the pipeline. In essence, build pipelines automate Steps 2-4 of the [pipeline workflow](working-with-pipelines.md). A diagram of the build pipeline process is shown below.

![Developer workflow](../../assets/images/d_steps_build_pipeline.svg)


Functionally, a build pipeline relies on a base Docker image that remains unchanged during the development process. Code and build assets are stored in Pachyderm itself and copied into the pipeline pod when it executes.

To enable this feature, add a `build` object to the pipeline spec's `transform` object, with the following fields:

- `path`: An optional string specifying where the source code is relative to the pipeline spec path (or the current working directory if the pipeline is fed into `pachctl` via stdin.)
- `language`: An optional string specifying what language builder to use (see below). Only works with official builders. If unspecified, `image` will be used instead.
- `image`: An optional string specifying what builder image to use, if a non-official builder is desired. If unspecified, the `transform` object's `image` will be used instead.

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

```shell
pachctl update pipeline -f <pipeline name>
```

## How it works

When a build pipeline is submitted, the following actions occur:

1. All files from the pipeline build `path` are copied to a PFS repository, `<pipeline name>_build`, which we can think of as the source code repository. In the case above, everything in `./source` would be copied to to the PFS `map_build` repository.

2. A pipeline that uses the same repo but a different branch starts, reads the source code and creates build assets (i.e. pulling in dependencies and/or compiling) by running a `build.sh` script.

3. The running pipeline, `<pipeline name>`, is updated with the the new source files and built assets then executes `sh /pfs/build/run.sh` when a job for that pipeline is started.

!!! note
      You can optionally specify a `.pachignore` file in the source root directory, which uses [ohmyglob](https://github.com/pachyderm/ohmyglob) entries to prevent certain files from getting pushed to this repo.

The updated pipeline contains the following PFS repos mapped in as inputs:

1. `/pfs/source` - source code that is required for running the pipeline.

1. `/pfs/build` - any artifacts resulting from the build process.

1. `/pfs/<input(s)>` - any inputs specified in the pipeline spec.

## Builders

The builder interprets the pipeline spec to determine:

* A Docker image to use as the base image.
* Steps to run for the build.
* Step to run upon deployment.
  
The `transform.build.language` field is solely required to use an official builder (currently `python` or `go`), which already have impelmentations for `build.sh` and `run.sh`.

### Python Builder

The Python builder relies on a file structure similar to the following:

```tree
./map
├── source
│   ├── requirements.txt
│   ├── ...
│   └── main.py
```
There must exist a `main.py` which acts as the entrypoint for the pipeline. Optionally, a `requirements.txt` can be used to specify pip packages that will be installed during the build process. Other supporting files in the directory will also be copied and available in the pipeline if they are not excluded by the `.pachignore`.


### Go Builder

The Go Builder follows the same format as the [Python Builder](#python-builder). There must be a main source file in the source root that imports and invokes the intended code.

### Creating a Builder

Users can author their own builders for languages other than Python and Go (or customizations to the official builders). Builders are somewhat similar to buildpacks in design, and follow a convention over configuration approach. The current [official builders](https://github.com/pachyderm/pachyderm/tree/master/etc/pipeline-build) can be used for reference.

A builder needs 3 things:

- A Dockerfile to bake the image specified in the build pipeline spec.
- A `build.sh` in the image workdir, which acts as the entry-point for the build pipeline.
- A `run.sh`, injected into `/pfs/out` via `build.sh`. This will act as the entry-point for the executing pipeline. By convention, `run.sh` should take an arbitrary number of arguments and forward them to whatever executes the actual user code.

And the build file structure would look similar to the following:

```tree
<language>
├── Dockerfile
├── build.sh
└── run.sh
```

The `transform.build.image` in the pipeline spec is used to define the base image for unofficial builders. The order of preference for determining the Docker image is:

1. `transform.build.language`
2. `transform.build.image`
3. `transform.image`

The convention is to provide `build.sh` and `run.sh` scripts to fulfill the build pipeline requirements; however, if a `transform.cmd` is specified, it will take precedence over `run.sh`.
