# Creating Analysis Pipelines

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



## Basing off of a 3rd party Image

e.g. 

```
FROM ubuntu

then ...
do our installation steps manually here
```


## Basing off of our image


```
FROM job-shim

# your custom requirements

```


Creating a pipeline
-------------------

pipeline spec

Adding pipline to Pachyderm


Next steps:
-----------
Track your data Provenance
Learn how to updating pipelines and explore your data