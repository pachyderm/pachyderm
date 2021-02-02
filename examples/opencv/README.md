# Getting started with Pachyderm - Image Processing with OpenCV

>![pach_logo](./img/pach_logo.svg) This example walks you through everything you need to know to get started with Pachyderm. We recommend reading through it all attentively. 

***Key concepts***

In particular, it will cover the following concepts:

- Find a list of all our [pachctl](https://docs.pachyderm.com/latest/reference/pachctl/pachctl/) commands here - pachctl is our CLI and a straightforward way to interact with pachyderm (pachd) on your cluster. Newcomer? This [autocompletion tool](https://docs.pachyderm.com/latest/deploy-manage/manage/pachctl_shell/) is for you.
- [Repositories](https://docs.pachyderm.com/latest/concepts/data-concepts/repo/), the location where you store your data inside Pachyderm.
- [Pipelines](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/pipeline/) responsible for reading data from a specified source, such as a Pachyderm repo, transforming it according to the [pipeline configuration](https://docs.pachyderm.com/latest/reference/pipeline_spec/), and writing the result to an output repo. Follow this [working with pipelines how-to](https://docs.pachyderm.com/latest/how-tos/developer-workflow/working-with-pipelines/) for more sequenced details on how to proceed.
- Pachyderm enables you to combine multiple input repositories in a single pipeline by using [cross inputs](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/cross-union/) in the pipeline specification. [See also](https://docs.pachyderm.com/latest/reference/pipeline_spec/#cross-input). Curious about the various types of inputs that combine multiple repositories? [Check this out](https://docs.pachyderm.com/latest/concepts/pipeline-concepts/datum/).



## 1. Getting ready
***Prerequisite***
- A workspace on [Pachyderm Hub](https://docs.pachyderm.com/latest/pachhub/pachhub_getting_started/) (recommended) or Pachyderm running [locally](https://docs.pachyderm.com/latest/getting_started/local_installation/).
- [pachctl command-line ](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl) installed, and your context created (i.e. you are logged in)

***Getting started***
- Clone this repo.
- Make sure Pachyderm is running. You should be able to connect to your Pachyderm cluster via the `pachctl` CLI. 
Run a quick:
```shell
$ pachctl version

COMPONENT           VERSION
pachctl             1.12.0
pachd               1.12.0
```
Ideally, have your pachctl and pachd versions match. At a minimum, you should always use the same major & minor versions of pachctl and pachd. 
## 2. Example walkthrough

A detailed walkthrough of this example is included in our docs [here](http://docs.pachyderm.com/latest/getting_started/beginner_tutorial.html). Please follow along.

>![pach_logo](./img/pach_logo.svg) You shouldn't need to build docker images for this example. The code and all necessary libraries have already been included in our pre-built image and pushed to Docker Hub. However, if you decide to modify the source code provided and want to re-build it, check this [How-To](https://docs.pachyderm.com/latest/how-tos/developer-workflow/working-with-pipelines/#step-3-push-your-docker-image-to-a-registry) section or take a look at this [build_tag_deploy.sh](https://github.com/pachyderm/pachyderm/blob/master/examples/joins/build_tag_deploy.sh) script in another of our examples. Make sure to allocate 12Gb (or more), as compiling OpenCV will otherwise fail.



