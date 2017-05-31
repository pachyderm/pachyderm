# Production

When deploying a cluster to production, you'll want to consider the following:

1. [Ingress and Egress schemes](production.html#ingress-and-egress-schemes)
2. [Resource Utilization](production.html#resource-utilization)
3. [Tools](production.html#tools)
4. [Deployment](production.html#deployment)
5. [Code management](production.html#code-management)

## Ingress and Egress Schemes

How are you getting data into and out of your cluster?

Common data ingress patterns are:

- a DB connector
- creating a commit per time window

Common data egress patterns are:

- uploading output to an object store for file-like viewing (refer to our 'egress' feature)
- a database adapter
- a persistent service

## Resource Utilizations

Making sure your cluster is up and running means you also want to think about the resources you're using.

The question you need to answer is how much data you need when.

Are the results of the pipelines needed every minute / hour / day?

Once you have an answer to that, you need to consider:

1) Configuring your pipelines to scale

To do this, take a look at our section on parallelism and writing jobs that utilize parallelism.

2) Specifying the resources that each pipeline needs

Perhaps one pipeline needs a bunch of CPU. Or a bunch of memory. Or both. Or just a ton of GPU.

You can specify the resources needed per pipeline with resource requests.

3) The pool of VMs

Once you specify what each pipeline needs to run, to guarantee that data flows through the DAG at the rate you need, you should examine:

- how many 'runs' of the whole DAG will be happening at once? which will help you answer:
- how many pipelines will be running at once?
- and how much resources in total the cluster will need at any one time?

Based on those answers, you can scale your pool of VMs accordingly.

Pachyderm workers will autoscale (as you set the 'spin down' interval). So you'll want a bit of extra bandwidth to accomodate peak load plus a buffer.

## Tools

Once you have the cluster up, it's critical to have insight to its state.

You'll need:

1) Monitoring

We recommend having visual monitoring in place for your cluster. We provide a make task (`make launch-monitoring`) that'll stand up a monitoring service for you (heapster / influxDB / grafana). You can refer to those service manifests for a bit more info on how to access the monitoring service.

Monitoring is critical for understanding HOW your pool of resources is being used. Tracking CPU/mem/io usage is important. Grafana also allows you to setup some alerts (which can be helpful for understanding when you hit a new threshold of scaling needs).

2) Logging

Logging is critical for understanding how your pipelines are running, and how healthy your cluster is.

We provide a make task (`make launch-logging`) that will spin up a logging service (fluentd / elastic-search / kibana) to allow you to aggregate and search through your logs. Without this service, there is no k8s default log aggregator ... your logs will disappear whenever k8s decides to garbage collect.

Logging will be critical for _your_ data engineers and data scientists to develop their pipelines. It will also be required for any debugging of the cluster itself.

3) Debugging

Your team will need access to the above tools when developing their pipelines and scaling the cluster. You'll need for them to have access to these tools. Each provider handles user management differently, but broadly speaking, this can be handled at the cloud provider layer.

For convenience sake, you should setup DNS for your cluster so that you're not subject to ephemeral IPs when trying to SSH, or when connecting pachctl to the cluster directly via the `ADDRESS` variable.

## Deployment

Just like any other software deployment, we recommend a staged deploy. Your needs are determined by your team size.

Some recommendations:

1) Data Integration

Write tests!

You should be able to use a CI (continuous integration) solution (e.g. circleCI, travisCI, or jenkins) to run tests for your pipelines!

We call this 'data integration'

You'll setup a minimally representative data set so that your pipelines can run, and be verified anytime you change your code.

2) Staging

Setup a staging server that runs as much like production as you can manage. Ideally:

- it ingests the same data / at the same rate and volume
- it runs on the same number of VMs / etc so you can triage any scaling issues from code changes

4) Embedding QC

Embedding quality control is a good idea.

If you have the pipeline:

```
A --> B --> C
```

But you can run some heuristics to validate the output at each step, you may want to do so. In this case the updated DAG would look like:

```
A --> validate_A --> B --> validate_B --> C
```

The advantage here is that if any validation step fails ... the data stops propagating. You can setup alerts / etc to detect this.

We've seen this prove useful for mission critical applications. But it can be heavy handed for applications that have a looser SLA. Either way - you'll always have the tools of data provenance at hand to debug any issues that arise.


## Code Management

Just as you would checking your code to git, it's important to checkin the k8s manifests you use to create your pipelines and cluster.

1) Cluster Manifest

So instead of:

```
$pachctl deploy amazon ...
```

We recommend:

```
$pachctl deploy --dry-run amazon ... > production-cluster.json
```

2) Pipeline Manifests

A single `pipeline.json` file can contain multiple pipelines. Keeping one or many in a file is up to you.

However, we do recommend keeping those pipeline manifests alongside the code that they run, and versioning them right alongside the code. That makes any future code debugging you need to do much easier.

