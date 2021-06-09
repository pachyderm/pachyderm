>![pach_logo](../../img/pach_logo.svg) INFO - Pachyderm 2.0 introduces profound architectural changes to the product. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Branch Master: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples
# Object detection

In this example we're going to use the [Tensorflow Object Detection API](https://github.com/tensorflow/models/tree/master/object_detection) to do some general object detection and we'll use Pachyderm to set up the necessary data pipelines to feed in the data. 

## Prerequisites
1. Clone this repo.
2. Install Pachyderm as described in [Local Installation](https://docs.pachyderm.com/1.13.x/getting_started/local_installation/).

## 1. Make Sure Pachyderm Is Running

You should be able to connect to your Pachyderm cluster via the `pachctl` CLI.  To verify that everything is running correctly on your machine, run the following:

```shell
$ pachctl version
COMPONENT           VERSION
pachctl             1.7.4
pachd               1.7.4
```

## 2. Create The Input Data Repositories

```shell
$ pachctl create repo training
$ pachctl create repo images
```
Make sure the repos are there

```shell
$ pachctl list repo
NAME     CREATED       SIZE
images   4 seconds ago 0B
training 8 seconds ago 0B
```

## 3. Fetch The Data And Extract It Locally

```shell
$ wget http://download.tensorflow.org/models/object_detection/ssd_mobilenet_v1_coco_11_06_2017.tar.gz
$ tar -xvf ssd_mobilenet_v1_coco_11_06_2017.tar.gz
```

## 4. Import Data Into The Pachyderm Repos
`cd` into the newly extracted folder

```shell
$ cd ssd_mobilenet_v1_coco_11_06_2017
```
Add in inference graph  
```shell
$ pachctl put file training@master -f frozen_inference_graph.pb
```

## 5. Build The Pachyderm Pipelines
```shell
cd ../
```

```shell
$ pachctl create pipeline -f model.json
$ pachctl create pipeline -f detect.json
```

Now we can check on the pipelines and make sure they're running

```shell
$ pachctl list pipeline
NAME   INPUT                 OUTPUT        CREATED        STATE
detect (images:/* ⨯ model:/) detect/master 9 seconds ago  running
model  training:/            model/master  17 seconds ago running
```

You can also see the jobs that were created by our pipelines as well as their status.

```shell
$ pachctl list job
ID                               OUTPUT COMMIT                           STARTED        DURATION  RESTART PROGRESS  DL       UL STATE
ad132094bcba4f89a4effffee8f7bb1c detect/da0ac9ffcbdc4f2fabeb79222f628a8d 9 seconds ago  3 seconds 0       0 + 0 / 0 0B       0B success
a0c71b182c0d4a649689673d4eb0d9ee model/b2a87f54356f48e29486ea7777326d63  18 seconds ago 3 seconds 0       1 + 0 / 1 27.83MiB 0B success
```

Another thing you'll notice is that these pipelines created two new repos (which we'll use in the next step).

```shell
$ pachctl list repo
NAME     CREATED            SIZE
detect   47 seconds ago     0B
model    56 seconds ago     27.83MiB
training About a minute ago 27.83MiB
images   About a minute ago 0B
```

## 6. Commit Images Into The `Images` Repo And Get The Output Of The Object Detection API

```shell
$ cd images
```
Add the `airplane.jpg` into your `images` repo

```shell
$ pachctl put file images@master -f airplane.jpg
```
Once the image has been evaluated by Object Detection API you'll be able to see the detection result in the `detect` repo. We can take a look at the result by running the following

```shell
# on macOS
$ pachctl get file detect@master:airplane.jpg | open -f -a /Applications/Preview.app

# on Linux
$ pachctl get file detect@master:airplane.jpg | display
```

![alt text](detected_airplane.jpg)

## 7. Your Turn
There are few other images in the directory. Run through step 6 again but this time use one of the other images and see what the Object Detection API returns.
