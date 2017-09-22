# Creating Analysis Pipelines
There are three steps to running an analysis in a Pachyderm "pipeline":

1. Write your code.
2. Build a [Docker](https://docs.docker.com/engine/getstarted/step_four/) image that includes your code and dependencies.
3. Create a Pachyderm "pipeline" referencing that Docker image.

Multi-stage pipelines (e.g., parsing -> modeling -> output) can be created by repeating these three steps to build up a graph of processing steps.

## 1. Writing your analysis code

Code used to process data in Pachyderm can be written using any languages or libraries you want. It can be as simple as a bash command or as complicated as a TensorFlow neural network.  At the end of the day, all your code and dependencies will be built into a container that can run anywhere (including inside of Pachyderm). We've got demonstrative [examples on GitHub](https://github.com/pachyderm/pachyderm/tree/master/doc/examples) using bash, Python, TensorFlow, and OpenCV and we're constantly adding more.

As we touch on briefly in the [beginner tutorial](../getting_started/beginner_tutorial.html), your code itself only needs to read and write files from a local file system. It does NOT have to import any special Pachyderm functionality or libraries.  You just need to be able to read files and write files.

For the reading files part, Pachyderm automatically mounts each input data repository as `/pfs/<repo_name>` in the running instances of your Docker image (called "containers"). The code that you write just needs to read input data from this directory, just like in any other file system.  Your analysis code also does NOT have to deal with data sharding or parallelization as Pachyderm will automatically shard the input data across parallel containers. For example, if you've got four containers running your Python code, Pachyderm will automatically supply 1/4 of the input data to `/pfs/<repo_name>` in each running container. That being said, you also have a lot of control over how that input data is split across containers. Check out our guide on :doc: `parallelization` to see the details of that.

For the writing files part (saving results, etc.), your code simply needs to write to `/pfs/out`. This is a special directory mounted by Pachyderm in all of your running containers. Similar to reading data, your code doesn't have to manage parallelization or sharding, just write data to `/pfs/out` and Pachyderm will make sure it all ends up in the correct place. 

## 2. Building a Docker Image

When you create a Pachyderm pipeline (which will be discussed next), you need to specify a Docker image including the code or binary you want to run.  Please refer to the [official documentation](https://docs.docker.com/engine/tutorials/dockerimages/) to learn how to build a Docker images.  Note, your Docker image should NOT specifiy a `CMD`.  Rather, you specify what commands are to be run in the container when you create your pipeline.

Unless Pachyderm is running on the same host that you used to build your image, you'll need to use a public or private registry to get your image into the Pachyderm cluster.  One (free) option is to use Docker's DockerHub registry.  You can refer to the [official documentation](https://docs.docker.com/engine/tutorials/dockerimages/#/push-an-image-to-docker-hub) to learn how to push your images to DockerHub. That being said, you are more than welcome to use any other public or private Docker registry.

Note, it is best practice to uniquely tag your Docker images with something other than `:latest`.  This allows you to track which Docker images were used to process which data, and will help you as you update your pipelines.  You can also utilize the `--push-images` flag on `update-pipeline` to help you tag your images as they are updated.  See the [updating pipelines docs](updating_pipelines.html) for more information.

## 3. Creating a Pipeline

Now that you've got your code and image built, the final step is to tell Pachyderm to run the code in your image on certain input data.  To do this, you need to supply Pachyderm with a JSON pipeline specification. There are four main components to a pipeline specification: name, transform, parallelism and input. Detailed explanations of the specification parameters and how they work can be found in the [pipeline specification docs](../reference/pipeline_spec.html). 

Here's an example pipeline spec:
```json
{
  "pipeline": {
    "name": "wordcount"
  },
  "transform": {
    "image": "wordcount-image",
    "cmd": ["/binary", "/pfs/data", "/pfs/out"]
  },
  "input": {
      "atom": {
        "repo": "data",
        "glob": "/*"
      }
  }
}
```

After you create the JSON pipeline spec (and save it, e.g., as `your_pipeline.json`), you can create the pipeline in Pachyderm using `pachctl`:

```sh
$ pachctl create-pipeline -f your_pipeline.json
```

(`-f` can also take a URL if your JSON manifest is hosted on GitHub or elsewhere. Keeping pipeline specifications under version control is a great idea so you can track changes and seamlessly view or deploy older pipelines if needed.)

Creating a pipeline tells Pachyderm to run the `cmd` (i.e., your code) in your `image` on the data in *every* finished commit on the input repo(s) as well as *all future commits* to the input repo(s). You can think of this pipeline as being "subscribed" to any new commits that are made on any of its input repos. It will automatically process the new data as it comes in. 

**Note** - In Pachyderm 1.4+, as soon as you create your pipeline, Pachyderm will launch worker pods on Kubernetes, such that they are ready to process any data committed to their input repos.

