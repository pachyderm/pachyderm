# Distributed Image Processing with OpenCV and Pachyderm

## Intro

In this tutorial we're going to do edge detection using the Canny edge
detection algorithm implemented in OpenCV. We'll be deploying our code as a
Pachyderm pipeline which means it will be both streaming and distributed.

This tutorial assumes that you've already got a Pachyderm cluster up and
running and that you can talk to it with pachctl. You'll know it's working if
`pachctl version` returns without any errors. If not head on over to the [setup
guide](/doc/deploying_setup.md) to get a cluster up and running.

## Load Images Into Pachyderm
The first thing we'll need to do is create a repo to store our images in, we'll
call the repo "images".

```sh
$ pachctl create-repo images
```

Now we need to put something in that repo.

## Build and Distribute The Code

## Deploy the Pipeline

## Stream More Data In.
