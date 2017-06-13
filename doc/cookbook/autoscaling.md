# Autoscaling

There are 2 levels of autoscaling you need to consider with Pachyderm.

Pachyderm can scale down workers when they're not in use.

Cloud providers can scale workers down/up based on resource utilization (most often CPU).

## Setting up Pachyderm Autoscaling of Jobs

[Refer to the scaleDownThreshold](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#scale-down-threshold-optional) field in the pipeline spec. This allows you to specify the time window after which any workers corresponding to a pipeline get removed. If new inputs come in on that pipeline, they get scaled back up.


## Setting up Pachyderm Autoscaling to complement Cloud Provider Autoscaling

Out of the box, autoscaling at the cloud provider layer won't work well with Pachyderm. It can if you configure it properly.


### Default Behavior With Cloud Autoscaling 

Normally when you create a pipeline, Pachyderm asks the k8s cluster how many nodes are available, and uses that number as the default value for the pipeline's parallelism. (To read more about parallelism, [refer to the pipeline spec](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html) ).

If you have autoscaling turned on when you first starting building your pipeline, it's likely it will be fully scaled down ... to a few or maybe even a single node.

Then when you create the pipeline, it's parallelism will be set to this low value (1 or 2). And more to the point, once the autoscale group notices that more nodes are needed, the parallelism of the pipeline won't increase, and you won't actually make effective use of those new nodes.

### Configure Your Pipeline to complement autoscaling

You goal using Cloud layer autoscaling is to:

- schedule nodes only as the processing demand necessitates it

Your goal using auto scaling at the Pachyderm layer is to:

- Make sure your job uses the maximum amount of parallelism it can
- So that you process the job efficiently

To accomplish both of these goals, we recommend:

- Setting a 'constant' parallelism, and setting it to a high number ... the number you'd expect it to be at when your nodes are fully scaled up
- Setting the `cpu` and/or `mem` resource requirements [in the `resource_spec` field on your pipeline](http://docs.pachyderm.io/en/latest/reference/pipeline_spec.html#resource-spec-optional)

To determine the right values for `cpu` / `mem` utilization, we recommend setting them high at first, and using the monitoring tools that come with your cloud provider (or [trying out our monitoring deployment](https://github.com/pachyderm/pachyderm/blob/master/Makefile#L330)) so you can see the CPU/mem utilization per pod.

### Watching it in action

Let's say you have a pipeline with parallelism set to 16.

You've set `cpu` to `1.0` and your instance type has 4 cores.

When new input comes in kicking off your pipeline, your cluster is in a scaled down state. Let's say there are 2 nodes running. After accounting for the pachyderm services, that leaves ~6 cores available. K8s schedules 6 of your workers. That accounts for all 8 of the CPUs across all your nodes in your instance group. Your autoscale group notices that all instances are being heavily utilized, and scales up to 5 nodes total. Now the rest of your workers get spun up (k8s can now schedule them), and your job proceeeds.

This type of setup is best suited for long running jobs, or jobs that take a lot of CPU time. That gives the cloud autoscaling mechanisms time to scale up and still have work to process.
