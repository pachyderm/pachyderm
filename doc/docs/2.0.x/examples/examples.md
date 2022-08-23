# Examples
A curated list of examples that use Pachyderm.

!!! Warning
        For **Informational Purposes** ONLY. Those examples might not be production-ready.
## OpenCV Edge Detection

This example does edge detection using OpenCV. This is our canonical starter demo. If you haven't used Pachyderm before, start here. We'll get you started running Pachyderm locally in just a few minutes and processing sample log lines.

[Open CV](https://docs.pachyderm.com/latest/getting-started/beginner-tutorial/){target=_blank}

## Word Count (Map/Reduce)

Word count is basically the "hello world" of distributed computation. This example is great for benchmarking in distributed deployments on large swaths of text data.

[Word Count](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/word_count){target=_blank}

## Periodic Ingress from a Database

This example pipeline executes a query periodically against a MongoDB database outside of Pachyderm.  The results of the query are stored in a corresponding output repository.  This repository could be used to drive additional pipeline stages periodically based on the results of the query.

[Periodic Ingress from MongoDB](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/db){target=_blank}

## Lazy Shuffle pipeline

This example demonstrates how lazy shuffle pipeline i.e. a pipeline that shuffles, combines files without downloading/uploading can be created. These types of pipelines are useful for intermediate processing step that aggregates or rearranges data from one or many sources.

[Lazy Shuffle pipeline](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/shuffle){target=_blank}

## Variant Calling and Joint Genotyping with GATK

This example illustrates the use of GATK in Pachyderm for Germline variant calling and joint genotyping. Each stage of this GATK best practice pipeline can be scaled individually and is automatically triggered as data flows into the top of the pipeline. The example follows [this tutorial](https://drive.google.com/open?id=0BzI1CyccGsZiQ1BONUxfaGhZRGc) from GATK, which includes more details about the various stages.

[GATK - Variant Calling](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/gatk){target=_blank}

## Pachyderm Pipelines

This section lists all the examples that you can run with various
Pachyderm pipelines and special features, such as transactions.

### Inner and Outer Joins Input

A join is a special type of pipeline that enables you to perform
data operations on files with a specific naming pattern.

[Inner and Outer joins 101 - A simplified retail use case](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/joins){target=_blank}
### Group Input

A group is a special type of pipeline input that enables 
you to aggregate files that reside in one or separate Pachyderm
repositories and match a particular naming pattern. 

[Group 101 - An Introductory example](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/group){target=_blank}
### Spouts

A spout is a special type of pipeline that you can use to ingest
streaming data and perform such operations as sorting, filtering, and other.

We have released a new *spouts 2.0* implementation
in Pachyderm 1.12. Please take a look at our examples:

- [Spout101](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/spouts/spout101){target=_blank}

- More extensive - Pachyderm's integration of spouts with RabbitMQ: https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/spouts/go-rabbitmq-spout 

### Transactions

Pachyderm transactions enable you to execute multiple
Pachyderm operations simultaneously.

[Use Transactions with Hyperparameter Tuning](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/transactions){target=_blank}

### err_cmd

The `err_cmd` parameter in a Pachyderm pipeline enables
you to specified actions for failed datums. When you do not
need all the datums to be successful for each run of your
pipeline, you can configure this parameter to skip them and
mark the job run as successful.

[Skip Failed Datums in Your Pipeline](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/err_cmd){target=_blank}

## Machine Learning

### Iris flower classification with R, Python, or Julia

The "hello world" of machine learning implemented in Pachyderm.  You can deploy this pipeline using R, Python, or Julia components, where the pipeline includes the training of a SVM, LDA, Decision Tree, or Random Forest model and the subsequent utilization of that model to perform inferences.

[R, Python, or Julia - Iris flower classification](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/ml/iris){target=_blank}

### Sentiment analysis with Neon

This example implements the machine learning template pipeline discussed in [this blog post](https://medium.com/pachyderm-data/sustainable-machine-learning-workflows-8c617dd5506d){target=_blank}.  It trains and utilizes a neural network (implemented in Python using Nervana Neon) to infer the sentiment of movie reviews based on data from IMDB. 

[Neon - Sentiment Analysis](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/ml/neon){target=_blank}

### pix2pix with TensorFlow

If you haven't seen pix2pix, check out [this great demo](https://affinelayer.com/pixsrv/){target=_blank}.  In this example, we implement the training and image translation of the pix2pix model in Pachyderm, so you can generate cat images from edge drawings, day time photos from night time photos, etc.

[TensorFlow - pix2pix](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/ml/tensorflow){target=_blank}

### Distributed Hyperparameter Tuning

This example demonstrates how you can evaluate a model or function in a distributed manner on multiple sets of parameters.  In this particular case, we will evaluate many machine learning models, each configured uses different sets of parameters (aka hyperparameters), and we will output only the best performing model or models.

[Hyperparameter Tuning](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/ml/hyperparameter){target=_blank}

### Spark Example
This example demonstrates integration of Spark with Pachyderm by launching a Spark job on an existing cluster from within a Pachyderm Job. The job uses configuration info that is versioned within Pachyderm, and stores it's reduced result back into a Pachyderm output repo, maintaining full provenance and version history within Pachyderm, while taking advantage of Spark for computation.

[Spark Example](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/spark/pi){target=_blank}

## Integration with Pachyderm
###  S3 gateway 
- Pachyderm - Seldon integration: Version Controlled Models¶

    In these 2 examples, we showcased how we have integrated Pachyderm's end-to-end pipelines,
    leveraging our data lineage capabilities, 
    with [Seldon-Core's deployment platform of ML models](https://www.seldon.io/tech/products/core/){target=_blank}.

    1. In this first simple example, we train a data-driven model using Pachyderm (LogisticRegression on the Iris dataset with sklearn),
    expose the model's artifacts through Pachyderm's [S3 getaway](https://docs.pachyderm.com/latest/reference/s3gateway-api/){target=_blank}, and serve this model in production using Seldon-core. https://github.com/SeldonIO/seldon-core/blob/2.0.x/examples/pachyderm-simple/index.ipynb

        !!! Highlights
            You can trace the model artifact's lineage right back to the version of the data that it was trained on.  

    1. CD for an ML process: In this example, we automate the provisioning of a Seldon deployment using Pachyderm pipelines when new training data enters a Pachyderm repository. 
    https://github.com/SeldonIO/seldon-core/blob/2.0.x/examples/pachyderm-cd4ml/index.ipynb

        !!! Highlights 
            - **Provenance** - The traceability of the model artifact's lineage all the way to the data provides the ability to do post-analysis on models performing poorly.  
            - **Automation** -  A new deployment in production is triggered when new model artifacts are exposed to Pachyderm's [S3 getaway](https://docs.pachyderm.com/latest/reference/s3gateway-api/){target=_blank}.

- Pachyderm - Label Studio

    We have integrated Pachyderm's versioned data backend with [Label Studio](https://labelstud.io/){target=_blank}
    to support versioning datasets and tracking the data lineage of pipelines built off the versioned datasets: https://github.com/pachyderm/examples/tree/master/label-studio
###  Spout 
This is a simple example of using the new implementation of
Pachyderm's spouts with RabbitMQ to process messages and write them to files.
This spout reads messages from a single configurable RabbitMQ queue. 
Please take a look; there is a little more to it, including a fault tolerance mechanism: https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/spouts/go-rabbitmq-spout 
