## Step 1 - Install and deploy Kubeflow and Pachyderm
Part of what makes [Pachyderm](https://pachyderm.com/) and [Kubeflow](https://www.kubeflow.org/) work so well together is that they’re built on [Kubernetes](https://kubernetes.io/), which means they can run virtually anywhere. While both have their own deployment instructions for various infrastructures, this instructional will be based on [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/). Before continuing, make sure following installed on your local machine:

### Prerequisites:
- [Pachyderm CLI (1.10 or above)](/getting-started/)
- [Kubeflow CLI](https://www.kubeflow.org/docs/started/getting-started/#installing-command-line-tools) 
- [Kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [gcloud CLI](https://cloud.google.com/sdk/gcloud/)
- [Docker](https://docs.docker.com/install/)

### Deploy:
To make it simple, we created a [bash script](https://github.com/pachyderm/pachyderm/tree/master/examples/kubeflow/mnist/gcp-kubeflow-pachyderm-setup.sh) specifically for this post and you can use it to deploy Pachyderm and Kubeflow together on GKE in no time. However, if you prefer to do this all on your local machine, or any other infrastructure, please refer to the links below:

- [Pachyderm Install Docs](http://docs.pachyderm.com/en/latest/getting_started/local_installation.html)
- [Kubeflow Install Docs](https://www.kubeflow.org/docs/started/getting-started/#installing-kubeflow)

Once everything is deployed the easiest way to connect to kubeflow is via port-forward:
`kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80`

### Working setup check: 
1. `kubectl get pods -n kubeflow` returns running pods.
2. `pachctl version` returns *both* pachctl and pachd versions.

## Step 2 - Checking in your data
With everything configured and working, it’s time to grab some data and then check it in to Pachyderm. To do so, download a mnist.npz dataset to your local machine, and proceed with checking it into Pachyderm. 

 Download the mnist.npz file to a blank directory on your local machine:

`curl -O https://storage.googleapis.com/tensorflow/tf-keras-datasets/mnist.npz`

2. Create a Pachyderm repo called `input-repo`:

`pachctl create repo input-repo`

3. We check-in our `mnist.npz` to `input-repo`.:

`pachctl put file input-repo@master:/mnist.npz -f mnist.npz`

This command copies the minst dataset from your local machine to the Pachyderm repo `input-repo` and Pachyderm will assign it a commit ID. Congratulations! Your data now has a HEAD commit, and Pachyderm has begun version-controlling the data!

Confirm that the data is checked-in by running the following command:

```
➜ pachctl list file input-repo@master
NAME            TYPE SIZE     
/mnist.npz file 10.96MiB
```

## Step 3 - Deploying code to work with MNIST
Now that the data is checked in, it’s time to deploy some code. In the same directory run the following:

`git clone https://github.com/pachyderm/pachyderm.git && cd pachyderm/examples/kubeflow/mnist/`

Next, let’s take a look at the Pachyderm pipeline spec file `pipeline.yaml` as this is the start of our explainable, repeatable, scalable mnist-pipeline.

```
pipeline:
  name: mnist
transform:
  image: pachyderm/mnist_pachyderm_pipeline:v1.0.1
  cmd: [ /app/pipeline.py ]
s3_out: true  # Must be set
input:
  pfs:
    name: input      # Name of the input bucket
    repo: input-repo # Pachyderm repo accessed via this bucket
    glob: /          # Must be exactly this
    s3: true         # Must be set
```

For the most part, it’s a standard Pachyderm pipeline spec. We have our usual transform step that declares a docker image, and tells it to run pipeline.py. Below that, is where you’ll see a few of the new S3 gateway  features being used, primarily, the `s3_out: true` and `s3: true`. The `s3_out: true` allows your pipeline code to write results out to an S3 gateway bucket instead of the typical pfs/out directory. Similarly, `s3: true` is what tells Pachyderm to mount the given input as an S3 gateway bucket.

Next, open up the pipeline.py file, and you’ll see that apart from a few Kubeflow-specific bits, lines 1-71 is a pretty standard MNIST training example using [Tensorflow](https://www.tensorflow.org/).

<script src="https://gist.github.com/Nick-Harvey/b353659e84e26d33b57a1ea9376ed27a.js"></script>

Kubeflow users will notice that from lines 73 down, we’re just declaring a Kubeflow pipeline (KFP) using the standard [Kubeflow Pipelines SDK](https://www.kubeflow.org/docs/pipelines/sdk/sdk-overview/). Then, On line 88, we call `create_run_from_pipeline_func` to run the KFP with a couple additional arguments which declare the S3 endpoints being provided by the Pachyderm S3 gateway. In our case, this will be the `input-repo` that contains our MNIST training data, and then the KFP will output the trained back out through the Pachyderm S3 Gateway to our output repo.

## Step 4 - Data Lineage in action
That takes care of the code. Next, let's move on deploying everything so we can train our model.

`pachctl create pipeline -f pipeline.yaml`

You can keep an eye on progress by either running `pachctl list job` or by looking at the Experiments tab in the Kubeflow Dashboard.

![Kubeflow Dashboard](/images/Releases/1.10/Kubeflow-Central-Dashboard.jpg)

Once the job (or “run” in kubeflow terms) is complete, you should see a model file in your Pachyderm mnist repo (created automatically when we ran the `pachctl create pipeline`). You can check yourself with:

`pachctl list file mnist@master`

```
➜ pachctl inspect commit mnist@master
Commit: mnist@a64bfc6da6714c23a44db5c984850db2
Original Branch: master
Parent: 60d23801e61f4b3c975b7f0be1f9f208
Started: 56 seconds ago
Finished: 24 seconds ago
Size: 4.684MiB
Provenance:  __spec__@bd9f2665811f42ea9b1e3a56dc70c0b1 (mnist)  input-repo@a8dbe24ea9c047129a170cae95aa292a (master)
```

Notice the Provenance line. That right there is proof you just version-controlled your data as well as the machine learning model that was created from it.

Because you incorporated data lineage into your workflow using Pachyderm, you can actually restore previous versions of your data and model. That’s incredible when you consider how often you have to answer the question “What exact data was used to train that model?” Thanks to Pachyderm, you can answer that with just one command. And when an auditor asks, “what data was used to train that model 3 months ago?” Well, that’s just one Pachyderm command away too.

## Epilogue - Data Lineage for Existing Kubeflow Pipelines

Some Kubeflow users might get the impression that they can only use Pachyderm if they rewrite their kubeflow pipelines to embed them in a Pachyderm pipeline. Not so! Pachyderm can run *preexisting* kubeflow pipelines, version their output, and track their data lineage as well. To illustrate, the alternative commands below show Pachyderm running an existing Kubeflow pipeline this way:

```
# Connect your local machine to kubeflow's pipeline API
kubectl -n kubeflow port-forward svc/ml-pipeline 41888:8888 &

# Create a kubeflow pipeline
# --------------------------
# Our pipeline.py script can also upload a standalone pipeline to an existing
# kubeflow cluster, like so:
./pipeline.py \
  --remote-host=http://localhost:41888 \
  --create-pipeline=mnist_pipeline

# Create a Pachyderm  pipeline to trigger the new 'mnist_pipeline'
pachctl create pipeline -f standalone_pipeline.yaml
```

As before, you'll see a pachyderm pipeline start, create a kubeflow run, and pass data to the newly-created kubeflow workers. Also as before, you'll see an output commit, containing the model parameters, with the input commit and trigger pipeline in its provenance:
```
$ pc list job
ID        PIPELINE      STARTED            DURATION   RESTART PROGRESS  DL UL STATE
f79...d76 trigger-mnist 16 seconds ago     15 seconds 0       1 + 0 / 1 0B 0B success

$ pc list repo
NAME          CREATED            SIZE (MASTER) DESCRIPTION
trigger-mnist 24 seconds ago     4.684MiB      Output repo for pipeline trigger-mnist.
input-repo    About a minute ago 10.96MiB
```

At the lowest level, the the `s3` and `s3_out` fields of our pipeline specs support this integration. If a kubeflow job can read and write data to/from an S3 bucket, then its data lineage can be tracked with Pachyderm.
