# Blog Section
What the few steps will show you is how Pachyderm and Kubeflow can come together to version-control data for distributed tensorflow jobs on Kubeflow. Lets get started.

In a perfect world, you should be able to instantly reproduce any given result or step in your data science pipeline. I'm going to show you how Pachyderm and Kubeflow can lay that foundation for you and as Pachyderm and Kubeflow come closer together, that foundation will only get stronger. 

## Step 1 - Install and deploy both Kubeflow and Pachyderm
Part of what makes Pachyderm and Kubeflow so amazing is that they both can be deployed virtually anywhere. While both of the platforms have instructions on how to deploy them on your choice of infrastructures, this instructional will be based on my personal favorite, GKE.  

### Install/Prerequisits:
  - [Pachyderm cli](http://docs.pachyderm.com/en/latest/getting_started/local_installation.html#install-pachctl). 
  - [Kubeflow cli](https://www.kubeflow.org/docs/started/getting-started/#installing-command-line-tools).  
  - [gcloud cli](https://cloud.google.com/sdk/gcloud/). 
  - [docker](https://docs.docker.com/install/).

### Deploy:
For the lazy, we have a simple bash script you can use to deploy Pachyderm and Kubeflow together, on GKE. Just keep in mind you will be billed for the resource-usage on Google Cloud Platform. If you prefer to do this all on your local machine or any other infrastructure, please refer to the links below. 

  - [Pachyderm Install Docs](http://docs.pachyderm.com/en/latest/getting_started/local_installation.html)
  - [Kubeflow Install Docs](https://www.kubeflow.org/docs/started/getting-started/#installing-kubeflow)

No matter which route you take you should end up with both Pachyderm and Kubeflow up and running ontop of Kubernetes. 


## Step 2 - Establish some core concepts: 
Here's what we're going to deploy:

<<diagram showing arch diagram>> 

Lets quickly go over the compoents:
  - Pachyderm is going to version-control the data and serve it up to Kubeflows tf-job via a s3-like api called the S3 Gateway. 
  - Kubeflow is going to run our ML code (via a docker image) on Kubernetes where it can be scaled/distributed if necessary. 
  - Our code is going to copy the data from the Pachyderm S3 Gateway into a local directory and then we'll train a model using that data, and save the model back to a pachyderm repo. 


## Step 3 - Steps & Code (Warning: Mnist example ahead) 
Just like every other data science project, we have to get the data. But before we do that, Let's setup a Pachyderm repo so we have a repo to check our data into:  

`pachctl create repo inputrepo`

And while we're here might as well create our outputrepo. 

`pachctl create repo outputrepo`

With our repos set, lets get the data and _check it in_. In a empty directory on your machine run:

`curl -O https://storage.googleapis.com/tensorflow/tf-keras-datasets/mnist.npz`

and then check it into Pachyderm: 

`pachctl put file inputrepo@master:/data/mnist.npz -f mnist.npz`

This will copy the file into the Pachyderm repo where it will get a commit id. Congrats, your data now has a HEAD commit. You can verify that with:

`pachctl list file outputrepo@master:/data/ --history all` _(the `--history` flag tells pachyderm to show commit information.)_

```
➜ pachctl list file inputrepo@master:/data/ --history all
COMMIT                           NAME            TYPE COMMITTED    SIZE     
a1a45ecc348a4e41bfddb2ce32df1475 /data/mnist.npz file 1 minute ago 10.96MiB
```

Now that our data is checked in, it's time to deploy some code. In the same directory run the following:

`git clone https://github.com/pachyderm/pachyderm.git && cd pachyderm/examples/kubeflow/tfjob-example/`

Next up is the code, particularly this part:

```
# this is the Pachyderm repo & branch we'll copy files from
input_bucket = os.getenv('INPUT_BUCKET', 'master.inputrepo')
# this is the Pachyderm repo & branch  we'll copy the files to
output_bucket  = os.getenv('OUTPUT_BUCKET', "master.outputrepo")
# this is local directory we'll copy the files to
data_dir  = os.getenv('DATA_DIR', "/data")

# Getting the Pachyderm stuff stared
#input_url = 's3://' + args.inputbucket + "/data/"
#output_url = 's3://' + args.outputbucket + "/data/"
def main(_):
  input_url = 's3://' + args.inputbucket + "/data/"
  output_url = 's3://' + args.outputbucket + "/data/"
  
  os.makedirs(args.datadir)

  # first, we copy files from pachyderm into a convenient
  # local directory for processing.  The files have been
  # placed into the inputpath directory in the s3path bucket.
  print("walking {} for copying files".format(input_url))
  for dirpath, dirs, files in file_io.walk(input_url, True):
    for file in files:
      uri = os.path.join(dirpath, file)
      newpath = os.path.join(args.datadir, file)
      print("copying {} to {}".format(uri, newpath))
      file_io.copy(uri, newpath, True)

  path = '/tmp/data/mnist.npz'
  (train_images, train_labels), (test_images, test_labels) = tf.keras.datasets.mnist.load_data(path)

  <<<< .... MNIST example below ..... >>>>
```
"The what" we're doing here is simply copying data from the Pachyderm repo via the S3 Gateway into a local directory in the container (`/tmp/data/` to be exact). "The how" is much more interesting, repos in Pachyderm have brances; very similar to code. In our case, the mnist data we checked in earlier is at `master`. Therefore, our `s3://` url is going to be something like `s3://<pachydermurl>:master.inputrepo:/data/`. Then all we're doing is copying everything in the `/data/` directory to a local directory on the container (`/tmp/data/`), so tensorflow and tfjob can operate on it. Finally, we're telling `tf.keras.datasets.mnist.load_data` to load the data at the specific path of `/tmp/data/mnist.npz`. 

And just like we copied data into the container via the Pachyderm S3 gateway, we can copy it back out. If you take a look at the `tfjob_mist.py` and scroll towards the bottom you'll see that we're just crawling the same `/tmp/data/` directory and copying it to the Pachyderm S3 Gateway output url:`s3://<pachyderminstance>/master.outputrepo:/data/` 

```
print("walking {} for copying to {}".format(args.datadir, output_url))
  for dirpath, dirs, files in os.walk(args.datadir, topdown=True):   
    for file in files:
      uri = os.path.join(dirpath, file)
      newpath = output_url + file
      print("copying {} to {}".format(uri, newpath))
      file_io.copy(uri, newpath, True)
```

Now, lets deploy the code so we can train our model. Start by taking a look at the `tf_job_s3_gateway.yaml`:

```
apiVersion: "kubeflow.org/v1beta2"
kind: "TFJob"
metadata:
  name: "s3-gateway-example"
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
              image: nickharvey/slim-tfjob:v52
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
              command:
                - "python3"
                - "tfjob.py"
                - "-i"
                - "master.inputrepo"
                - "-o"
                - "master.outputrepo"
                - "-d"
                - "/tmp/data"
```
This is basically a set of instructions for kubeflow and kubernetes to take our code which is packaged in a docker container called `nickharvey/kubeflow-tfjob` (Dockerfile is also in the repo should you want to build your own). Notice that we're telling kubeflow and ultimately kubernetes, the Pachyderm S3 Gateway and information. Lastly, we're specfiying what Python program we want to run and how to run it.

To deploy just run the following:
`kubectl create -f tf_job_Pachyderm.yaml`

That's it! The data is flowing and the model is training.

We can check on things by going to the kubeflow UI and click on TFJob Dasboard. You should see something like:

<<.... Screenshot of dashboard ....>

Once our job is completed we should have a model the `outputrepo` we setup earlier. Trust but verify is what they always say, so lets run: 

```
➜ pachctl list file outputrepo@master:/data/ --history all
COMMIT                           NAME              TYPE COMMITTED   SIZE     
a0f654c7a65f42f69e8bddd1a2035a7e /data/mnist.npz   file 7 hours ago 10.96MiB 
eb7e51147b4d42cfbcd555c84be778d2 /data/my_model.h5 file 7 hours ago 4.684MiB 
```

Perfect, the data is exactly where it's supposed to be. Now, lets see how Pachyderm can show us where it came from. Simply run the following:

```
➜ pachctl inspect commit outputrepo@eb7e51147b4d42cfbcd555c84be778d2
Commit: outputrepo@eb7e51147b4d42cfbcd555c84be778d2
Original Branch: master
Parent: a0f654c7a65f42f69e8bddd1a2035a7e
Started: 7 hours ago
Finished: 7 hours ago
Size: 15.64MiB
```

How cool is that, our `my_model.h5` (commit: eb7e51147b4d42cfbcd555c84be778d2) has a _parent_ of `a0f654c7a65f42f69e8bddd1a2035a7e` which is the commit id of `mnist.npz` file we checked in earlier. Now when anyone asks, "What data was used to train that model?" With just one command you can tell them. 

### Wrap up! 
Fantastic! We were able to lay the foundations of a true data lineage solution using Pachyderm and Kubeflow in just a few short steps. Now, we know that the example above isn't a complete data lineage solution. We were able to version-control data, but true provenance is understanding the entire journey from start to finish. But don't worry, we're hard at work over here at Pachyderm working with the kubeflow community and customers to improve inter-operability. 

And for those of you who are wondering how would this work in a pipeline. Don't worry, we'll have a seperate post on the subject shortly.

If you're interested in exploring Pachyderm further please go to Pachyderm.com or check us out on [github](https://github.com/pachyderm/pachyderm)

