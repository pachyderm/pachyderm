# Autoscaling a Pachyderm Cluster

There are 2 levels of autoscaling in Pachyderm:

- Pachyderm can scale down workers when they're not in use.
- Cloud providers can scale workers down/up based on resource utilization (most often CPU).

## Pachyderm Autoscaling of Workers

You can configure autoscaling for workers by setting the [standby](../../../reference/pipeline_spec#standby-optional) field in the pipeline specification. This parameter enables you suspend your pipeline pods when no data is available to process. If new inputs come in on the pipeline corresponding to the suspended workers, Kubernetes automatically resumes them.

## Cloud Provider Autoscaling

Out of the box, autoscaling at the cloud provider layer doesn't work well with Pachyderm. However, if configure it properly, cloud provider autoscaling can complement Pachyderm autoscaling of workers.

### Default Behavior with Cloud Autoscaling

Normally when you create a pipeline, Pachyderm asks the k8s cluster how many nodes are available. Pachyderm then uses that number as the default value for the pipeline's parallelism. (To read more about parallelism, [refer to the distributed processing docs](../../concepts/advanced-concepts/distributed_computing.md)).

If you have cloud provider autoscaling activated, it is possible that your number of nodes will be scaled down to a few or maybe even a single node.  A pipeline created on this cluster would have a default parallelism will be set to this low value (e.g., 1 or 2). Then, once the autoscale group notices that more nodes are needed, the parallelism of the pipeline won't increase, and you won't actually make effective use of those new nodes.

### Configuration of Pipelines to Complement Cloud Autoscaling

The goal of Cloud autoscaling is to:

- To schedule nodes only as the processing demand necessitates it.

The goals of Pachyderm worker autoscaling are:

- To make sure your job uses a maximum amount of parallelism.
- To ensure that you process the job efficiently.

Thus, to accomplish both of these goals, we recommend:

- Setting a `constant`, high level of parallelism.  Specifically, setting the constant parallelism to the number of workers you will need when your pipeline is active.
- Setting the `cpu` and/or `mem` resource requirements [in the `resource_requests` field on your pipeline](../../../reference/pipeline_spec#resource-requests-optional).

To determine the right values for `cpu` / `mem`, first set these values rather high.  Then use the monitoring tools that come with your cloud provider (or [try out our monitoring deployment](https://github.com/pachyderm/pachyderm/blob/master/Makefile#L330)) so you can see the actual CPU/mem utilization per pod.

### Example Scenario

Let's say you have a certain pipeline with a constant parallelism set to 16.  Let's also assume that you've set `cpu` to `1.0` and your instance type has 4 cores.

When a commit of data is made to the input of the pipeline, your cluster might be in a scaled down state (e.g., 2 nodes running). After accounting for the pachyderm services (`pachd` and `etcd`), ~6 cores are available with 2 nodes. K8s then schedules 6 of your workers. That accounts for all 8 of the CPUs across the nodes in your instance group. Your autoscale group then notices that all instances are being heavily utilized, and subsequently scales up to 5 nodes total. Now the rest of your workers get spun up (k8s can now schedule them), and your job proceeds.

This type of setup is best suited for long running jobs, or jobs that take a lot of CPU time. Such jobs give the cloud autoscaling mechanisms time to scale up, while still having data that needs to be processed when the new nodes are up and running.
