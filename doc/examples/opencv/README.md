# Distributed Image Processing with OpenCV and Pachyderm

## Intro

In this tutorial we're going to do edge detection using the Canny edge detection algorithm implemented in OpenCV. We'll be deploying our code as a Pachyderm pipeline which means it will be both streaming and distributed.

This tutorial assumes that you've already got a Pachyderm cluster up and running and that you can talk to it with the `pachctl` CLI tool. You'll know it's working if `pachctl version` returns without any errors. If not head on over to the [local installation](http://pachyderm.readthedocs.io/en/latest/getting_started/local_installation.html) or [production deployment](http://pachyderm.readthedocs.io/en/latest/development/deploy_intro.html) docs to get a cluster up and running. You should also clone the repo as we'll use the code and files provided. 

## Load Images Into Pachyderm
The first thing we'll need to do is create a repo to store our images in, we'll call the repo "images".

```sh
$ pachctl create-repo images
```

Now we need to put some images in that repo. You can do this with `put-file`:

```sh
$ pachctl put-file images master -c -i images.txt
```

With this command you're telling Pachyderm that you want to put files in the `images` repo on the branch `master`. You're also passing two flags, `-c` and `-i`.`-c` means that you want Pachyderm to start and finish the commit for you. If you wanted to have multiple `put-files` write to the same commit you'd use `start-commit` and `finish-commit`. `-i` lets you specify a line-delimited input file. Each line can be a URL to scrape or a local file to add. In our case, we've got a file `images.txt` with a an image URL that we'll add to Pachyderm.

## Build and Distribute the Docker Image

Now that you've got some data in Pachyderm, it's time to process it. First, you need to create a Docker image containing the code we will use to perform our edge detection. For this example, you can use the supplied `Dockerfile`. You can build the Docker image with:

```sh
$ docker build -t opencv .
```

To understand what's going on, take a look at the [Dockerfile](./Dockerfile) and [edges.py](./edges.py).

### Distribute the Docker Image
You'll need to push the image to a registry, such as DockerHub, so Pachyderm can find it. You can do this with:

```sh
$ docker tag opencv <your-docker-hub-username>/opencv
$ docker push <your-docker-hub-username>/opencv
```

Now the image can be referenced on any Docker host as `<your-docker-hub-username>/opencv` and Docker will be able to find it.

## Deploy the Pipeline

Now that you have an image, you need to tell Pachyderm how to run it. To do this, you'll need to create a pipeline. The pipeline for this example is [edges.json](./edges.json).

First, notice that this file references the image you just created in the field `transform.image`. Since you just pushed your image to Dockerhub, you'll need to tell Pachyderm to go there to find it. Open `edges.json` in a text editor, and change the following line:

```sh
"image": "opencv"
## should become ##
"image": "<your-docker-hub-username>/opencv"
```

Now that `edges.json` points to your image, you can create the pipeline with:

```sh
# this cmd assumes you've cloned our repo. If not, use the GitHub raw file URL since `-f` can take a URL as well.
$ pachctl create-pipeline -f edges.json
```

You can check on the status of the pipeline with:

```sh
$ pachctl inspect-pipeline edges
```

You should see that a single job has been run (it was run on the initial commit you made in the images repo).

## Inspect the Results

Results from a pipeline are stored in an output repo whose name matches the name of the pipeline. You can see it with `list-repo`

```sh
$ pachctl list-repo
NAME    CREATED             SIZE
edges   11 minutes ago      557.7 KiB
images  15 minutes ago      1.136 MiB
```

You can view this data by mounting it locally:

```sh
$ mkdir /tmp/pfs
$ pachctl mount /tmp/pfs &
# This command will block (so we run in the background)
```

While the mount is up navigate to [`file:///tmp/pfs`](file:///tmp/pfs) in your web browser. Here you can browse the contents of pfs which should be the raw images and the results of the edge detection.

## Stream More Data In

Pipelines are smart. They don't just process the input data that's present when they're created, they also process any new data that's added later. You can process any image on the internet or your local disk, all you have to do is put it in the images repo (with `put-file`) and your pipeline will automatically trigger and run the edge detection code.

## Next Steps

OpenCV can do a lot more than edge detection, and now that you've built the container image, you can use anything in OpenCV on Pachyderm. OpenCV has several [Python tutorials available online](https://opencv-python-tutroals.readthedocs.io/en/latest/py_tutorials/py_tutorials.html) if you're looking for inspiration.
