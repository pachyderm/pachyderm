>![pach_logo](../img/pach_logo.svg) INFO - Pachyderm 2.0 introduces profound architectural changes to the product. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Branch Master: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples
# Distributed Image Processing with OpenCV and Pachyderm

A detailed walkthrough of this example is included in our docs [here](http://docs.pachyderm.com/latest/getting_started/beginner_tutorial.html). Please follow that guide to deploy this pipeline.

## Building docker images

You shouldn't need to build docker images, as this example will pull our pre-built ones. If for some reason you do need to, make sure to allocate 12gb (or more), as compiling OpenCV will otherwise fail.
