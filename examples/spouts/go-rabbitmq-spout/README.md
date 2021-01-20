# Commit messages from RabbitMQ
   
This is a simple example of using spouts with [RabbitMQ](https://www.rabbitmq.com/) to process messages and write them to files.

This example spout connects to a RabbitMQ instance and reads messages from a queue. These messages are written into a single text file 
in the output repository for downstream processing. 

## Prerequisites

The Pachyderm code in this example requires a Pachyderm cluster version 1.12.0 or later and a functioning RabbitMQ deployment. 

## Introduction

RabbitMQ is a very simple messaging system. It is lightweight and easy to deploy, which makes it ideal for particular applications.
While not itself cloud native, it isn't too challenging to stand up a deployment in Kubernetes. If you need a lightweight message queue
and don't need a full scale Kafka cluster, RabbitMQ is one possible alternative depending on your architecture. 

Pachyderm spouts are a way to ingest data into Pachyderm 
by having your code get the data from inside a Pachyderm pipeline.

This is a simple implementation of a Pachyderm version 2 spout, but has additional bells and whistles and hopefully can serve as the basis
to build a more robust spout. 

This spout reads messages from a single configurable RabbitMQ queue. These messages are pushed into a local buffer (go slice)
which is written into a newline delimited file (e.g. NDJSON) when full or at a user configurable flush interval. Every new file creates a new 
commit on the `COMMIT_BRANCH`. After a commit is finalized, all messages read from the RabbitMQ queue are acknowledged at once. This provides
fault tolerance in case the pipeline crashes at any point. You don't need to save your place and you do not need to be concerned about the 
number of consumers (within RabbitMQ's limits, that is). A separate goroutine also reads each commit hash and commits the latest finalized 
commit at a configurable interval (e.g. 60 seconds) to control the rate at which downstream pipelines are triggered. 

### Pachyderm setup

1. If you would simply like to use the prebuilt spout image,
you can simply create the spout with the pachctl command
using the pipeline definition available in the `pipelines` directory

```shell
$ pachctl create pipeline -f pipelines/spout.pipeline.json
```


2. To create your own version of the spout,
you may modify the pipeline file and point it at your own container registry


The Makefile has targets for `create-pipeline` and `update-pipeline`, 
or you may simply make the image with `docker-image`.

### Configuration/Customization

| Variable Name | Description | Default Value |
|---------------|-------------|---------------|
| `PREFETCH` | The prefetch size on RabbitMQ. How many messages will be written into a single file. | 2000   |
| `EXTENSION` | The file extension.                                                                  | ndjson |
| `FLUSH_INTERVAL_MS` | The amount of time to flush messages to a file/commit in milliseconds                | 10000 |
| `SWITCH_INTERVAL_MS` | How often to commit to `master` and trigger a downstream pipeline                   | 60000 |
| `RABBITMQ_HOST`  | The transport endpoint for RabbitMQ (port included) | `rabbitmq.default.svc.cluster.local:5672` |
| `RABBITMQ_USER`  | The username for RabbitMQ | peter |
| `RABBITMQ_PASSWORD` | (Secret) The RabbitMQ password | `rabbitmq-password` |
| `SWITCH_BRANCH` | The branch to switch to periodically | master |
| `COMMIT_BRANCH` | The branch to commit to | staging |
Furthermore, the following command line arguments are available for the rabbitmq spout:
| Flag  | Description |
|-------|-------------|
| -topic | The name of the messaging topic to read from |
| -overwrite | Whether or not to overwrite output |
