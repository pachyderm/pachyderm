# How to Design Your Pachyderm Workflow

Pachyderm provides various ways to flow your data in and out of
Pachyderm. For best results, you need to pick the right type of
a pipeline and parameters that suit your use case and provide
maximum performance. The best way to think about your data pipeline
is to imagine it as a
[Directed Acyclic Graph (DAG)](https://en.wikipedia.org/wiki/Directed_acyclic_graph)
in which your data has an input source, a transformation block,
and an output directory.

However, a pipeline can represent a more sophisticated use case.
For example, you can configure your pipeline to listen on a specific
port and consume streaming data continuously. Alternatively, you can
set a pipeline that streamlines the data to an outside funnel, such
as a Kubernetes `Service` which can expose the data to the outside
world in different forms. For example, in the form of a dashboard.

The following diagram describes the types of pipelines that you
can create in Pachyderm and how you can use them to upload data into
Pachyderm and expose it out to third-party tools and systems:

[![Pipeline types](../../../assets/images/d_pipeline_select.svg)](../../../assets/images/d_pipeline_select.png)

In the diagram above, you see types of pipelines that pull and
push data in and out of Pachyderm. You decide based on your use
case wether you want to trigger events in Pachyderm, or you want
Pachyderm to act on specific events as soon as they occur.

**User-initiated actions:**

* Push in — the user uploads data to
  Pachyderm by running the `pachctl put file` command or by creating a
  spout-service. `pachctl put file` is a standard and most straightforward
  way to upload your data. With `put file`, you do not need to configure
  any additional APIs on your data source side. However, if your data
  source has no notion of Pachyderm API, or, if you want to automate
  data flowing into Pachyderm, `pahctl put file` might not be the best
  choice.
  In a spout-service, a Pachyderm spout pipeline automatically pushes
  data into Pachyderm, and a Pachyderm service pipeline provides
  HTTP APIs to push the data out of Pachyderm. An example of a
  spout-service is an automated data streaming from Git into Pachyderm.

* Pull out — the user initiates data pulling from the output repository.
  An example of user-initiated data pulling out of Pachyderm is a service
  pipeline. A service pipeline has only an input repository and does not
  have an output repository. It reads the data from the input repository,
  and the users can pull out the data as needed. For example, to create
  a dashboard from the collected data.

**Pachyderm-initiated actions:**

* Pull in — Pachyderm triggers a cron pipeline on a
  specific schedule and the pipeline performs the configured transformations,
  such as a website scraping or GitHub querying, and so on.

  Unlike a cron pipeline, a spout pipeline runs continuously and writes
  the data according to the specification in the Docker image. A spout
  creates a *named pipe* in `pfs/out`, and the user's code writes `.tar`
  streams into that pipe while continuously waiting for new data being
  generated and submitted into the named pipe.

* Push out — Pachyderm can push out data to external object storage,
  such as an S3 store, by using the `egress` property in the pipeline
  specification or to an external database by using standard pipeline
  parameters.
