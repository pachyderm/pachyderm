# Creating Analysis Pipelines
There are three steps to writing analysis in Pachyderm. 

1. Write your code
2. Generate a Docker image with your code
3. Create/add your pipeline in Pachyderm


## Writing your analysis code

Analysis code in Pachyderm can be written using any languages or libraries you want. At the end of the day, all the dependencies and code will be built into a container so it can run anywhere. We've got demonstrative [examples on GitHub](https://github.com/pachyderm/pachyderm/tree/master/examples) using bash, Python, TensorFlow, and OpenCV and we're constantly adding more.

As we touch on briefly in the [beginner tutorial](../getting_started/beginner_tutorial), your code itself only needs to read and write data from the local file system. 

For reading, Pachyderm will automatically mount each input data repo as `/pfs/<repo_name>`. Your analysis code doesn't have to deal with distributed processing as Pachyderm will automatically shard the input data across parallel containers. For example, if you've got four containers running your Python code, `/pfs/<repo_name>` in each container will only have 1/4 of the data. You have a lot of control over how that input data is split across containers. Check out our guide on :doc: `parallelization` to see the details of that.

All writes simply need to go to `/pfs/out`. This is a special directory that is available in every container. As with reading, your code doesn't have to manage parallelization, just write data to `/pfs/out` and Pachyderm will make sure it all ends up in the correct place. 

## Building a Docker Image

As part of a pipeline, you need to specify a Docker image including the code you want to run.

There are two ways to construct the image. Both require some familiarity with [Dockerfiles](https://docs.docker.com/engine/tutorials/dockerimages/#/building-an-image-from-a-dockerfile).

In short, you need to include a bit of Pachyderm code, called the Job Shim, in your image, but a Dockerfile can only have a single `FROM` directive. Therefore, you can either add your code to our job-shim image or you can add our job-shim code to your own image. 

### Using the Pachyderm Job Shim

Use this method if your dependencies are simple. Just base your image off of Pachyderm's job shim image:

```
FROM pachyderm/job-shim:latest

```

[Here is an example](https://github.com/pachyderm/pachyderm/blob/master/examples/word_count/Dockerfile) where the transformation code was written in Go, so in addition to using the job-shim image, this Dockerfile installed go1.6.2 and compiled the program needed for the transformation. 

### Adding the Job-shim Code to Your Image

Use this method if your dependencies are pretty complex or you're using a published 3rd-party image such as [TensorFlow](https://github.com/pachyderm/pachyderm/blob/master/examples/tensor_flow/Dockerfile).

In this case, the `FROM` directive will specify the 3rd party image of your choice and then you'll add the Pachyderm code to your Dockerfile. Below is the code you need to add (you can also [view it on GitHub](https://github.com/pachyderm/pachyderm/blob/master/etc/user-job/Dockerfile)).


```
FROM `Your Image`

# then ...
# Install FUSE
RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    git \
    ca-certificates \
    curl \
    fuse && \
  apt-get clean && \
  rm -rf /var/lib/apt

# Install Go 1.6.0 (if you don't already have it in your base image)
RUN \
  curl -sSL https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
ENV GOPATH /go
ENV GOROOT /usr/local/go

# Install Pachyderm job-shim
RUN go get github.com/pachyderm/pachyderm && \
	go get github.com/pachyderm/pachyderm/src/server/cmd/job-shim && \
    cp $GOPATH/bin/job-shim /job-shim
```


## Creating a Pipeline

Now that you've got your code and image built, the final step is to add a pipeline manifest to Pachyderm. Pachdyerm pipelines are described using a JSON file. There are four main components to a pipeline: name, transform, parallelism and inputs. Detailed explanations of parameters and how they work can be found in the [pipeline_spec](./pipeline_spec.html). 

Here's a template pipeline:
```json
{
  "pipeline": {
    "name": "my-pipeline"
  },
  "transform": {
    "image": "my-image",
    "cmd": [ "my-binary", "arg1", "arg2"],
    "stdin": [
        "my-std-input"
    ]
  },
  "parallelism": "4",
  "inputs": [
    {
      "repo": {
        "name": "my-input"
      },
      "method": "map"
    }
  ]
}
```

After you create the JSON manifest, you can add the pipeline to Pachyderm using:

```
 $ pachctl create-pipeline -f your_pipeline.json
```
`-f` can also take a URL if your JSON manifest is hosted on GitHub for instance. Keeping pipeline manifests under version control too is a great idea so you can track changes and seamlessly view or deploy older pipelines if needed.

Creating a pipeline tells Pachyderm to run your code on *every* finished
commit in the input repo(s) as well as *all future commits* that happen after the pipeline is created. 

You can think of this pipeline as being "subscribed" to any new commits that are made on any of its input repos. It will automatically process the new data as it comes in. 



