# Tensor Flow using GPUs

Pachyderm has alpha support for training TensorFlow models with GPUs.

This recipe will guide you through how to deploy a pipeline on pachyderm that uses TensorFlow with GPU resources.

1) [Writing your pipeline](tensorflow_gpu.html#writing-your-pipeline)
2) [Test locally (optional)](tensorflow_gpu.html#test-locally)
3) [Deploy a GPU enabled pachyderm cluster](tensorflow_gpu.html#deploy-a-gpu-enabled-pachyderm-cluster)

---

## Writing Your Pipeline

For your Docker image, you'll want to use a TensorFlow image, e.g:

```
FROM tensorflow/tensorflow:0.12.0-gpu
...
```

Tensorflow provides the cuda shared libraries already as part of the image. However, we'll need to make sure that TensorFlow can find these shared libraries, and the NVidia shared libraries that come from installing the NVidia drivers.

To do that, we need to make sure that we set `LD_LIBRARY_PATH` appropriately.

In this case we want it to be:

```
LD_LIBRARY_PATH="/usr/lib/nvidia:/usr/local/cuda/lib64:/rootfs/usr/lib/x86_64-linux-gnu"
```

Additionally, we need to tell the Pachyderm cluster that this pipeline needs a GPU resource. To do that we'll add a `gpu` entry to the `resources` field on the pipeline.

So, an example pipeline definition for a GPU enabled Pachyderm Pipeline is:

```
{
  "pipeline": {
    "name": "train"
  },
  "transform": {
    "image": "acme/your-gpu-image",
    "cmd": [ "/bin/bash" ],
    "stdin":[
        "export LD_LIBRARY_PATH=\"/usr/lib/nvidia:/usr/local/cuda/lib64:/rootfs/usr/lib/x86_64-linux-gnu\"",
        "./Training_Script.py",
    ]
  },
  "parallelism_spec": {
    "strategy": "CONSTANT",
    "constant": 1
  },
  "resource_spec": {
      "memory": "250M",
      "cpu": 1,
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

**NOTE:** You can also [test running Pachyderm using GPUs locally](#test-locally)



## Deploy a GPU Enabled Pachyderm Cluster

To deploy a Pachyderm cluster with GPU support we assume:

- you're using kops for your deployment (which means you're using AWS or GCE, not GKE). Other deploy methods are available, but these are the ones we've tested most thoroughly. 
- you have a working pachyderm cluster already up and running that you can connect to with `kubectl`

This has been tested with these versions of k8s/kops:

```
$kubectl version
Client Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.1", GitCommit:"b0b7a323cc5a4a2019b2e9520c21c7830b7f708e", GitTreeState:"clean", BuildDate:"2017-04-03T20:44:38Z", GoVersion:"go1.7.5", Compiler:"gc", Platform:"linux/amd64"}
Server Version: version.Info{Major:"1", Minor:"6", GitVersion:"v1.6.2", GitCommit:"477efc3cbe6a7effca06bd1452fa356e2201e1ee", GitTreeState:"clean", BuildDate:"2017-04-19T20:22:08Z", GoVersion:"go1.7.5", Compiler:"gc", Platform:"linux/amd64"}
$kops version
Version 1.6.0-beta.1 (git-77f222d)
```

First, you'll need to:

1) Edit your kops cluster to add GPU machines

For example:

```
$kops create ig gpunodes --name XXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYYY --subnet us-west-2c
```
And your spec will look something like:

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

In this example we're using Amazon's p2.xlarge instance which contains a single
GPU node.

2) Edit your kops cluster to enable GPU at the K8s level

```
$kops edit cluster --name XXXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYY
```

And add the fields:

```
  hooks:
  - execContainer:
      image: pachyderm/nvidia_driver_install:dcde76f919475a6585c9959b8ec41334b05103bb
  kubelet:
    featureGates:
      Accelerators: "true"
```

**Note:** It's YAML and spaces are very important.

These lines provide an image that gets run on every node's startup. This image
will install the nvidia drivers on the host machine, update the host machine to
mount the device at startup, and restart the host machine.

The feature gate enables k8s GPU detection. That's what gives us the
`alpha.kubernetes.io/nvidia-gpu: "1"` resources.

4) Update your cluster


```
$kops update cluster --name XXXXXXX-pachydermcluster.kubernetes.com --state s3://k8scom-state-store-pachyderm-YYYY --yes
```

This will spinup the new 'gpunodes' instance group, and apply the changes to
your kops cluster. 

5) Check the status

You'll know the cluster is ready to schedule GPU resources when:

a) you see the new node in the output of `kubectl get nodes` and the state is
`Ready`

and

b) the node has the `alpha.kubernetes.io/nvidia-gpu: "1"` field set (and the
	value is 1 not 0)

e.g.

```
$kc get nodes/ip-172-20-38-179.us-west-2.compute.internal -o yaml | grep nvidia
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu-name: Tesla-K80
    alpha.kubernetes.io/nvidia-gpu: "1"
    alpha.kubernetes.io/nvidia-gpu: "1"
```

6) Poke k8s if necessary

If you're not seeing the node, its possible you're hitting [a known k8s
bug](https://github.com/kubernetes/kubernetes/issues/45753).

In this case, it helps to restart the k8s api server. To do that, run:

```
kubectl --namespace=kube-system get pod | grep kube-apiserver | cut -f 1 -d " " | while read pod; do kubectl --namespace=kube-system delete po/$pod; done
```

It can take a few minutes for the node to get recognized by the k8s cluster
again.

Now you're all done and set up to train models using TensorFlow and GPUs in your Pachyderm cluster. If you've got a further questions, you can come find us in our public [Slack Channel](slack.pachyderm.io) or on [Github](github.com/pachyderm/pachyderm).


## Test Locally

If you want to test that your pipeline is working on a local cluster (you're running docker/kubernetes/pachyderm in a local cluster), you can do so, but you'll need to attach the NVidia drivers correctly. This has only been tested on a local linux machine.

There are two methods for this.

1) Fresh install

If you haven't installed the nvidia drivers locally, you can do this. If you're not sure, run `which nvidia-smi` and if it returns no result, you probably don't have them installed.

To install them, you can run the same command that we have the VM's run to install the drivers. Warning! This command will restart your system (and will modify your `/etc/rc.local` file, which you may want to backup).

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

2) Hook in existing drivers

Pachyderm expects to find the shared libraries it needs under `/usr/lib`. It mounts in `/usr/lib` into the container as `/rootfs/usr/lib` (only when you've specified a GPU resource). In this case, if your drivers are not found, you can update the `LD_LIBRARY_PATH` in your container as appropriate.
