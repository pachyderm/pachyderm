# Creating Analysis Pipelines
There are three steps to running an analysis in a Pachyderm "pipeline":

1. Write your code.
2. Build a [Docker](https://docs.docker.com/engine/getstarted/step_four/) image that includes your code and dependencies.
3. Create a Pachyderm "pipeline" referencing that Docker image.

Multi-stage pipelines (e.g., parsing -> modeling -> output) can be created by repeating these three steps to build up a graph of processing steps.

## 1. Writing your analysis code

Code used to process data in Pachyderm can be written using any languages or
libraries you want. It can be as simple as a bash command or as complicated as
a TensorFlow neural network.  At the end of the day, all your code and
dependencies will be built into a container that can run anywhere (including
inside of Pachyderm). We've got demonstrative [examples on
GitHub](https://github.com/pachyderm/pachyderm/tree/master/examples) using
bash, Python, TensorFlow, and OpenCV and we're constantly adding more.

As we touch on briefly in the [beginner
tutorial](../getting_started/beginner_tutorial.html), your code itself only
needs to read and write files from a local file system. It does NOT have to
import any special Pachyderm functionality or libraries.  You just need to be
able to read files and write files.

For the reading files part, Pachyderm automatically mounts each input data
repository as `/pfs/<repo_name>` in the running instances of your Docker image
(called "containers"). The code that you write just needs to read input data
from this directory, just like in any other file system.  Your analysis code
also does NOT have to deal with data sharding or parallelization as Pachyderm
will automatically shard the input data across parallel containers. For
example, if you've got four containers running your Python code, Pachyderm will
automatically supply 1/4 of the input data to `/pfs/<repo_name>` in each
running container. That being said, you also have a lot of control over how
that input data is split across containers. Check out our guide on [parallelism
and distributed computing](distributed_computing.html) for more details on that
subject.

For the writing files part (saving results, etc.), your code simply needs to
write to `/pfs/out`. This is a special directory mounted by Pachyderm in all of
your running containers. Similar to reading data, your code doesn't have to
manage parallelization or sharding, just write data to `/pfs/out` and Pachyderm
will make sure it all ends up in the correct place.

## 2. Building a Docker Image

When you create a Pachyderm pipeline (which will be discussed next), you need
to specify a Docker image including the code or binary you want to run.  Please
refer to the [official
documentation](https://docs.docker.com/engine/tutorials/dockerimages/) to learn
how to build a Docker images.

Note: You specify what commands should run in the container in your
pipeline specification (see **Creating a Pipeline** below) rather than the
`CMD` field of your Dockerfile, and Pachyderm runs that command inside the
container during jobs rather than relying on Docker to run it. The reason is
that Pachyderm can't execute your code immediately when your container starts,
so it runs a shim process in your container instead, and then calls your
pipeline specification's `cmd` from there.

Unless Pachyderm is running on the same host that you used to build your image,
you'll need to use a public or private registry to get your image into the
Pachyderm cluster.  One (free) option is to use Docker's DockerHub registry.
You can refer to the [official
documentation](https://docs.docker.com/engine/tutorials/dockerimages/#/push-an-image-to-docker-hub)
to learn how to push your images to DockerHub. That being said, you are more
than welcome to use any other public or private Docker registry.

Note, it is best practice to uniquely tag your Docker images with something
other than `:latest`.  This allows you to track which Docker images were used
to process which data, and will help you as you update your pipelines.  You can
also utilize the `--push-images` flag on `update-pipeline` to help you tag your
images as they are updated.  See the [updating pipelines
docs](updating_pipelines.html) for more information.

## 3. Creating a Pipeline

Now that you've got your code and image built, the final step is to tell
Pachyderm to run the code in your image on certain input data.  To do this, you
need to supply Pachyderm with a JSON pipeline specification. There are four
main components to a pipeline specification: name, transform, parallelism and
input. Detailed explanations of the specification parameters and how they work
can be found in the [pipeline specification
docs](../reference/pipeline_spec.html).

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
      "pfs": {
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

Creating a pipeline tells Pachyderm to run the `cmd` (i.e., your code) in your
`image` on the data in the HEAD (most recent) commit of the input repo(s) as
well as *all future commits* to the input repo(s). You can think of this
pipeline as being "subscribed" to any new commits that are made on any of its
input repos. It will automatically process the new data as it comes in.

As soon as you create your pipeline, Pachyderm will launch worker pods on
Kubernetes. These worker pods will remain up and running, such that they are
ready to process any data committed to their input repos. This allows the
pipeline to immediately respond to new data when it's committed without having
to wait for their pods to "spin up". However, this has the downside that pods
will consume resources even while there's no data to process. You can trade-off
the other way by setting the `standby` field to true in your pipeline spec.
With this field set, the pipelines will "spin down" when there is no data to
process, which means they will consume no resources. However, when new data
does come in, the pipeline pods will need to spin back up, which introduces some
extra latency. Generally speaking, you should default to not setting standby
until cluster utilization becomes a concern. When it does, pipelines that
run infrequently and are highly parallel are the best candidates for `standby`.
