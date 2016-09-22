Pachyderm Pipeline System (PPS)
===============================

## Get Started

To get started using Pipelines, refer to our [Beginner Tutorial](./getting_started/beginner_tutorial.html) or [Pipeline](./development/custom_pipelines.html) docs.

## Overview

Pachyderm Pipeline System is a parallel, containerized analysis platform

It is designed to:

1. Write your analysis in any language of your choosing ([enabling Accessibility](https://pachyderm.io/dsbor.html)).
2. Allow you to compose your analyses
3. Allow you to reproduce your input data, your processing step, and your output data ([enabling Reproducibility](https://pachyderm.io/dsbor.html))
4. Allow you to understand the [Provenance](./advanced/provenance.html) of your data

## Components

PPS has two components, and understand each gives you a full picture of PPS.

### Jobs

Jobs are transformations that are only run once.

Broadly, they take the following inputs:

- a transformation image, [refer to the pipeline spec](./development/pipeline_spec.html) for instructions on creating your own image
  - an entry point to run the transformation
  - some other configuration options about how to run the job (parallelism, partitioning method, etc)
- at least one PFS input `Repo` containing some data
  - a `Commit` ID per input repo

When creating a job, PPS:

- creates an output `Repo` with the same name as the job
- uses kubernetes to spin up containers w the image you specify, in the configuration you specify
- mounts the input `Repo` at the `Commit` specified at `/pfs/your_repo_name` for use by your code on that container
- mounts `/pfs/out` for writing output, which is connected to the newly created output `Repo`
- runs the containers with the entry point you provided
- the output is stored in a new commit on the new output `Repo`

### Pipeline

Pipelines are configured once, but run every time new data is present in the form of a new `Commit` on any of their input `Repo`s. You can think of them as automatically up-to-date long-running jobs.

For detailed instructions on pipelines, [refer to the pipeline spec](./development/pipeline_spec.html)

## Provenance

You'll be using and composing pipelines frequently with PPS. Quickly, you're going to want to understand how your outputs are related to the inputs.

[Check out the flush-commit](./development/getting_your_results.html) docs for specifics on how to track provenance.

## Debugging tools

Beyond provenance, your primary triaging tool is [pachctl's logs](./pachctl/pachctl_get-logs.html). This allows you to see the log output per `Job` / `Pipeline` and debug any errors.

