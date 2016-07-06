# Game of Thrones / Tensor Flow Example

Adapted from the tensor flow LSTM example [here](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#recurrent-neural-networks)

## Overview

In this example, you'll generate a new Game of Thrones script based on a bunch of previous GoT scripts.

To do so, we'll be adapting [this LSTM Neural Net example](https://www.tensorflow.org/versions/r0.8/tutorials/recurrent/index.html#recurrent-neural-networks) from Tensor Flow. We won't cover any LSTM or Neural Net theory in this example. For background we recommend reading that example and the resources they link to.

This guide assumes you already have a [working pachyderm cluster](../../SETUP.md), and you have a basic grasp of Pachyderm repos and pipelines. If you don't, you may want to start w the [fruit stand](../fruit-stand/README.md) example.

## How

Getting this neural net running on Pachyderm will require a few steps:

1. Creating the data repo, and initializing it with some GoT scripts
2. Creating the Docker image that includes the TF library and our code
3. Creating the Pachyderm Pipeline that trains the neural net, and generates the new script

---

### Initializing the Data

Since this data set isn't tiny, we've included some helpers to create the repos we need and input the data. To initialize the data set, just run:

```shell
make input-data 
```

This task does 2 things:

1. It grabs the data set in the form of a tarball from a URL, and extracts the data
2. It inputs this data into pachyderm by:
    - creating a new repo `GoT_scripts`
    - starting a commit
    - mounting the Pachyderm File System at `./mnt`
    - copying the data over to the `./mnt/GoT_scripts/{commitID}/` path
    - finishing the commit

The result is a new repo w all the data we need stored inside. To confirm the setup, you can do:

```shell
pachctl list-repo
# you should see 'GoT_scripts' w a nonzero size
pachctl list-commit
# you should see a single commit
```

---

### Creating the Transformation Image

Using Tensor Flow with Pachyderm is easy! Since Pachyderm Processing System (PPS) allows you to use any Docker image, getting the code in place is straightforward. In fact, since Tensor Flow provides a docker image to work with, they've done most of the work for us!

To construct the image, we need to:

1. Make sure we use an image w the Tensor Flow library installed
2. Make sure the image includes our code
3. Update the image to include Pachyderm's job shim
4. Actually compile the image

If you take a look at the [Dockerfile](./Dockerfile) in this directory, you'll notice a couple things.

1) The top line specifies that we're basing our image off of Tensor Flow's image:

```
FROM tensorflow/tensorflow 
```

2) On line 25 we're including the [code found locally](./code) in the Docker image:

```
ADD code /code
```

3) The rest of the Dockerfile installs the dependencies that Pachyderm needs to run the transformation. 

This boils down to: 

- installing FUSE so that PFS can expose the data to the container
- installing `job-shim` which is the binary that actually runs the transformation

4) Generate the image

Again, we have a helper. If you run:

```
make docker-build
```

It will compile the Docker image above and name it `tensor_flow_rnn_got`

#### A few things to note:

This represents one of two ways to construct a custom image to run on pachyderm. In this example, we use the `FROM` directive to base our image off of a 3rd party image of our choosing. This works well - we just need to add the dependencies that pachyderm needs to run. This is the more common use case when constructing custom images with complicated dependencies -- you probably have an image in mind that has the dependencies you need.

The alternative is using the `FROM` directive to base the image off of Pachyderm's standard image. [You can see an example here](./TODO). This usage pattern is more helpful for simple dependencies, or if you don't already have a Docker image you need to use.

If you're familiar w Dockerfiles, one thing that seems noticeably absent is an `ENTRYPOINT` command to tell the image what to run. In Pachyderm Processing System you specify this when creating a pipeline, as we'll see in a moment.


---

### Creating the Pipeline

Now that we have the data and the image ready, we can specify how we want to process the data. To do that, we use a pipeline manifest. Take a look at `pipeline.json` in this directory. Here you'll see a lot of information.

#### Training:

We've specified two pipelines here. The first is `GoT_train`, which represents the processing needed to train the neural net on the input data. You can see that it takes `GoT_scripts` as its input.

Now note that we've specified the `tensor_flow_rnn_got` image, and we specify an entry point by specifying the command and the stdin we pass. In our case, the meat boils down to this line:

```
cd /code && python ptb_word_lm.py --data_path=/data --model=small --model_path_prefix=/data > /pfs/out/model-results
```

This line tells Pachyderm Processing System (PPS) how to run the job. Here you can see we specify how to run the code w the arguments we want.

You'll also notice we have a few different kinds of output. The first one is we redirect the stdout of running the script to a logfile. That's handy if we want to see some of the debug output of the script being run. We will find that useful in a moment.

The other outputs are dictated by the script. In this case its really two types of files -- the training model weights output into a Tensor Flow `checkpoint` file, and two `.json` files containing a map of ID -> word and vice versa. These files will be output to the repo matching the pipeline name: `GoT_train`. We'll use these files in the next pipeline.


#### Generating:

This pipeline uses the model from the training stage to generate new text. It needs to consume the trained model as input, as well as the ID -> word and word -> ID json maps to output human readable text. To enable this, we specify the `GoT_train` repo as our input.

Notice that we specify the same image as above, just a different entrypoint via the `transform` property of the manifest. In this step, we're outputting the results to a file called `new_script.txt`


#### Running:

Now that we have all of the pieces, we can run the pipeline. Do so by running:

`make run`

This creates a pipeline from `pipeline.json`, and since a commit exists on the very first input repo `GoT_scripts`, the pipeline runs automatically. To see what's happening, you can run:

`pachctl list-job`

and 

`pachctl list-commit GoT_generate`

The first command will show you the current status of the jobs. In our case, we expect the training pipeline to run for much longer than the generate step.

#### Results:

By default, we've set the size of the model to train to `test`. You can see this on each of the entrypoints specified in the `pipeline.json` file. For this size model, the training pipeline should run for a couple minutes, and the generate step much quicker than that. Let's take a look at the output.

Once `pachctl list-commit GoT_generate` shows a single commit, we can take a look at the output. To do so, you can run:

```
pachctl get-file GoT_generate {the commit id from the above command} new_script.txt
```

And you should see some text!

Keep in mind, the model we just trained was very simplistic. Doing the 'test' model suffices as a proof of concept that we can train and use a neural net using Tensor Flow on Pachyderm. That said, the test model is dumb, and the output won't be that readable.

Actually, you can see how 'dumb' the model is. If you [read through the Tensor Flow example](todo) they describe how 'perplexity' is used to measure how good this model will perform. Let's look at the perplexity of your model.

```
pachctl list-commit GoT_trian
pachctl get-file GoT_train {commit from command above} model_results
```

Remember, this file is the stdout from the python script while its training the model. It outputs the 'perplexity' at standard intervals on the training set, as well as outputs the 'perplexity' for the validation and test sets. You should see a perplexity of about XXX

That's not great. As a next step, you can improve that measure, and the readability of the output script!

#### Next Iteration

As they reference in their example, a 'perplexity' of less than 100 starts to get pretty good / readable. Running the 'small' model on the GoT data set should produce a perplexity lower than 100.

To do this, you'll need to tear down this pipeline and re-create it. Specifically:

1) Delete the existing data / pipeline by running:

```
pachctl delete-all
```

2) Change the entrypoint commands in `pipeline.json` under the transformation property to use the `small` model not the `test` model. Save your changes.

3) Re-initialize the data / compile the image / create the pipeline

Now you can do this in one step:

```
make all
```

The small model runs in about an hour. Once its complete, you can look at the output again. This time, the perplexity should be south of 100 and the output script should be semi-readable.


