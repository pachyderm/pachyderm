# mnist with TFJob and Pachyderm

This example uses the canonical mnist dataset, Kubeflow, TFJobs, and Pachyderm to demonstrate an end-to-end machine learning workflow with data provenance.

## Step 1 - Install and deploy both Kubeflow and Pachyderm
Part of what makes [Pachyderm](https://pachyderm.com/) and [Kubeflow](https://www.kubeflow.org/) work so well together is that they're built on [Kubernetes](https://kubernetes.io/), which means they can run virtually anywhere. While both Pachyderm and Kubeflow have their own deployment instructions for various infrastructures, this instructional will be based on my personal favorite, GKE. Before continuing, make sure you have the following installed on your local machine. 

#### Prerequisites:
  - [Pachyderm CLI](https://docs.pachyderm.com/latest/getting_started/local_installation/#install-pachctl)
  - [Kubeflow CLI](https://www.kubeflow.org/docs/started/getting-started/#installing-command-line-tools)
  - [Kubectl CLI](https://kubernetes.io/docs/tasks/tools/install-kubectl/)  
  - [gcloud CLI](https://cloud.google.com/sdk/gcloud/) 
  - [Docker](https://docs.docker.com/install/)

#### Deploy:
To make it simple, we created a simple [bash script](github.com/pachyderm/pachyderm/examples/kubeflow/mnist_with_tfjob/gcp-kubeflow-pachyderm-setup.sh) specifically for this post, and you can use it to deploy Pachyderm and Kubeflow together on GKE in no time. However, If you prefer to do this all on your local machine, or any other infrastructure, please refer to the links below.

  - [Pachyderm Install Docs](https://docs.pachyderm.com/latest/getting_started/local_installation/)
  - [Kubeflow Install Docs](https://www.kubeflow.org/docs/started/getting-started/#installing-kubeflow)

#### Working setup check
1. `kubectl get pods -n kubeflow` returns running pods.
2. `pachctl version` returns *both* pachctl and pachd versions.
3. `pachctl enterprise get-state` returns: `Pachyderm Enterprise token state: ACTIVE`. If it doesn't, register Pachyderm as described in [Activate by Using the Dashboard](https://docs.pachyderm.com/latest/enterprise/deployment/#activate-by-using-the-dashboard).

## Step 2 - Checking in your data 
With everything configured and working, we're going to grab our data and then check it in to Pachyderm. To do so, you need to download a mnist.npz dataset to your local machine, create a Pachyderm repo, and then upload the dataset to the Pachyderm repo.

1. Download the `mnist.npz` file to a blank directory on your local machine:  

`➜ curl -O https://storage.googleapis.com/tensorflow/tf-keras-datasets/mnist.npz`

2. Create our two Pachyderm repos called `inputrepo` and `outputrepo`:  
`➜ pachctl create repo inputrepo`
`➜ pachctl create repo outputrepo`

3. Add `mnist.npz` to `inputrepo`:  
`pachctl put file inputrepo@master:/data/mnist.npz -f mnist.npz`

This command copies the minst dataset from your local machine to your Pachyderm repo `inputrepo` and Pachyderm will assign it a commit ID. Congratulations! Your data now has a HEAD commit, and Pachyderm has begun version-controlling the data!

4. Confirm that the data is checked-in running the following command:  
```sh
➜ pachctl list file inputrepo@master:/data/
NAME            TYPE SIZE     
/data/mnist.npz file 10.96MiB
```  

5. Next, we create a master branch on the outputrepo, so it's visible via the S3 Gateway
```sh
pachctl create branch outputrepo@master
```

## Step 3 - Deploying code to work with MNIST
Now that our data is checked in, it's time to deploy some code. 
Before we actually deploy it,
let's take a look at how things are working under the hood. 
### Understanding the code
Below is a snippet of `tfjob.py`

```python
# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket  = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir  = os.getenv('DATA_DIR', "/tmp/data")
# this is the training data file in the input repo
training_data = os.getenv('TRAINING_DATA', "mninst.npz")
# this is the name of model file in the output repo
model_file = os.getenv('MODEL_FILE', "my_model.h5")

def main(_):
  input_url = 's3://' + args.inputbucket + "/"
  output_url = 's3://' + args.outputbucket + "/"
  
  os.makedirs(args.datadir)

  # first, we copy files from pachyderm into a convenient
  # local directory for processing.  
  input_uri = os.path.join(input_url, args.trainingdata)
  training_data_path = os.path.join(args.datadir, args.trainingdata)
  print("copying {} to {}".format(input_uri, training_data_path))
  file_io.copy(input_uri, training_data_path, True)
  
  (train_images, train_labels), (test_images, test_labels) = tf.keras.datasets.mnist.load_data(path=training_data_path)

  <<<< .... MNIST example below ..... >>>>
```
The comments in the code provide a pretty good description of what's going on line by line. 
However, a quick breakdown is this:

1. We're copying mnist.npz from our Pachyderm repo `inputrepo@master` via the [S3 Gateway](https://docs.pachyderm.com/latest/enterprise/s3gateway/) into a local directory in the container (`/tmp/data/`). 
2. Then we tell [TensorFlow](https://www.tensorflow.org/) to load that data and start training. 
3. Saving the trained model back to Pachyderm.
Once our code trains the model, it needs to save it somewhere. 
Just like we copied data into the container,
we can copy it back out again. 
And, 
thanks to Pachyderm, 
we can maintain some sense of lineage. 
If you take a look at the `tfjob_mist.py` and scroll towards the bottom,
you'll see that the code is just copying the `my_model.h5` to the Pachyderm S3 Gateway `outputrepo` via `url:s3://<pachyderminstance>/master.outputrepo:/data/`
```python
# Save entire model to a HDF5 file
  model_file =  os.path.join(args.datadir,args.modelfile)
  model.save(model_file)
  # Copy file over to Pachyderm
  output_uri = os.path.join(output_url,args.modelfile)
  print("copying {} to {}".format(model_file, output_uri))
  file_io.copy(model_file, output_uri, True)
```

That takes care of the code. 

### Deploying the code
To deploy the code, we first download the manifest that will run it in Kubeflow and then deploy that manifest.
Before we deploy, 
we'll look at the manifest so you can understand what it's doing.

1. In the same directory as [step 2, above](#step-2---checking-in-your-data) run the following:
```sh
git clone https://github.com/pachyderm/pachyderm.git && cd pachyderm/examples/kubeflow/mnist-tfjob
```

2. Next, let's move onto how we'll use Kubeflow to deploy our code. 
Start by taking a look at the `tf_job_s3_gateway.yaml`:

```yaml
apiVersion: "kubeflow.org/v1beta2"
kind: "TFJob"
metadata:
  name: "mnist-pach-s3-gateway-example2"
  namespace: kubeflow 
spec:
  cleanPodPolicy: None 
  tfReplicaSpecs:
    Worker:
      replicas: 1 
      restartPolicy: Never
      template:
        spec:
          containers:
            - name: tensorflow
              image: pachyderm/mnist_klflow_example:v1.3
              env:
                # This endpoint assumes that the pachd service was deployed
                # in the namespace pachyderm.
                # You may replace this with pachd.<namespace> if you deployed
                # pachyderm in another namespace. For example, if deployed
                # in default it would be pachd.default. You may also
                # hard code in the pachd CLUSTER-IP address you obtain from
                # kubectl get services -n <namespace>
                - name: S3_ENDPOINT
                  value: "pachd.pachyderm:600"
                - name: S3_USE_HTTPS
                  value: "0"
                - name: S3_VERIFY_SSL
                  value: "0"
                - name: S3_REQUEST_TIMEOUT_MSEC
                  value: "600000"
                - name: S3_CONNECT_TIMEOUT_MSEC
                  value: "600000"
                - name: S3_REGION
                  value: "us-east1-b"
              command:
                - "python3"
                - "/app/tfjob.py"
```
The `mnist_tf_job_s3_gateway.yaml` is our spec file 
that [Kubeflow](https://www.kubeflow.org/) and [Kubernetes](https://kubernetes.io/) will use to deploy our code. 
You can find out everything you need to know about this spec file in the [Kubeflow TFjobs Docs](https://www.kubeflow.org/docs/components/training/tftraining/). 
Notice the Pachyderm repos, branches, and data/model locations are being declared at the bottom.

3. To deploy just run the following:
```sh
kubectl create -f mnist_tf_job_s3_gateway.yaml
```

## Step 4 - Monitoring our TFJob

We can check on things by going to the and click on TFJob Dasboard. If you deployed Pachyderm and Kubeflow from our sample deployment script, you can run `open $KF_URL` from macOS or `xdg_open $KF_URL` from Linux to get to the Kubeflow dashboard. You should see something like:

![tfjob-dashboard](tfjob-dashboard.png)

## Step 5 - Trust but verify: data versioning in Kubeflow with Pachyderm
You know the old saying,  “always trust but verify”. Let's confirm that you actually trained your model and that data provenance was maintained as the data worked its way through.

```
➜ pachctl list file outputrepo@master
NAME         TYPE SIZE     
/my_model.h5 file 4.684MiB 
```
Perfect, the data is exactly where it should be. Now, let’s take a mental snap-shot of how things look from a Pachyderm perspective, because things are about to get really interesting. Simply run the following:

```
➜ pachctl list file inputrepo@master --history 1
COMMIT                           NAME  TYPE COMMITTED         SIZE     
3fa46b65d4ce4ff8b9d50068a3bc2ada /data dir  2 minutes ago     10.96MiB
``` 

```
➜ pachctl list file outputrepo@master --history 1
COMMIT                           NAME         TYPE COMMITTED         SIZE     
5977593e8c604f158f4d80f42c8233ef /my_model.h5 file 2 minutes ago     4.684MiB
```

## Step 6 - Adding lineage: a basic model for using Pachyderm's data lineage in Kubeflow
To demonstrate data lineage, we need to create two versions of something. For the sake of time, and the fact that there's only one handwriting mnist dataset, you're just going to reupload the `mnist.npz`. You and I are just going to pretend that it's a new batch of handwriting data and that the model needs to be retrained. 

1. Rename the tfjob in our `mnist_tf_job_s3_gateway.yaml` with something like:
`name: "mnist-pach-s3-gateway-example2"`

2. Repeat checking in the `mnist.npz` dataset from step 2 and start a new job

`pachctl put file inputrepo@master:/data/mnist.npz -f mnist.npz && kubectl create -f mnist_tf_job_s3_gateway.yaml`

3. Once that's complete (should be pretty quick), you can then take a look at the lineage by running:

```
➜ pachctl list commit inputrepo@master 
REPO      BRANCH COMMIT                           PARENT                           STARTED     DURATION           SIZE     
inputrepo master 35080bc2a2504878a48f81e43117711c 3fa46b65d4ce4ff8b9d50068a3bc2ada 2 hours ago Less than a second 10.96MiB 
inputrepo master 3fa46b65d4ce4ff8b9d50068a3bc2ada <none>                           2 hours ago Less than a second 10.96MiB
```
and  
```
➜ pachctl list commit outputrepo@master
REPO       BRANCH COMMIT                           PARENT                           STARTED     DURATION           SIZE     
outputrepo master 5977593e8c604f158f4d80f42c8233ef c02cccded8ae434c840f25c5835dc535 2 hours ago Less than a second 4.684MiB 
outputrepo master c02cccded8ae434c840f25c5835dc535 <none>                           2 hours ago Less than a second 4.684MiB 
```

When you re-trained your model after _receiving new mnist data_ (remember we're pretending), 
the new model, `my_model.h5`, got copied to the `outputrepo` via the S3 Gateway, 
and Pachyderm assigned it a new commit-id
whose `PARENT` is the previous commit. 
You can feel free to repeat this step as many times as you'd like,
but the end-result will always be the same: version-controlled data.

Now when anyone asks "What data was used to train that model?",
you can tell them with just one command. 
And for when that auditor asks, 
"What data was used to train that model 3 months ago?"
Well, that's just one command away too.

### Going further with data lineage and data provenance
Great work! You started down the road to better data control and laid the groundwork for mastering data lineage.  

Of course, that’s just the start, there's more work to be done. A [true data lineage](https://www.pachyderm.io/dsbor.html) solution gives users a complete understanding of the entire journey of data, model, and code from top to bottom. Everything gets versioned and tracked as it changes, including the relationships between all three of those key pieces of every data science project.  

What we did here was take our first few steps by introducing version-control for the data alone. Don't worry though, Pachyderm and the Kubeflow community are on it, and we're collaborating to create the best possible solution for every AI infrastructure team to get a handle on their pipelines from start to finish.  

And for those of you wondering, "How would this work in a pipeline?" don't worry, we've got a separate post on just that so stay tuned!  

If you're interested in exploring data lineage and more, please go to [Pachyderm.com](https://pachyderm.com) or check us out on [GitHub](https://github.com/pachyderm/pachyderm) and be sure to join our [Slack](http://slack.pachyderm.io/) if you need help getting going fast.
