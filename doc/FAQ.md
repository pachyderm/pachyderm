# FAQ
[Data Storage:](#data-storage)
* [How is data storage handled in Pachyderm?](#how-is-data-storage-handled-in-pachyderm)
* [What object storage backends are currently supported?](#what-object-storage-backends-are-supported)
* [What is version control for data?](#what-is-version-control-for-data)
* [What are the benefits of version control for data?](#what-are-the-benefits-of-version-control-for-data)
* [How do you guarantee I won’t lose data in Pachyderm (i.e. replication and persistence)?](#how-do-you-guarantee-i-wont-lose-data-in-pachyderm-ie-replication-and-persistence)
* [How do I get data from other sources into Pachyderm?](#how-do-i-get-data-from-other-sources-into-pachyderm)
* [How do I get data out of Pachyderm into another system?](#how-do-i-get-data-out-of-pachyderm-into-another-system)
* [Does Pachyderm handle data locality?](#does-pachyderm-handle-data-locality)

[Deployment:](#deployment)
* [Where/how can I deploy Pachyderm?](#wherehow-can-i-deploy-pachyderm)
* [Can I use other schedulers such as Docker Swarm or Mesos?](#can-i-use-other-schedulers-such-as-docker-swarm-or-mesos)
* [Can I run Pachyderm locally?](#can-i-run-pachyderm-locally)

[Computation:](#computation)
* [What are containerized analytics?](#what-are-containerized-analytics)
* [What is the data access model?](#what-is-the-data-access-model)
* [What are jobs and how do they work?](#what-are-jobs-and-how-do-they-work)
* [What are pipelines and how do they work?](#what-are-pipelines-and-how-do-they-work)
* [How does Pachyderm manage pipeline dependencies?](#how-does-pachyderm-manage-pipeline-dependencies)
* [How do I perform batched analytics in Pachyderm?](#how-do-i-perform-batched-analytics-in-pachyderm)
* [How do I perform streaming analytics in Pachyderm?](#how-do-i-perform-streaming-analytics-in-pachyderm)
* [How is my computation parallelized?](#how-is-my-computation-parallelized)
* [How does Pachyderm let me do incremental processing?](#how-does-pachyderm-let-me-do-incremental-processing)
* [Is there a SQL interface for Pachyderm?](#is-there-a-sql-interface-for-pachyderm)

[Product/Misc:](#productmisc)
* [How does Pachyderm compare to Hadoop?](#how-does-pachyderm-compare-to-hadoop)
* [How does Pachyderm compare to Spark?](#how-does-pachyderm-compare-to-spark)
* [What are the major use cases for Pachyderm?](#what-are-the-major-use-cases-for-pachyderm)
* [Is Pachyderm enterprise production ready?](#is-pachyderm-enterprise-production-ready)
* [How does Pachyderm handle logging?](#how-does-pachyderm-handle-logging)
* [Does Pachyderm only work with Docker containers?](#does-pachyderm-only-work-with-docker-containers)
* [How do I get enterprise support for Pachyderm?](#how-do-i-get-enterprise-support-for-pachyderm)
* [What if I find bugs or have questions about using Pachyderm?](#what-if-i-find-bugs-or-have-questions-about-using-pachyderm)
* [How do I start contributing to Pachyderm?](#how-do-i-start-contributing-to-pachyderm)


## Data Storage
##### How is data storage handled in Pachyderm?
Pachyderm stores your data in any generic object storage (S3, GSC, Ceph, etc).
You can link your object storage backend to Pachyderm by following our
[cloud deployment guide](development/deploying_on_the_cloud.html)
and passing your credentials as a Kubernetes secret.

##### What object storage backends are supported?
S3 and GCS are fully supported and are the recommended backends for Pachyderm.
Support for Ceph and others are coming soon! Want to help us support more
storage backends? Check out the [GH issue](https://github.com/pachyderm/pachyderm/issues/211)!

##### What is version control for data?
We’ve all used version control for code before — Pachyderm gives you the same
semantics for petabytes of data. We even borrow our terminology from Git. In
Pachyderm, data is organized into `repos`. If you want to add or change data in
a repo, you simply `start` a `commit` make your changes, and then `finish` the
`commit`. This will create an immutable snapshot of the data that you can
reference later. Just like in Git, only the diff of the data is saved so
there is no duplication. Pachyderm exposes data as a set of diffs so you can
easily view how your data has changed over time, run a job over a previous view
of your data, or revert to a known good state if something goes wrong. Finally,
Pachyderm also let’s you branch entire data sets so you can manipulate files
and explore the data without affecting anyone else’s work. Just like with
branching in Git, Pachyderm doesn't create multiple copies of the data when you
create a branch, we just store the changes you make to it.

##### What are the benefits of version control for data?
_Instant revert_: If something goes wrong with your data, you can immediately
revert your live cluster back to a known good state.

_View diffs_: Analyze how your data is changing over time.

_Incrementality_: Only process the new data instead of recomputing everything.

_Immutable data_: Run analysis written today over your data from last month.

_Team collaboration_: Everyone can manipulate and work on the same data without stepping on each others toes.

##### How do you guarantee I won’t lose data in Pachyderm (i.e. replication and persistence)?
Your data doesn’t actually live in Pachyderm, it stays in object storage (S3 or
GCS), so it has all the safety guarantees of those underlying systems.

##### How do I get data from other sources into Pachyderm?
Pachyderm has three main methods for getting data into the system.

1. A [protobufs API](https://github.com/pachyderm/pachyderm/blob/master/src/client/pfs/pfs.proto)  that you can access through the Golang SDK. Other languages will be supported soon!
2. The [pachctl CLI](pachctl/pachctl.html), which
allows you to put files into Pachyderm.
3. You can mount Pachyderm locally and add files directly to the filesystem
   through the [FUSE interface](getting_started/beginner_tutorial.html#exploring-the-file-system).

##### How do I get data out of Pachyderm into another system?
In addition to using the same ways you get data into the system, you can also
use pipelines. Users often want to move the final results of a pipeline into
another tool such as Redshift or MySQL so that it can be easily queried through
BI tools. To accomplish this, it’s common to add a final stage to your pipeline
which reads data from Pachyderm and writes it directly to whatever other tool
you want. Redshift for example, can load data directly from an S3 bucket so the
last pipeline stage can just write to that specific bucket.

##### Does Pachyderm handle data locality?
Most object stores like S3 and GCS don’t provide any notion of locality and so
Pachyderm similarly can't provide data locality in our API. In practice, we’ve
generally found that data locality is not a bottleneck when optimizing for
performance in modern data centers.

## Deployment:
##### Where/how can I deploy Pachyderm?
Once you have Kubernetes running, Pachyderm is just a one line deploy. Since
Pachyderm’s only dependency is Kubernetes, it can be run locally, AWS, Google Cloud,
or on-prem. Check out our [local installation](getting_started/local_installation.html) and [cloud deployment](development/deploying_on_the_cloud.html) guides.

##### Can I use other schedulers such as Docker Swarm or Mesos?
Right now, Pachyderm requires Kubernetes, but we’ve purposely built it to be
modular and work with the other schedulers as well. Swarm and Mesos support
will be added in the future!

##### Can I run Pachyderm locally?
Yup! Pachyderm can be run locally using Minikube (recommended) or directly in Docker. Check out our [local installation guide](getting_started/local_installation.html)
to get started.

## Computation
##### What are containerized analytics?
Rather than thinking in terms of map or reduce jobs, Pachyderm thinks in terms
of pipelines expressed within a container. To process data, you simply create a
containerized program which reads and writes to the local filesystem. Since
everything is packaged up in a container, pipelines are easily portable,
completely isolated, and simple to monitor.

##### What is the data access model?
To process data, you simply create a containerized program which reads and
writes to the local filesystem at `/pfs/...`. Pachyderm will take your
container and inject data into it by way of a FUSE volume. We'll then
automatically replicate your container, showing each copy a different chunk of
data and processing it all in parallel. 

Check out our [beginner tutorial](getting_started/beginner_tutorial.html) or [OpenCV demo](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/opencv) to see how this works in action!

##### What are jobs and how do they work?
A job in Pachyderm is a one-off transformation or processing of data. To run a
job use the `create-job` command. In Pachyderm, jobs are meant for
experimentation or exploring your data. Once you have a job that's working well and producing useful results, you can “productionize” it by turning it into a `pipeline`.

##### What are pipelines and how do they work?
Pipelines are data transformations that are “subscribed” to data changes on
their input repos and automatically create jobs to process the new data as it comes in. A pipeline is defined by a JSON spec that describes one or more transformations
to execute when new input data is committed. All the details of a [pipeline spec](development/pipeline_spec.html) are outlined in our documentation and demonstrated in our [examples](examples/examples.html).

##### How does Pachyderm manage pipeline dependencies?
Dependencies for pipelines are handled explicitly in the pipeline spec.
Pipelines output their results to a repo of the same name.  The “input” for a
pipeline can be any set of repos, either those containing raw data or one that
was automatically created by another pipeline. For example, a pipeline stage
called “filter” would create a repo also called “filter” where it would store
the output data. A second pipeline called “sum” could have “filter” as an
input. By this method Pachyderm, actually creates a
[DAG](https://en.wikipedia.org/wiki/Directed_acyclic_graph) of data, not jobs.
The full picture would look like this: raw data enters Pachyderm which triggers
the “filter" pipeline. The “filter" pipeline outputs its results in a commit to
the “filter" repo which triggers the “sum" pipeline. The final results would be
available in the "sum" repo. Check out our [Fruit Stand demo](getting_started/beginner_tutorial.html) to see exactly this example.

##### How do I perform batched analytics in Pachyderm?
Batched analytics are the bread and butter of Pachyderm. Often times the first
stage in a batched job is a database dump or some other large swath of new data
entering the system. In Pachyderm, this would create a new commit on a repo
which would trigger all your ETL and analytics pipelines for that data. One-off
batched jobs can also be manually run on any data.

##### How do I perform streaming analytics in Pachyderm?
Streaming and batched jobs are done exactly the same way in Pachyderm. Creating
a commit is an incredibly cheap operation so you can even make one commit per
second if you want! By just changing the frequency of commits, you can
seamlessly transition from a large nightly batch job down to a streaming
operation processing tiny micro-batches of data.

##### How is my computation parallelized?
Both jobs and pipelines have a “paralellism” parameter as outlined in the [pipeline spec](development/pipeline_spec.html). This parameter dictates how
many containers Pachyderm spins up to process your data in parallel. For
example, `“paralellism”: 10` would create up to 10 containers that each process 1/10 of
the data. Each pipeline can have a different parallelization factor, giving you
fine-grain control over the utilization of your cluster. `paralellism` can be set to
`0` in which case Pachyderm will set it automatically based on the size of the
cluster.

##### How does Pachyderm let me do incremental processing?
Pachyderm exposes all your data in diffs, meaning we show you the new data that
has been added since the last time a pipeline was run. Pachyderm will smartly
only process the new data and append those results to the output from the
previous run. We have extensive [documentation on incrementality](advanced/incrementality.html) that'll show you the fine-grain control you can have to optimizing computation.

##### Is there a SQL interface for Pachyderm?
Not yet, but it’s coming soon! If you want to query your data using SQL, you
can easily create a pipeline that pushes data from Pachyderm into your favorite
SQL tool such as PostGres.

### Product/Misc
##### How does Pachyderm compare to Hadoop?
Pachyderm is inspired by the Hadoop ecosystem but shares no code with it.
Instead, we leverage the container ecosystem to provide the broad functionality
of Hadoop with the ease of use of Docker. Similar to Hadoop, Pachyderm offers
virtually infinite horizontal scaling for both storage and processing power.
That said, there are two bold new ideas in Pachyderm:

1. Containers as the core processing primitive — You can do analysis using any languages or libraries you want.
2. Version Control for data — We let your team collaborate effectively on data using a commit-based distributed filesystem (PFS), similar to what Git does with code.

##### How does Pachyderm compare to Spark?
The only strong similarity between Pachyderm and Spark is that our versioning
of data is somewhat similar to how Spark uses RDD’s and data frames to speed up computation. Spark is a fantastic interface for exploring your data or running queries. In our opinion, Spark is one of the best parts of the Hadoop ecosystem and in the
near future, we’ll be offering a connector that lets you use the Spark
interface on top of Pachyderm.

##### What are the major use cases for Pachyderm?
__Data Lake__:
A data lake is a place to dump and process gigantic data sets. This is where
you send your nightly production database dumps, store all your raw log files
and whatever other data you want. You can then process that data using any code
you can put in a container. Martin Fowler has a great [blog post](http://martinfowler.com/bliki/DataLake.html)
describing data lakes.

__Containerized ETL__:
ETL (extract, transform, load) is the process of taking raw data and turning it
into a useable form for other services to ingest. ETL processes usually involve
many steps forming a DAG ([Directed Acyclical Graph](https://en.wikipedia.org/wiki/Directed_acyclic_graph)) — pulling raw
data from different sources, teasing out and structuring the useful details, and
then pushing those structures into a data warehouse or BI (business
intelligence) tool for querying and analysis.

Pachyderm completely manages your ETL DAG by giving you explicit control over
the inputs for every stage of your pipeline. We also give you a simple API —
just read and write to the local file system inside a container — so it’s easy
to push and pull data from a variety of sources.

__Automated ML pipelines__:
Developing machine learning pipelines is always an iterative cycle of
experimenting, training/testing, and productionizing. Pachyderm is ideally
suited for exactly this type of process.

Data scientists can create jobs to explore and process data. Pachyderm will
automatically let you down-sample data or develop analysis locally without
having to copy any data around.

Building training/testing data sets is incredibly easy with version-controlled
data. Since you have all your historical data at your fingertips, you can
simply train a model on data from last week and then test it on this week’s
data. Getting training/testing pairs involves zero data copying or moving.

Finally, once your analysis is ready to go, you simply add your job to
Pachyderm as a pipeline. Now it’ll automatically run and continue updating as
new data comes into the system, letting you seamlessly transition from
experimentation to a productionized deployment of your new model.

##### Is Pachyderm enterprise production ready?
Yes! Pachyderm hit v1.2 and is ready for production use! If you need help
with your deployment or just want to talk to us about the details, we’d love to
hear from you on [Slack](http://slack.pachyderm.io) or email us at [support@pachyderm.io](mailto:support@pachyderm.io).

##### How does Pachyderm handle logging?
Kubernetes actually handles all the logging for us. You can use `pachctl
get-logs` to get logs from your jobs. Kubernetes also comes with it’s own tools
for pushing those logs to whatever other services you use for log aggregation
and analysis.

##### Does Pachyderm only work with Docker containers?
Right now yes, but that's mostly because Kubernetes doesn't yet support other
runtimes. Pachyderm has no strict dependencies on Docker so we’ll have support
for rkt and other container formats soon.

##### How do I get enterprise support for Pachyderm?
If you’re using Pachyderm in production or evaluating it as a potential
solution, we’d love to chat with you! support@pachyderm.io

##### What if I find bugs or have questions about using Pachyderm?
You can submit bug reports, questions, or PR’s on
[Github](https://github.com/pachyderm/pachyderm/issues), join our [users channel on Slack](http://slack.pachyderm.io), or email us at [support@pachyderm.io](mailto:support@pachyderm.io) and we can help you right away.

##### How do I start contributing to Pachyderm?
We're thrilled to have you contribute to Pachyderm! Check out our
[contributor guide](https://github.com/pachyderm/pachyderm/blob/master/CONTRIBUTING.md)
to see all the details. If you're not sure where to start, recent issues on
[Github](https://github.com/pachyderm/pachyderm/issues) or ones that are labeled as
“noob-friendly” are good places to begin.

