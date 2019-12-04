#  Reading and writing to Pachyderm from Kubeflow TFJobs

## Introduction

Pachyderm makes production data pipelines repeatable, scalable and provable.
Data scientist and engineers use Pachyderm pipelines to connect data acquisition, cleaning, processing, modeling, and analysis code,
while using Pachyderm's versioned data repositories to keep a complete history of all the data, models, parameters and code
that went into producing each result anywhere in their pipelines. 
This is called data provenance.

If you're currently using Kubeflow to manage your machine learning workloads,
Pachyderm can add value to your Kubeflow deployment in a couple of ways.
You can use Pachyderm's pipelines and containers to call Kubeflow API's to connect and orchestrate Kubeflow jobs.
You can use Pachyderm's versioned data repositories to provide data provenance to the data, models and parameters you use with your Kubeflow code.

We'll address the latter in this example,
where we will use the Tensorflow `file_io` library
to read data from a Pachyderm versioned data repository
in a Kubeflow TFJob
using the Pachyderm S3 Gateway.
We'll then copy the data to another Pachyderm repo.
This is intended to be a simple example
that shows interoperability among the three frameworks.
Rather than use a complicated Tensorflow/Kubeflow example that shows distributed training,
this example will focus on the minimum needed to work with Pachyderm's S3 Gateway from a Kubeflow TFJob.

We will use the example Kubeflow [tfoperator example mnist_with_summaries](https://github.com/kubeflow/tf-operator/tree/master/examples/v1beta2/mnist_with_summaries) as a basis, so you can easily adapt the code for your needs.
For simplicity, we will use example data files from the mnist training set,
but you can easily store any data you wish.
Anywhere you use the Tensorflow `file_io` library,
you can access data, models, parameters, etc. that are stored in Pachyderm's versioned data repositories.

**NOTE**
It's best to test this connectivity first with authentication and SSL turned off in Pachyderm Enterprise Edition.
Once you've confirmed the basic functionality of reading and writing to Pachyderm repos from TFJobs in Kubeflow,
contact Pachyderm support to configure authentication and SSL for use with the S3 Gateway.

### Prerequisites

You should deploy Pachyderm and Kubeflow in your environment.
**NOTE**
If deploying to Google Cloud Services, it may be better to manually create the cluster using the instructions for deploying Pachyderm *first*, 
as it requires the flag `--scopes storage-rw` to the `gcloud container clusters create` command.
Modifying an existing cluster to add that scope is difficult. 
You can then configure Kubeflow to deploy without creating the cluster,
using the `--skip-init-gcp-project` to the `kfctl init` command.


## The Pachyderm S3 Gateway

One of the ways you can use Pachyderm's version-controlled data repositories with Kubeflow is via Pachyderm Enterprise Edition's S3 Gateway feature.
The S3 Gateway provides an way to access data stored in Pachyderm using any S3-compatible tool such as `s3cmd`, `mc` or the Kubeflow `file_io` library.

The S3 Gateway is enabled via the Enterprise Edition activation code
you received from Pachyderm.
[This document](https://docs.pachyderm.com/latest/enterprise/deployment/) takes you through activating Enterprise Edition
with that code 
in the web-based dashboard UI or 
via the command line.

Once Pachyderm Enterprise Edition has been activated, 
you can connect to Pachyderm's S3 gateway.

Complete instructions for setting up and working with the S3 Gateway using various S3 clients, 
[s3cmd](https://github.com/s3tools/s3cmd),
[the AWS CLI](https://aws.amazon.com/cli/) and
the [minio client (mc)](https://github.com/minio/mc) are available
at the [documentation for the S3 Gateway](https://docs.pachyderm.com/latest/enterprise/s3gateway/).

### Accessing Pachyderm repos from TFJobs and the Tensorflow apis

Once you have deployed both Pachyderm and Kubeflow,
enabled Pachyderm Enterprise Edition with your activation key,
and confirmed that you can access the S3 gateway,
you can easily create a TFJob
that can read from and write to Pachyderm's version-controlled repositories.

In this example, 
you will find code that uses the Kubeflow [tfoperator example mnist_with_summaries](https://github.com/kubeflow/tf-operator/tree/master/examples/v1beta2/mnist_with_summaries) images as its base and
does simple reads and writes to a Pachyderm repo.

The script `tf_job_s3_gateway.py` takes three flags:
* `-i` or `--inputbucket`, 
  which specifies the name of the bucket, 
  without the s3:// prefix, 
  that we'll copy the data from.
  For Pachyderm repos, 
  it'll be of the form `branch`.`<repo name>`
* `-d` or `--datadir`,
  which specifies the local directory that we'll copy the data into,
  so Tensorflow may operate on it.
* `-o` or `--outputbucket`,
  which specifies the name of the bucket, 
  without the s3:// prefix, 
  that we'll copy the results into. 
  For Pachyderm repos, 
  it'll be of the form `branch`.`<repo name>`

In `Dockerfile.tf_job_s3_gateway`, 
this script is built into a Docker image.

We'll use as source a subset of the mnist training data,
located in the mnist directory.

The relevant parts of the code from  `tf_job_s3_gateway.py` are below.
This code simply walks through a Pachyderm versioned data repository,
copying files to a directory using the Tensorflow `file_io` library.
The files can be operated on using Tensorflow functions.
The data is then copied to another Pachyderm versioned data repository.

```python
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


    # here is where you would apply your training to the data in args.datadir
    # it might operate on the data directly, or place additional
    # data in the same directory

    # finally, we copy the output from those operations to
    # another pachyderm repo
    print("walking {} for copying to {}".format(args.datadir, output_url))
    for dirpath, dirs, files in os.walk(args.datadir, topdown=True):   
      for file in files:
        uri = os.path.join(dirpath, file)
        newpath = output_url + file
        print("copying {} to {}".format(uri, newpath))
        file_io.copy(uri, newpath, True)
```
This simple test allows you to see if your Kubeflow and Pachyderm deployments are set up to exchange data.

If you want to test this in your own Kubeflow/Pachyderm deployment, 
you can either build your own container using the `Dockerfile.tf_job_s3_gateway.yml` included with this example or
use the prebuilt container.
You create a TFJob using a TFJob manifest.
We've included one with this example, 
`tf_job_s3_gateway.yaml`, 
which you would use to deploy the TFJob to your Kubeflow app
after you've edited it for your environment.

Here are the steps you need to take to run this in your environment using our prebuilt images and
Kubeflow installed with Pachyderm using the defaults for a joint deployment: 
Kubeflow in the `kubeflow` namespace and Pachyderm in the `pachyderm` namespace.

1. Deploy Pachyderm and Kubeflow.
2. Create a repo in Pachyderm using the command  `pachctl create repo inputrepo`.
3. Add the sample data from the mnist directory in this example. 
   There are two ways of the sample data: 
   either loading the data directly from GitHub or from a local copy.

   * _Loading directly from GitHub_.
     You can load the data directly from GitHub with this bash command,
     assuming you have the `seq` executable installed from the [coreutils](https://www.gnu.org/software/coreutils) library:
     ```sh
     for file in $(seq 0 9)
     do pachctl put file inputrepo@master:/data/img_${file}.jpg  \
          -f https://raw.githubusercontent.com/pachyderm/pachyderm/tree/master/examples/kubeflow/tfjob/mnist/img_${file}.jpg 
     done
     ```
      
   * _Copying from local disk_.
     You can clone this example locally or download the data to your local disk.
     You can then add the contents of the mnist directory to `inputrepo` using the command below.
     ```sh
     cd /path/where/you/put/this/example ; pachctl put file inputrepo@master:/data -r -f mnist/
     ```

4. Create a repo in Pachyderm using the command  `pachctl create repo outputrepo`.
5. Create a branch in that repo using the command `pachctl create branch outputrepo@master`.
6. Deploy the manifest to your Kubeflow using  `kubectl apply -f tf_job_s3_gateway.yaml`
7. Using your Kubeflow TFJob dashboard or `kubectl`, monitor the TFJob pod created until it completes.
   The pod will be named after the name of the TFJob in the manifest,
   which is `s3-gateway-example` by default, 
   and would be something like `s3-gateway-example-worker-0`.
8. Using your Kubeflow TFJob dashboard or `kubectl`, look at the logs for that TFJob.
   A sample log is included in the file `sample_tf_job_logs.txt` in this example.

This demonstrates that anywhere that the `file_io` library is used, 
you can put data into and get data out of Pachyderm versioned data repositories.

### Extra: Custom deployments and customizations
* Change the name of the bucket used as the parameter to the `-b` argument in `tf_job_s3_gateway.yaml`
  to match the name and branch of a repo you create in Pachyderm.
* Edit the `S3_ENDPOINT` environment variable in `tf_job_s3_gateway.yaml` to reflect your deployment.
  The default value,
  `pachd.pachyderm:600`, 
  is for Pachyderm deployed into the namespace `pachyderm` with the default S3 Gateway port.
  If, 
  for example, 
  you deployed Pachyderm into the default namespace,
  you would configure it as `pachd.default.600`, 
  or you can hardcode it using the CLUSTER-IP address of the `pachd` service obtained through `kubectl get service -n <namespace>`.
* Edit the `namespace` metadata  in `tf_job_s3_gateway.yaml` to reflect your Kubeflow deployment's namespace.


### Extra: Using the Minio client libraries

To further demonstrate compatibility and interconnectivity,
an example using the Minio client libraries is included.
The files have "minio" in their name.
It's left as an exercise for you to get them to run in your environment.









