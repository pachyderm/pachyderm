Deploying Pachyderm
===================

## Setup

First time user? [Follow the setup instructions](./deploying_setup.html)

## Pipelines

Pipelines are the bread and butter of [Pachyderm Pipeline System (PPS)](./pachyderm_pipeline_system.html). You'll find yourself creating and composing these a lot. For more information on `Pipelines` refer to the [Pipeline Spec](./pipeline_spec.html)

## Inputting data into PFS

To start using Pachyderm, you'll need to input some of your data into a [PFS](./pachyderm_file_system.html) repo.

There are a handful of ways to do this:

1) You can do this via a [PFS](./pachyderm_file_system.html) mount

The [Fruit Stand](https://github.com/pachyderm/pachyderm/tree/master/examples/fruit_stand) example uses this method. This is probably the easiest to use if you're poking around and make some dummy repo's to get the hang of Pachyderm.

2) You can do it via the [pachctl](./pachctl.html) CLI

This is probably the most reasonable if you're scripting the input process.

3) You can use the golang client

You can find the docs for the client [here](https://godoc.org/github.com/pachyderm/pachyderm/src/client)

Again, helpful if you're scripting something, and helpful if you're already using go.

4) You can dump SQL into pachyderm

You'll probably use the golang client to do this. Example coming soon.

## Generating your custom image

When creating a pipeline, you need to specify an image. Pachyderm Pipeline System (PPS) allows you to use any Docker image you'd like so that you can transform your data with your tools of choice.

There are two ways to construct the image. Both require some familiarity w [Dockerfiles](https://docs.docker.com/engine/tutorials/dockerimages/#/building-an-image-from-a-dockerfile).

Since a Dockerfile can only have a single `FROM` directive, the two options center around if its easier for you to base your image off of a Pachyderm image or your own.

### Using the Pachyderm Job Shim

Use this method if your dependencies are simple.

In this case, you would base your image off of Pachyderm's job shim image:

```
FROM pachyderm/job-shim:latest
```

[Here is an example](https://github.com/pachyderm/pachyderm/blob/master/examples/word_count/Dockerfile)

In this example, the transformation code was written in go, so in addition to using the job-shim image, this Dockerfile installed go1.6.2 and compiled the program needed for the transformation. 

### Using a 3rd Party Image

Use this method if your dependencies are complex.

Most often this method is used if there is a standard published 3rd party image that already contains your dependencies (e.g. [tensor flow](https://github.com/pachyderm/pachyderm/blob/master/examples/tensor_flow/Dockerfile)) or if you have a custom image you manage yourself.

In this case, the `FROM` directive will specify the 3rd party image of your choice. But to satisfy the requirements that Pachyderm Pipeline System (PPS) needs, you'll have to add a few lines to your Dockerfile. In addition to the tensor flow example above, [here is the canonical Dockerfile](https://github.com/pachyderm/pachyderm/blob/master/etc/user-job/Dockerfile) that lists the dependencies you need to add.

## Going to Production

When you're ready to deploy your pipelines in production, we recommend the following:

1) Use [the Kubernetes Dashboard](https://github.com/kubernetes/dashboard) for monitoring cluster stats / usage
2) Making sure you hook in a 3rd party object store

By default, a local k8s cluster / pachyderm cluster will use the local filesystem for storage. This is fine for development, but you should make sure you have S3/GCS setup to go into production. [Refer to the setup instructions](./deploying_setup.html) for more info.


3) Make sure the code behind your transform images is under source control
4) Load test / scale up your nodes so that you're confident you can handle the scale you need

