>![pach_logo](../img/pach_logo.svg) INFO Each new minor version of Pachyderm introduces profound architectual changes to the product. For this reason, our examples are kept in separate branches:
> - Branch Master: Examples using Pachyderm 2.1.x versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 2.0.x: Examples using Pachyderm 2.0.x versions - https://github.com/pachyderm/pachyderm/tree/2.0.x/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13.x versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples
# Distributed Image Processing with OpenCV and Pachyderm

A detailed walkthrough of this example is included in our docs [here](http://docs.pachyderm.com/latest/getting-started/beginner-tutorial.html). Please follow that guide to deploy this pipeline.

## Building docker images

You shouldn't need to build docker images, as this example will pull our pre-built ones. If for some reason you do need to, make sure to allocate 12gb (or more), as compiling OpenCV will otherwise fail.
