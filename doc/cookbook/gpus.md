# Utilizing GPUs

Pachyderm has alpha support for utilizing GPUs within Pachyderm pipelines (e.g., for training machine learning models).  To do this you will need to:

1) [Create a Docker image that is able to utilize GPUs](gpus.html#creating-a-gpu-enabled-docker-image)
2) [Write a pipeline spec that specifies GPU nodes](gpus.html#writing-your-pipeline-specification)
3) [Deploy a GPU enabled pachyderm cluster](gpus.html#deploy-a-gpu-enabled-pachyderm-cluster)

For a concrete example, see our [example Tensorflow pipeline](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/ml/tensorflow) for image-to-image translation, which includes a pipeline specification for running model training on a GPU node.

## Creating a GPU Enabled Docker Image

For your Docker image, you'll want to use or build an image that can utilize GPU resources.  If you are using Tensorflow, for example, you could build your Docker image `FROM` the public GPU enabled Tensorflow image:

```
FROM tensorflow/tensorflow:0.12.0-gpu
...
```

Or you might follow [this guide](https://github.com/NVIDIA/nvidia-docker) for working with Docker and NVIDIA (although we haven't full tested this guide).

## Writing Your Pipeline Specification

### Ensuring that your environment can access GPU drivers

You can bake this into your Docker image in some cases, but other images, such as the TensorFlow base image, may require that you explicitly tell your application about shared libraries (e.g., CUDA).  To do that, you may need to set one or more environmental variables.  This will be application/framework dependent.  For example, if we were using the Tensorflow base image, we would need to make sure that we set `LD_LIBRARY_PATH`, such that TensorFlow knows about CUDA:

```
LD_LIBRARY_PATH="/usr/lib/nvidia:/usr/local/cuda/lib64:/rootfs/usr/lib/x86_64-linux-gnu"
```

Again, this can be baked into your Docker image via an `ENV` statement in your Dockerfile:

```
ENV LD_LIBRARY_PATH /usr/lib/nvidia:/usr/local/cuda/lib64:/rootfs/usr/lib/x86_64-linux-gnu
```

or it can be defined in your [pipeline specification](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html) via the `env` field (as shown below).  

### Creating your pipeline specification with access to GPU resources

In addition to properly setting up the environment, we need to tell the Pachyderm cluster that our pipeline needs a GPU resource. To do that we'll add a `gpu` entry to the `resources` field in the [pipeline specification](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html).

An example pipeline definition for a GPU enabled Pachyderm Pipeline is as follows:

```
{
  "pipeline": {
    "name": "train"
  },
  "transform": {
    "image": "acme/your-gpu-image",
    "cmd": [ 
      "python",
      "train.py"
    ],
    "env": {
      "LD_LIBRARY_PATH": "/usr/lib/nvidia:/usr/local/cuda/lib64:/rootfs/usr/lib/x86_64-linux-gnu"
    }
  },
  "resource_spec": {
      "gpu": 1
  },
  "inputs": {
    "atom": {
      "repo": "data",
      "glob": "/*"
    }
  ]
}
```

## Deploy a GPU Enabled Pachyderm Cluster

**NOTE:** You can also [test Pachyderm + GPUs locally](#test-locally)

**NOTE:** The following has been tested with these versions of k8s/kops:

```
$kubectl version
Client Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.1", GitCommit:"b0b7a323cc5a4a2019b2e9520c21c7830b7f708e", GitTreeState:"clean", BuildDate:"2017-04-03T20:44:38Z", GoVersion:"go1.7.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.2", GitCommit:"477efc3cbe6a7effca06bd1452fa356e2201e1ee", GitTreeState:"clean", BuildDate:"2017-04-19T20:22:08Z", GoVersion:"go1.7.5", Compiler:"gc", Platform:"linux/amd64"}
$kops version
Version 1.6.0-beta.1 (git-77f222d)
```

To deploy a Pachyderm cluster with GPU support we assume:

- You're using kops for your deployment (which means you're using AWS or GCE, not GKE). Other deploy methods are available, but these are the ones we've tested most thoroughly. 
- You have a working pachyderm cluster already up and running that you can connect to with `kubectl`.

### Add GPU nodes to your k8s cluster

You can create GPU nodes by first using `kops` to create a new instance group:

```
$kops create ig gpunodes --name XXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYYY --subnet us-west-2c
```

where your specification will look something like:

```
  1 apiVersion: kops/v1alpha2
  2 kind: InstanceGroup
  3 metadata:
  4   creationTimestamp: null
  5   name: gpunodes
  6 spec:
  7   image: kope.io/k8s-1.5-debian-jessie-amd64-hvm-ebs-2017-01-09
  8   machineType: p2.xlarge
  9   maxSize: 1
 10   minSize: 1
 11   role: Node
 12   subnets:
 13   - us-west-2c
```

In this example we used Amazon's p2.xlarge instance which contains a single GPU node.

**Node** -  If you upped your `rootVolumeSize` (and set the `rootVolumeType` in your other instance group), you should do the same here. In the absence of GPU jobs, normal jobs could get scheduled on this node, in which case you'll have the same disk requirements as the rest of your cluster. There is currently no way of setting "disk" resource requests, so we have to use a convention instead.

### Enable GPUs at the k8s level

Again, you can use `kops` to edit your cluster:

```
$kops edit cluster --name XXXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYY
```

and add the fields:

```
  hooks:
  - execContainer:
      image: pachyderm/nvidia_driver_install:dcde76f919475a6585c9959b8ec41334b05103bb
  kubelet:
    featureGates:
      Accelerators: "true"
```

**Note:** It's YAML and spaces are very important. Also, if you see "fields were not recognized," you likely need to update the version of `kops`.

These lines provide an image that gets run on every node's startup. This image will install the NVIDIA drivers on the host machine, update the host machine to mount the device at startup, and restart the host machine.

The feature gate enables k8s GPU detection. That's what gives us the `alpha.kubernetes.io/nvidia-gypu: "1"` resources.

### Update your cluster

Finally, we "update" our cluster to actually make the above changes:

```
$kops update cluster --name XXXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYY --yes
```

This will spin up the new `gpunodes` instance group, and apply the changes to your kops cluster. 

### Sanity check

You'll know the cluster is ready to schedule GPU resources when:

- you see the new node in the output of `kubectl get nodes` and the state is `Ready`, and 
- the node has the `alpha.kubernetes.io/nvidia-gpu: "1"` field set (and the value is 1 not 0)

```
$kubectl get nodes/ip-172-20-38-179.us-west-2.compute.internal -o yaml | grep nvidia
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu-name: Tesla-K80
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu: "1"
```

### Deal with known issues (if necessary)

If you're not seeing the node, its possible that your resource limits (from your cloud provider) are preventing you from creating the GPU node(s).  You should check your resource limits and ensure that GPU nodes are available in your region/zone (as further discussed [here](../managing_pachyderm/deploy_troubleshooting.html#gpu-node-never-appears)).  

If you have checked your resource limits and everything seems ok, its very possible that you're hitting [a known k8s bug](https://github.com/kubernetes/kubernetes/issues/45753).  In this case, you can try to overcome the issue by restarting the k8s api server. To do that, run:

```
kubectl --namespace=kube-system get pod | grep kube-apiserver | cut -f 1 -d " " | while read pod; do kubectl --namespace=kube-system delete po/$pod; done
```

It can take a few minutes for the node to get recognized by the k8s cluster again.

## Test Locally

**NOTE** - This has only been tested on a linux machine.

If you want to test that your pipeline is working on a local cluster (you're Pachyderm in a local cluster), you can do so, but you'll need to attach the NVIDIA drivers correctly.  There are two methods for this:

### 1. Fresh install

Install the NVIDIA drivers locally if you haven't already. If you're not sure, run `which nvidia-smi`. If it returns no result, you probably don't have them installed.  To install them, you can run the following command.  Warning! This command will restart your system and will modify your `/etc/rc.local` file, which you may want to backup.

```
$sudo /usr/bin/docker run -v /:/rootfs/ -v /var/run/dbus:/var/run/dbus -v /run/systemd:/run/systemd --net=host --privileged pachyderm/nvidia_driver_install:dcde76f919475a6585c9959b8ec41334b05103bb
```

After the restart, you should see the nvidia devices mounted:

```
$ls /dev | grep nvidia
nvidia0
nvidiactl
nvidia-modeset
nvidia-uvm
```

At this point your local machine should be recognized by kubernetes. To check you'll do something like:

```
$kubectl get nodes
NAME        STATUS    AGE       VERSION
127.0.0.1   Ready     13d       v1.6.2
$kubectl get nodes/127.0.0.1 -o yaml | grep nvidia
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu-name: Quadro-M2000M
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu: "1"
```

If you don't see any `alpha.kubernetes.io/nvidia-gpu` fields it's likely that you didn't deploy k8s locally with the correct flags. An example of [the right flags can be found here](https://github.com/pachyderm/pachyderm/blob/master/etc/kube/internal.sh). You can clone git@github.com:pachyderm/pachyderm and run `make launch-kube` locally if you're already running docker on your local machine.

### 2. Hook in existing drivers

Pachyderm expects to find the shared libraries it needs under `/usr/lib`. It mounts in `/usr/lib` into the container as `/rootfs/usr/lib` (only when you've specified a GPU resource). In this case, if your drivers are not found, you can update the `LD_LIBRARY_PATH` in your container as appropriate.
