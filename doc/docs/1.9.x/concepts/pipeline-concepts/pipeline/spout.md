# Spout

!!! note
    This document addresses spouts 1.0 implementation
    prior to Pachyderm 1.12.
    The implementation in spouts 2.0 is significantly different.
    We recommend upgrading 
    to the latest version
    of Pachyderm
    and using the spouts 2.0 implementation.
    
A spout is a type of pipeline that ingests
streaming data. Generally, you use spouts for
situations when the interval between new data generation
is large or sporadic, but the latency requirement to start the
processing is short. Therefore, a regular pipeline
with a cron input that polls for new data
might not be an optimal solution.

Examples of streaming data include a message queue,
a database transactions log, event notifications,
and others. In spouts, your code runs continuously and writes the
results to the pipeline's output location, `pfs/out`.
Every time you create a complete `.tar` archive,
Pachyderm creates a new commit and triggers the pipeline to
process it.

One main difference from regular pipelines is that
spouts ingest their data from outside sources. Therefore, they
do not take an input.

Another important aspect is that in spouts, `pfs/out` is
a *named pipe*, or *First in, First Out* (FIFO), and is not
a directory like in standard pipelines. Unlike
the traditional pipe, that is familiar to most Linux users,
a *named pipe* enables two system processes to access
the pipe simultaneously and gives one of the processes read-only and the other
process write-only access. Therefore, the two processes can simultaneously
read and write to the same pipe.

To create a spout pipeline, you need the following items:

* A source of streaming data
* A Docker container with your spout code that reads from the data source
* A spout pipeline specification file that uses the container

Your spout code performs the following actions:

1. Connects to the specified streaming data source.
1. Opens `/pfs/out` as a named pipe.
1. Reads the data from the streaming data source.
1. Packages the data into a `tar` stream.
1. Writes the `tar` stream into the `pfs/out` pipe. In case of transient
errors produced by closing a previous write to the pipe, retries the write
operation.
1. Closes the `tar` stream and connection to `/pfs/out`, which produces the
commit.

A minimum spout specification must include the following
parameters:

| Parameter   | Description |
| ----------- | ----------- |
| `name`      | The name of your data pipeline and the output repository. You can set an <br> arbitrary name that is meaningful to the code you want to run. |
| `transform` | Specifies the code that you want to run against your data, such as a Python <br> or Go script. Also, it specifies a Docker image that you want to use to run that script. |
| `overwrite` | (Optional) Specifies whether to overwrite the existing content <br> of the file from previous commits or previous calls to the <br> `put file` command  within this commit. The default value is `false`. |

The following text is an example of a minimum specification:

!!! note
    The `env` property is an optional argument. You can define
    your data stream source from within the container in which you run
    your script. For simplicity, in this example, `env` specifies the
    source of the Kafka host.

```
{
  "pipeline": {
    "name": "my-spout"
  },
  "transform": {
    "cmd": [ "go", "run", "./main.go" ],
    "image": "myaccount/myimage:0.1"
    "env": {
        "HOST": "kafkahost",
        "TOPIC": "mytopic",
        "PORT": "9092"
    },
  },
  "spout": {
    "overwrite": false
  }
}
```
