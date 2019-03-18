# Utilizing GPUs

Pachyderm currently supports GPUs through Kubernetes device plugins. If you already have a GPU enabled Kubernetes cluster through device plugins, then skip to [Using GPUs in Pipelines](gpus.html#using-gpus-in-pipelines).

## Setting up a GPU Enabled Kubernetes Cluster

For guidance on how to set up a GPU enabled Kubernetes cluster through device plugins, refer to the Kubernetes [docs](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/).  

Setting up a GPU enabled Kubernetes cluster can be a difficult process depending on the application/framework and hardware being used. Some general things to check for if you are running into issues are:

1. The correct software is installed on the GPU machines such that applications running in Docker containers can use the GPUs. This is going to be highly dependent on the manufacturer of the GPUs and how you are using them. The most straightforward approach is to get a VM image with this pre-installed and/or use management software such as kops ([nvidia-device-plugin](https://github.com/kubernetes/kops/tree/master/hooks/nvidia-device-plugin)).
2. Kubernetes is exposing the GPU resources. This can be checked by describing the GPU nodes with `kubectl describe node`. You should see the GPU resources marked as allocatable/scheduleable if they are setup properly.
3. Your application/framework can access and use the GPUs. This may be as simple as making shared libraries accesible by the application/framework running in your container. Which can be done by baking environment variables into the Docker image or passing in environment variables through the pipeline spec.

## Using GPUs in Pipelines

 If you already have a GPU enabled Kubernetes cluster through device plugins, then using GPUs in your pipelines is as simple as setting up a GPU resource limit with the type and number of GPUs. An example pipeline spec for a GPU enabled pipeline is as follows:

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
  },
  "resource_limits": {
    "memory": "1024M",
    "gpu": {
      "type": "nvidia.com/gpu",
      "number": 1
    }
  },
  "inputs": {
    "pfs": {
      "repo": "data",
      "glob": "/*"
    }
  ]
}
```
