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

Now we need to put some images in that repo. You can do this with `put-file`:

```sh
$ pachctl put-file images master -c -i examples/opencv/images.txt
```

With this command you're telling Pachyderm that you want to put files in the
`images` repo on the branch `master`. You're also passing two flags, `-c` means
that you want Pachyderm to start and finish the commit for you. If you wanted
to have multiple `put-files` write to the same commit you'd use `start-commit`
and `finish-commit`. `-i` is where you actually specify the images you're
inserting, `examples/opencv/images.txt` contains URLs for images, one per line.
Those images get scraped and inserted into Pachyderm, if you want to use your
own images pass a different file to `-i`. You can also reference files on the
local filesystem instead of URLs.

## Build and Distribute the Docker Image

Now that you've got some data in Pachyderm it's time to process it. To process
data in Pachyderm you first have to create a Docker image describing the
environment you want it to run in. You can build this image with:

```sh
$ docker build -t opencv examples/opencv
```

To understand what's going on take a look at
[`examples/opencv/Dockerfile`](/examples/opencv/Dockerfile). And
[`examples/opencv/edges.py`](/examples/opencv/edges.py).

## Deploy the Pipeline

## Stream More Data In.
