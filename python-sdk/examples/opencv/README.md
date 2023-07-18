# Distributed Image Processing with OpenCV and Python Pachyderm

This is a reproduction of Pachyderm's opencv example in python. A walkthrough is available in the Pachyderm *[docs](https://docs.pachyderm.io/en/latest/getting_started/beginner_tutorial.html)*

This example showcases the pachyderm-sdk analogs of common `pachctl` commands, such as creating repos and pipelines or getting a file.

The image being used to run the pipeline code is built from the Dockerfile located at [pachyderm/examples/opencv](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv). You will also find the edges.py script there.

**Prerequisites:**
- A running [Pachyderm cluster](https://docs.pachyderm.com/latest/get-started/)
- Install the latest package of pachyderm-sdk

**To run:**
```shell
$ python opencv.py
```