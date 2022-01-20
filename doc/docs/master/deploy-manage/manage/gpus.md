# Use GPUs

Pachyderm currently supports GPUs through [Kubernetes device plugins](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/). If you
already have a GPU enabled Kubernetes cluster through device plugins,
skip to [Configure GPUs in Pipelines](#configure-gpus-in-pipelines).

## Set up a GPU-enabled Kubernetes Cluster

For instructions on how to set up a GPU-enabled Kubernetes cluster
through device plugins, see the [Kubernetes documentation](https://kubernetes.io/docs/tasks/manage-gpus/scheduling-gpus/){target=_blank}.

Depending your hardware and applications, setting up a GPU-enabled
Kubernetes cluster might require significant effort. If you are run
into issues, verify that the following issues are addressed:

1. The correct software is installed on the GPU machines so
that applications that run in Docker containers can use the GPUs. This is
highly dependent on the manufacturer of the GPUs and how you use them.
The most straightforward approach is to get a VM image with this
pre-installed and use management software such as
[kops nvidia-device-plugin](https://github.com/kubernetes/kops/tree/master/hooks/nvidia-device-plugin){target=_blank}.

2. Kubernetes exposes the GPU resources. You can check this by
describing the GPU nodes with `kubectl describe node`. If the GPU resources
available configured correctly, you should see them as available for scheduling.

3. Your application can access and use the GPUs. This may be as simple as making
shared libraries accessible by the application that runs in your container. You
can  configure this by injecting environment variables into the Docker image or
passing environment variables through the pipeline spec.

## Configure GPUs in Pipelines

Once your GPU-enabled Kubernetes cluster is set, 
you can request a GPU tier in your pipeline speifications
by [setting up GPU resource limits](../../../reference/pipeline_spec/#resource-requests-optional), along with its type and number of GPUs. 

!!! Important
      By default, Pachyderm workers are spun up and wait for new input. That works great for pipelines that are processing a lot of new incoming commits. However, for lower volume of input commits, you could have your pipeline workers 'taking' the GPU resource as far as k8s is concerned, but 'idling' as far as you are concerned. 

        - Make sure to set the `autoscaling` field to `true` so that if your pipeline is not getting used, the worker pods get spun down and you free the GPU resource.
        - Additionally, specify how much of GPU your pipeline worker will need via the `resource_requests` fields in your [pipeline specification](../../../reference/pipeline_spec/#resource-requests-optional) with `ressource_requests` <= `resource_limits`.


Below is an example of a pipeline spec for a GPU-enabled pipeline from our [market sentiment analysis example](https://github.com/pachyderm/examples/tree/master/market-sentiment){target=_blank}:

```yaml
{{ gitsnippet('pachyderm/examples', 'market-sentiment/pachyderm/train_model.json', '2.0.x') }}
```


