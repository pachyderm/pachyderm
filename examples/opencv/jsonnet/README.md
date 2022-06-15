# Jsonnet Pipeline Specs Applied To Pachyderm's openCV Beginner Tutorial

>![pach_logo](../../img/pach_logo.svg) This example requires Pachyderm 2.1 or later versions.

We adapted our opencv tutorial to use jsonnet pipeline specs. In this adapted version, we use 2 jsonnet specs, `edges.jsonnet` and `montage.jsonnet`, to generate our opencv pipelines.
We then extend the use case and re-apply the montage jsonnet specs a second time, creating another pipeline, identical to the first `montage-1` but different in name `montage-2` and input repositories.
## Prerequisites

We assume that you have a running version of Pachyderm 2.1 or later, its associated pachctl CLI, and are familiar with the original opencv tutorial.

```shell
pachctl version
```
```
COMPONENT           VERSION
pachctl             2.1.0
pachd               2.1.0
```

To prepare for this example, create and populate the `images` repository ahead of time by running:
```shell
pachctl create repo images
pachctl put file images@master:liberty.png -f http://imgur.com/46Q8nDz.png
pachctl put file images@master:AT-AT.png -f http://imgur.com/8MN9Kg0.png
pachctl put file images@master:kitten.png -f http://imgur.com/g2QnNqa.png
```

## Create The Edge Detection Pipeline

Using [`edges.jsonnet`](../edges.jsonnet), we are going to create the pipeline `edges-1`. The pipeline takes the repository `images` as its input repo:

```shell
pachctl create pipeline --jsonnet edges.jsonnet  --arg suffix=1 --arg src=images
```

Check the state of your pipeline by running `pachctl list pipeline`.
```
NAME    VERSION INPUT     CREATED       STATE / LAST JOB  DESCRIPTION
edges-1 1       images:/* 2 minutes ago running / success A pipeline that performs image edge detection by using the OpenCV library.
```

Hopefully, your pipeline ran successfully. An output repo of the same name should have been created.
```shell
pachctl list repo
```

```
NAME    CREATED            SIZE (MASTER) DESCRIPTION
edges-1 35 seconds ago     ≤ 133.6KiB    Output repo for pipeline edges-1.
images  About a minute ago ≤ 238.3KiB
```
To look at one of its files, run:

- On macOS, run:
  ```shell
  pachctl get file edges-1@master:liberty.png | open -f -a Preview.app
  ```
- On Linux 64-bit, run:
  ```shell
  pachctl get file edges-1@master:liberty.png | display
  ```

## Create The Montage Pipeline

Let's add a montage pipeline that arranges our images and their "edge-detected" version into a single montage. The pipeline takes 2 input repositories (`images` and `edges-1`) which we will pass as arguments to the second jsonnet pipeline spec `montage.jsonnet` along with the suffix `1`.

```shell
pachctl create pipeline --jsonnet montage.jsonnet  --arg suffix=1 --arg left=images --arg right=edges-1
```

Note that you can pass the full url of your jsonnet location.

Again, check the state of your pipeline by running `pachctl list pipeline`.
```
NAME    VERSION INPUT     CREATED       STATE / LAST JOB  DESCRIPTION
NAME      VERSION INPUT                  CREATED        STATE / LAST JOB  DESCRIPTION
montage-1 1       (images:/ ⨯ edges-1:/) 25 seconds ago running / success A pipeline that combines images from images@master and edges-1@master into a montage.
edges-1   1       images:/*              2 minutes ago  running / success OpenCV edge detection on images
```

An output repo of the same name should have been created.
```shell
pachctl list repo
```

```
NAME      CREATED        SIZE (MASTER) DESCRIPTION
montage-1 53 seconds ago ≤ 1.292MiB    Output repo for pipeline montage-1.
edges-1   3 minutes ago  ≤ 133.6KiB    Output repo for pipeline edges-1.
images    4 minutes ago  ≤ 238.3KiB
```

Take a look at your first montage by running:

- On macOS, run:
  ```shell
  pachctl get file montage-1@master:montage.png | open -f -a Preview.app
  ```
- On Linux 64-bit, run:
  ```shell
  pachctl get file montage-1@master:montage.png | display
  ```

## Next: Re-Apply The Montage Jsonnet Spec To Different Input Repositories

Let's re-apply the previous `montage.jsonnet` to our `images` repository and the `montage-1` output repo this time. We will call it `montage-2`.

```shell
pachctl create pipeline --jsonnet montage.jsonnet  --arg suffix=2 --arg left=edges-1 --arg right=montage-1
```

`pachctl list pipeline` shows that you have indeed created a 3rd pipeline.

```
NAME      VERSION INPUT                    CREATED        STATE / LAST JOB  DESCRIPTION
montage-2 1       (edges-1:/ ⨯ montage-1:/) 8 seconds ago  running / running A pipeline that combines images from images@master and montage-1@master into a montage.
montage-1 1       (images:/ ⨯ edges-1:/)   15 minutes ago running / success A pipeline that combines images from images@master and edges-1@master into a montage.
edges-1   1       images:/*                18 minutes ago running / success OpenCV edge detection on images
```

Take a look at your end result:

- On macOS, run:
  ```shell
  pachctl get file montage-2@master:montage.png | open -f -a Preview.app
  ```
- On Linux 64-bit, run:
  ```shell
  pachctl get file montage-2@master:montage.png | display
  ```
You just added your last montage to your edges images.
