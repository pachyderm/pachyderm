# The Repo Pipeline Spec v0.1

## Preamble
This spec is in a very rough form right now. It has a lot of Pachyderm specific
stuff in it which isn't what we want for this spec long term. Our hope is that
it can serve to start the conversation so we can eventually arrive at something
that makes everyone happy.

## Motivation
The Pachyderm Pipeline Spec is a means to defining distributed computation
pipelines with a simple directory structure. It's designed to be implemented by
distributed computation engines and coupled with post receive hooks to give
users a generic `git push` interface to distributed computing engines.

## The Spec
The spec uses a simple directory structure to define to express the pipeline.
For a directory to satisfy the spec it must contain the following
subdirectories:

### `/image`
The `/image` directory is used to build a Docker image for the computation. The
file `/image/Dockerfile` must be present for this to work.  The image is built
with the command:

```shell
docker build /image
```

### `/job`
The `/job` directory defines the pipeline of jobs which will be run by the
engine. Each file be a JSON description of a job according to the [pfs job
spec](https://github.com/pachyderm/pfs/blob/master/README.md#creating-a-new-job-descriptor).
Jobs may reference the image which will be built from the `/image` directory as
`{{REPO_IMAGE}}`. The job may reference any data which it expects to be stored
in the filesystem.

### `/install`
The `/install` directory contains scripts for installing the pipeline on various
engines.
Each engine is represented by a directory in `/install`. For example, pachyderm
is represented by `/install/pachyderm`.  Each directory must contain a script
called `install`.  The script should install the repo on an engine when called
like this:

```shell
install/pachyderm/install {{HOST_NAME}}
```

the directory can contain other files too but hopefully shouldn't be bigger
than some small limit. [TODO figure out what that limit should be].

### `/data (optional)`
The `/data` is an optional directory that can contain samples of the data one
should expect to find on the other end. This is useful because it allows users
to test jobs locally which speeds up the development cycle. Maintainers are
encouraged to keep stats in their engines about which files produce the most
failures and include those in `/data` to increase the efficacy of local
testing.
