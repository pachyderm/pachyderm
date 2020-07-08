# Pipeline Makefile and Config template

An attempt to automate as much as possible of the work needed to test and setup

# Introduction

This is an attempt to automate the following procedures required for the
pipeline creation:

-   Build, push and pull the Docker image
-   Create secrets for Docker registry and the container
-   Create pipeline configuration file and the actual pipeline
-   Automate local tests and put them in an environment as close to production
    pipeline as possible
-   Cleanup if something goes wrong during testing or deploy

What this doesn't do:

-   Create or remove data repositories (due to dependencies)
-   Setup Kubernetes, Minikube, Docker Hub or any of the required infrastructure
    elements
-   Commit data to repositories
-   Write the core logic of your pipeline (that is still up to you)

## Configuring the pipeline

The folder have the following structure:

-   Configuration files are stored in the [config](./config) folder
-   Source files for magic are stored in the [src](./src) folder
-   `Makefile` holds all the voodoo for putting things together
-   `Dockerfile` tells docker how to build the container

All the configuration variables for the creation of the pipeline are stored in
the [pipeline.conf](./config/pipeline.conf) file. This includes pipeline name,
where the pipeline takes the input from etc. All the variables are commented in
the file so read on there for more details.

[pipeline.json](./config/pipeline.json) file holds the pachyderm specs for the
pipeline. For more information, see:
[Pipeline specs](https://docs.pachyderm.com/latest/reference/pipeline_spec/).

## Creating the pipeline

1. Make any customizations to `config/pipeline.conf` that you need
2. Ensure the repo specified in `config/pipeline.conf`'s `PIPELINE_REPO` exists
   in pachyderm: `pachctl create repo foobar`
3. Ensure these env vars are specified:

-   `$DOCKER_REGISTRY_USERNAME`
-   `$DOCKER_REGISTRY_PASSWORD`
-   `$DOCKER_REGISTRY_EMAIL`
-   Any env vars used in `config/secrets.yaml`

4. Run `make`. This will create a `target` folder with required configuration
   files.
5. Run `make install` to create a pipeline based on the created configuration.
6. After a while, run `make verify` to see if the job ran ok.

## Cleanup

`make clean` removes any files created during the installation but does not
remove the pipe. To explicitly remove the pipe, run `make pipe.delete`.

## Testing

How to run local tests:

-   By default, sample input data should be put in `./test/in` and expected
    output shows up in `./test/out`
-   Any environmental variables needed for testing should be put in the
    [docker.test.env](./config/docker.test.env) file and present in env when
    test is run

## Creating a new pipeline

In ~three~ nine easy steps:

1.  Copy an existing pipeline folder of your liking to a new folder:
    `cp -R old_pipe new_pipe`
2.  Change the input repository `$PIPELINE_REPO` variable in
    `new_pipe/config/pipeline.conf` to the appropriate new repo so your new
    pipeline gets the right input
3.  Put your source code in `new_pipe/src` and update `run.sh` to reflect the
    changes
4.  Update `Dockerfile` to include all the dependencies needed for your code
5.  Update `config/secrets.yaml` with any variables that are needed for your
    source code to run
6.  Put sample data in `test/in` and update the `config/docker.test.env`
    variables to your test needs. run `make test` and check `test/out` if
    everything works
7.  Save everything and run `make` and then `make install`
8.  (magic)
9.  Observe data flowing ...

## Platform-Specific Caveats

If you're on mac, make sure to install GNU gettext. Via homebrew:

```
brew install gettext
brew link --force gettext
```
