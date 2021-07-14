# Examples

## OpenCV Edge Detection

This example does edge detection using OpenCV. This is our canonical starter demo. If you haven't used Pachyderm before, start here. We'll get you started running Pachyderm locally in just a few minutes and processing sample log lines.

[Open CV](https://docs.pachyderm.com/latest/getting_started/beginner_tutorial/)

## Word Count (Map/Reduce)

Word count is basically the "hello world" of distributed computation. This example is great for benchmarking in distributed deployments on large swaths of text data.

[Word Count](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/word_count)

## Periodic Ingress from a Database

This example pipeline executes a query periodically against a MongoDB database outside of Pachyderm.  The results of the query are stored in a corresponding output repository.  This repository could be used to drive additional pipeline stages periodically based on the results of the query.

[Periodic Ingress from MongoDB](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/db)

## Lazy Shuffle pipeline

This example demonstrates how lazy shuffle pipeline i.e. a pipeline that shuffles, combines files without downloading/uploading can be created. These types of pipelines are useful for intermediate processing step that aggregates or rearranges data from one or many sources.

[Lazy Shuffle pipeline](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/shuffle)

## Variant Calling and Joint Genotyping with GATK

This example illustrates the use of GATK in Pachyderm for Germline variant calling and joint genotyping. Each stage of this GATK best practice pipeline can be scaled individually and is automatically triggered as data flows into the top of the pipeline. The example follows [this tutorial](https://drive.google.com/open?id=0BzI1CyccGsZiQ1BONUxfaGhZRGc) from GATK, which includes more details about the various stages.

[GATK - Variant Calling](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/gatk)

## Pachyderm Pipelines

This section lists all the examples that you can run with various
Pachyderm pipelines and special features, such as transactions.

### Joins

A join is a special type of pipeline that enables you to perform
data operations on files with a specific naming pattern.

[Matching files by name pattern](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/joins)

### Spouts

A spout is a special type of pipeline that you can use to ingest
streaming data and perform such operations as sorting, filtering, and other.

* [Email Sentiment Analyzer](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/spouts/EmailSentimentAnalyzer)
* [Commit Messages from a Kafka Queue](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/spouts/go-kafka-spout)
* [Amazon SQS S3 Spout](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/spouts/SQS-S3)

### Transactions

Pachyderm transactions enable you to execute multiple
Pachyderm operations simultaneously.

[Use Transactions with Hyperparameter Tuning](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/transactions)

### err_cmd

The `err_cmd` parameter in a Pachyderm pipeline enables
you to specified actions for failed datums. When you do not
need all the datums to be successful for each run of your
pipeline, you can configure this parameter to skip them and
mark the job run as successful.

[Skip Failed Datums in Your Pipeline](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/err_cmd)

## Machine Learning

### Iris flower classification with R, Python, or Julia

The "hello world" of machine learning implemented in Pachyderm.  You can deploy this pipeline using R, Python, or Julia components, where the pipeline includes the training of a SVM, LDA, Decision Tree, or Random Forest model and the subsequent utilization of that model to perform inferences.

[R, Python, or Julia - Iris flower classification](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/iris)

### Sentiment analysis with Neon

This example implements the machine learning template pipeline discussed in [this blog post](https://medium.com/pachyderm-data/sustainable-machine-learning-workflows-8c617dd5506d#.hhkbsj1dn).  It trains and utilizes a neural network (implemented in Python using Nervana Neon) to infer the sentiment of movie reviews based on data from IMDB. 

[Neon - Sentiment Analysis](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/neon)

### pix2pix with TensorFlow

If you haven't seen pix2pix, check out [this great demo](https://affinelayer.com/pixsrv/).  In this example, we implement the training and image translation of the pix2pix model in Pachyderm, so you can generate cat images from edge drawings, day time photos from night time photos, etc.

[TensorFlow - pix2pix](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/tensorflow)

### Recurrent Neural Network with Tensorflow

Based on [this Tensorflow example](https://www.tensorflow.org/tutorials/recurrent#recurrent-neural-networks), this pipeline generates a new Game of Thrones script using a model trained on existing Game of Thrones scripts.

[Tensorflow - Recurrent Neural Network](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/rnn) 

### Distributed Hyperparameter Tuning

This example demonstrates how you can evaluate a model or function in a distributed manner on multiple sets of parameters.  In this particular case, we will evaluate many machine learning models, each configured uses different sets of parameters (aka hyperparameters), and we will output only the best performing model or models.

[Hyperparameter Tuning](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/hyperparameter)

### Spark Example
This example demonstrates integration of Spark with Pachyderm by launching a Spark job on an existing cluster from within a Pachyderm Job. The job uses configuration info that is versioned within Pachyderm, and stores it's reduced result back into a Pachyderm output repo, maintaining full provenance and version history within Pachyderm, while taking advantage of Spark for computation.

[Spark Example](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/spark/pi)


