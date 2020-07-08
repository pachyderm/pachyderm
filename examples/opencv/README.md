# Distributed Image Processing with OpenCV and Pachyderm

A detailed walkthrough of this example is included in our docs
[here](http://docs.pachyderm.com/latest/getting_started/beginner_tutorial.html).
Please follow that guide to deploy this pipeline.

## Building docker images

You shouldn't need to build docker images, as this example will pull our
pre-built ones. If for some reason you do need to, make sure to allocate 12gb
(or more), as compiling OpenCV will otherwise fail.
