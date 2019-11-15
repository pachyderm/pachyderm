# Commit messages from a Kafka queue

This is a simple example of using spouts with [Kafka](https://kafka.apache.org) to process messages and write them to files.


This example spout connects to a Kafka queue and reads a topic.
The spout then writes each message in the topic to a file named by the topic and offset. 
It uses Kafka group IDs to maintain a cursor into the offset in the topic, 
making it resilient to restarts.

## Introduction

Apache® Kafka is a distributed streaming platform
that is used in a variety of applications to provide communications between microservices.  
Many Pachyderm users use Kafka to ingest data from legacy data sources using Pachyderm spouts.

Pachyderm spouts are a way to ingest data into Pachyderm 
by having your code get the data from inside a Pachyderm pipeline.

This is the simplest possible implementation of a Pachyderm spout using Kafka to ingest data.
The data ingested is simply the message posted to Kafka.
The filename is derived from the Kafka topic name and the message's offset in the topic.
You should be able to easily adapt it to your needs.

## Setup

This example includes a pre-configured Kafka cluster that you can deploy on Amazon EKS,
adapted from [Craig Johnston's excellent blog post](https://imti.co/kafka-kubernetes/).
If you'd like to adapt it for your own cluster on GCP, Azure, or on-premises Kubernetes deployment,
the file 001-storage-class.yaml is probably the only thing you'd need to change.
You can replace the parameters and provisioner with the appropriate one for your environment.

If you already have a Kafka cluster setup, you may skip step 1 of the Kafka setup.

To correctly build the Docker container from source, 
it will be necessary to have the Pachyderm source repo structure around this example.
It depends on the directory `../../../vendor/github.com/segmentio/kafka-go`,
relative to this one,
containing the correct code.
You can, of course, set up your Go development environment to achieve the same result.

### Kafka setup

1. In the directory `additional_manifests`, 
you'll find a numbered sequence of Kubernetes manifests for creating a fully-functioning Kafka deployment.
You can use the makefile target `make kafka`,
which will deploy a kafka cluster a `kafka` namespace, created in the first step.
If you'd like to see the order in which the manifests will be loaded into Kubernetes,
run the command
```sh
make -n kafka
```
You can confirm that the kafka cluster is running properly by checking to see if all the pods are running.
```sh
$ kubectl get pods -n kafka
NAME                READY   STATUS    RESTARTS   AGE
kafka-0             1/1     Running   0          3d19h
kafka-1             1/1     Running   0          3d19h
kafka-2             1/1     Running   0          3d19h
kafka-test-client   1/1     Running   0          3d19h
kafka-zookeeper-0   1/1     Running   0          3d19h
kafka-zookeeper-1   1/1     Running   0          3d19h
kafka-zookeeper-2   1/1     Running   0          3d19h
```
2. Once the Kafka cluster is running, create the topic you'd like to consume messages from.
The example is configured to look for a topic called `test-topic`.
You may modify the Makefile to use another topic name, of course.
To use the example's Kafka environment,
you may use the following command to create the topic:
```sh
$ kubectl -n kafka exec kafka-test-client -- /usr/bin/kafka-topics --zookeeper \
      kafka-zookeeper.kafka:2181 --topic test --create \
      --partitions 1 --replication-factor 1
Created topic "test".
```
Note that the command is using Kubernetes DNS names to specify the Kafka zookeeper service,
`kafka-zookeeper.kafka`.
3. You can start populating the topic with data using the `kafka-console-producer` command.
It provides you with a `>` prompt for entering data,
delimited by lines for each offset into the topic.
In the example below, the messages at offset 0 is `yo`, 
at offset 1, `man`,
and so on.
Data entry is completed with an end-of-file character,
`Control-d` in most shells.
```sh
$ kubectl -n kafka exec -ti kafka-test-client --  /usr/bin/kafka-console-producer \
   --broker-list kafka.kafka:9092 --topic test 
>yo 
>man
>this 
>is so
>cool!!
>
```
4. You can see if the data has been added to the topic with the `kafka-console-consumer` command.
In the example below,
the session was terminated with `Control-C` keystrokes.
```sh
$ kubectl -n kafka exec -ti kafka-test-client -- /usr/bin/kafka-console-consumer 
   --bootstrap-server kafka:9092 --topic test --from-beginning
yo
man
this
is so
cool!!
^CProcessed a total of 5 messages
command terminated with exit code 130
```
### Pachyderm setup

This guide assumes that you already have a Pachyderm cluster running and have configured `pachctl` to talk to the cluster and `kubectl` to talk to Kubernetes.
[Installation instructions can be found here](http://pachyderm.readthedocs.io/en/stable/getting_started/local_installation.html).

1. If you would simply like to use the prebuilt spout image,
you can simply create the spout with the pachctl command
using the pipeline definition available in the `pipelines` directory
```
$ pachctl create pipeline -f pipelines/kafka_spout.json
```

2. To create your own version of the spout,
you may modify the Makefile to use your own Dockerhub account, tag and version
by changing these variables accordingly
```
CONTAINER_VERSION := 1.9.8
DOCKER_ACCOUNT := pachyderm
CONTAINER_NAME := kafka_spout
```
The Makefile has targets for `create-dag` and `update-dag`, 
or you may simply make the image with `docker-image`.

3. Once the spout is running, 
if the `VERBOSE_LOGGING` variable is set to anything other than `false`,
you will see verbose logging in the `kafka_spout` pipeline logs.
```sh
$ pachctl logs -p kafka_spout -f
creating new kafka reader for kafka.kafka:9092 with topic 'test_topic' and group 'test_group'
reading kafka queue.
opening named pipe /pfs/out.
opening tarstream
processing header for topic test_topic @ offset 0
processing data for topic  test_topic @ offset 0
closing tarstream.
closing named pipe /pfs/out.
cleaning up context.
reading kafka queue.
opening named pipe /pfs/out.
opening tarstream
processing header for topic test_topic @ offset 1
processing data for topic  test_topic @ offset 1
closing tarstream.
closing named pipe /pfs/out.
cleaning up context.
reading kafka queue.
opening named pipe /pfs/out.
opening tarstream
processing header for topic test_topic @ offset 2
processing data for topic  test_topic @ offset 2
closing tarstream.
closing named pipe /pfs/out.
cleaning up context.
reading kafka queue.
opening named pipe /pfs/out.
opening tarstream
processing header for topic test_topic @ offset 3
processing data for topic  test_topic @ offset 3
closing tarstream.
closing named pipe /pfs/out.
cleaning up context.
reading kafka queue.
opening named pipe /pfs/out.
opening tarstream
processing header for topic test_topic @ offset 4
processing data for topic  test_topic @ offset 4
closing tarstream.
closing named pipe /pfs/out.
...
```
And you will see the message files in the `kafka_spout` repo
```sh
$ pachctl list file kafka_spout@master
NAME          TYPE SIZE 
/test_topic-0 file 2B   
/test_topic-1 file 3B   
/test_topic-2 file 4B   
/test_topic-3 file 5B   
/test_topic-4 file 6B   
```
## Pipelines

### kafka_spout

The file `source/main.go` contains a simple Pachyderm spout that processes messages from Kafka,
saving them to files in a Pachyderm repo named for the topic and message offset.

It is configurable via environment variables and command-line flags. 
Flags override environment variable settings.
If your Go development environment is set up correctly,
you can see the settings by running the command:
```
$ go run source/main.go --help
Usage of /var/folders/xl/xtvj4xtx0tv1llxcnbvlmwc40000gq/T/go-build997659573/b001/exe/main:
  -kafka_group_id string
    	the Kafka group for maintaining offset state (default "test")
  -kafka_host string
    	the hostname of the Kafka broker (default "kafka.kafka")
  -kafka_port string
    	the port of the Kafka broker (default "9092")
  -kafka_timeout int
    	the timeout in seconds for reading messages from the Kafka queue (default 5)
  -kafka_topic string
    	the Kafka topic for messages (default "test")
  -named_pipe string
    	the named pipe for the spout (default "/pfs/out")
  -v	verbose logging
exit status 2
```
The environment variables are as shown 
in this excerpt from the `pipelines/kafka_spout.pipeline` file:
```sh
            "KAFKA_HOST": "kafka.kafka",
            "KAFKA_PORT": "9092",
            "KAFKA_TOPIC": "test_topic",
            "KAFKA_GROUP_ID": "test_group",
            "KAFKA_TIMEOUT": "5",
            "NAMED_PIPE": "/pfs/out",
            "VERBOSE_LOGGING": "false"
```

