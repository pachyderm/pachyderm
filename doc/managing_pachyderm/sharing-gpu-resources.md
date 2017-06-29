# Sharing GPU Resources

Often times, teams are running big ML models on instances with GPU resources.

GPU instances are expensive! You want to make sure that you're utilizing the GPUs you're paying for!

# Without configuration

[To deploy a pipeline that relies on GPU](http://docs.pachyderm.io/en/latest/cookbook/tensorflow_gpu.html), you'll already have set the `gpu` resource requirement in the pipeline specification. But Pachyderm workers by default are long lived ... the worker is spun up and waits for new input. That works great for pipelines that are processing a lot of new incoming commits.

For ML workflows, especially during the development cycle, you probably will see lower volume of input commits. Which means that you could have your pipeline workers 'taking' the GPU resource as far as k8s is concerned, but 'idling' as far as you're concerned.

Let's use an example.

Let's say your cluster has a single GPU node with 2 GPUs. Let's say you have a pipeline running that requires 1 GPU. You've trained some models, and found the results were surprising. You suspect your feature extraction code, and are delving into debugging that stage of your pipeline. Meanwhile, the worker you've spun up for your GPU training job is sitting idle, but telling k8s it's using the GPU instance.

Now your coworker is actively trying to develop their GPU model with their pipeline. Their model requires 2 GPUs. But your pipeline is still marked as using 1 GPU, so their pipeline can't run!

# Configuring your pipelines to share GPUs

Whenever you have a limited amount of a resource on your cluster (in this case GPU), you want to make sure you've specified how much of that resource you need via the `resource_spec` as [part of your pipeline specification](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html). But, you also need to make sure you set the `scaleDownThreshold` flag so that if your pipeline is not getting used, the worker pods get killed, and you free the resource.

In the example above, both you and your coworker already have the `gpu` request set, but you both should add the `scaleDownThreshold` field. You probably want it set to something like `15m` if usually only one of you is running something at a time. If you're both running stuff at the same time quite often, you can set it lower, to `1m`. In this case, after your model has run, and you're working on your code locally, your pipeline's workers will get spun down, and your coworker's training pipeline will have access to the resources it needs to run. Then when your coworker's GPU training job is done, same thing! Their worker pods will get scaled down to 0 and the resources will be freed for whomever runs the next job requiring GPU resources.
