# Estimate Pi Using Spark

This example demonstrates integration of Spark with Pachyderm by launching
a Spark job on an existing cluster from within a Pachyderm Job. The job uses
configuration info that is versioned within Pachyderm, and stores it's reduced
result back into a Pachyderm output repo, maintaining full provenance and
version history within Pachyderm, while taking advantage of Spark for
computation.

The example assumes that you have:

- A Pachyderm cluster running - see [this guide](https://docs.pachyderm.com/latest/getting_started/local_installation/) to get up and running with a local Pachyderm cluster in just a few minutes.
- The `pachctl` CLI tool installed and connected to your Pachyderm cluster - see [the relevant deploy docs](https://docs.pachyderm.com/latest/deploy-manage/deploy/) for instructions.
- The `kubectl` CLI tool installed (you will likely have installed this while [setting up your local Pachyderm cluster](https://docs.pachyderm.com/latest/getting_started/local_installation/))

Note: if deploying on Minikube, you'll need to increase the default memory
allocation to accomodate the deploy of a Spark cluster. When running `minikube
start`, append `--memory 4096`.

## Set up Spark Cluster

The simpelst way to run this example is by deploying a Spark cluster into the
same Kubernetes cluster on which Pachyderm is running. We'll do so with Helm.
(Note: if you already have an external Spark cluster running, you can skip this
section. Be sure to read [the note about connecting to an existing Spark
cluster](#connecting-to-an-existing-spark-cluster))

### Install Helm

If you don't already have the Helm client installed, you can do so by following
[the instructions
here](https://docs.helm.sh/using_helm/#installing-the-helm-client) (or, for the
bold, by running `curl
https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash`.)

### Set up Helm/Tiller

In order to use Helm with your Kubernetes cluster, you'll need to install
Tiller:

```
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller --upgrade
```

Tiller will take about a minute to initialize and enter `Running` status. You
can check it's status by running: `kubectl get pod -n kube-system -l
name=tiller`


### Install Spark

Finally, once Tiller is `Running`, use Helm to install Spark:

```
helm install --name spark stable/spark
```

This will again take several minutes to pull the relevant Docker images and
start running. You can check the status with `kubectl get  pod -l
release=spark`

## Deploy Pachyderm Pipeline

Once your Spark cluster is running, you're ready to deploy the Pachyderm
pipeline:


```
# create a repo to hold configuration data that acts as input to the pipeline
pachctl create repo estimate_pi_config

# create the actual processing pipeline
pachctl create pipeline -f estimate_pi_pipeline.json

# kick off a job with 1000 samples
echo 1000 | pachctl put file estimate_pi_config@master:num_samples

# check job status
pachctl list job --pipeline estimate_pi

# once job has completed, retrieve the results
pachctl get file estimate_pi@master:pi_estimate

```

## Connecting to an existing Spark cluster

By default, this example makes use of Kubernetes' service discovery to connect
your Pachyderm pipeline code to your Spark cluster. If you wish to connect to
a different Spark cluster, you can do so by adding the `--master` flag to the
list of arguments provided to `cmd` in the pipeline spec: append `"--master"`
and `"spark://$MYSPARK_MASTER_SERVICE_HOST:$MYSPARK_MASTER_SERVICE_PORT"` to
the `cmd` array.

To test a manually-specified connection, deploy a Spark cluster into
a different name in Kubernetes:

```
helm install --name my-custom-spark stable/spark
```

