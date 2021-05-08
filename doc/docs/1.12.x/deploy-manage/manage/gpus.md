# Use GPUs

Pachyderm currently supports GPUs through Kubernetes device plugins. If you
already have a GPU enabled Kubernetes cluster through device plugins,
skip to [Configure GPUs in Pipelines](#configure-gpus-in-pipelines).

## Set up a GPU-enabled Kubernetes Cluster

For instructions on how to set up a GPU-enabled Kubernetes cluster
through device plugins, see the [Kubernetes documentation](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/).

Depending your hardware and applications, setting up a GPU-enabled
Kubernetes cluster might require significant effort. If you are run
into issues, verify that the following issues are addressed:

1. The correct software is installed on the GPU machines so
that applications that run in Docker containers can use the GPUs. This is
highly dependent on the manufacturer of the GPUs and how you use them.
The most straightforward approach is to get a VM image with this
pre-installed and use management software such as
[kops nvidia-device-plugin](https://github.com/kubernetes/kops/tree/master/hooks/nvidia-device-plugin).

2. Kubernetes exposes the GPU resources. You can check this by
describing the GPU nodes with `kubectl describe node`. If the GPU resources
available configured correctly, you should see them as available for scheduling.

3. Your application can access and use the GPUs. This may be as simple as making
shared libraries accessible by the application that runs in your container. You
can  configure this by injecting environment variables into the Docker image or
passing environment variables through the pipeline spec.

## Configure GPUs in Pipelines

If you already have a GPU-enabled Kubernetes cluster through device plugins,
then using GPUs in your pipelines is as simple as setting up a GPU resource
limit with the type and number of GPUs. The following text is an example
of a pipeline spec for a GPU-enabled pipeline:

!!! example
    ```json hl_lines="12 13 14 15 16"
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
