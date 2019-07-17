# Spout

A spout is a type of pipeline that processes
streaming data that is continuously generated from
multiple sources. Generally, you use spouts for
situations when the interval between new data generation
is large, but the latency requirement to start the
processing is short. Therefore, a regular pipeline
with periodic processing might not be an optimal solution.

Examples of streaming data include a message queue,
a database transactions log, event notifications,
and others. In spouts, your code runs continuously and writes the
results to the pipeline's output location `pfs/out`.
Every time you create a complete `.tar` archive,
Pachyderm creates a new commit and triggers the pipeline to
process it.

One main difference from regular pipelines is that
you cannot specify input in the spout pipeline.

Another important aspect is that in spouts, `pfs/out` is
a *named pipeline*, or *First in, First Out* (FIFO), and is not
a directory like in standard pipelines. Unlike
the traditional and familiar to most Linux users
pipe, a *named pipe* enables two system processes to access
the pipe simultaneously and gives one of the processes read-only and the other
process write-only access. Therefore, the two processes can simultaneously
read and write to the same pipe.

To create a spout pipeline, you need the following items:

* A source of streaming data
* Docker container with your spout code that reads from the data source
* A spout pipeline specification file that uses the container

A spout code performs the following actions:

1. Connects to the specified streaming data source.
1. Reads the data from the streaming data source.
1. Packages the data into a `tar` stream.
1. Writes the `tar` stream into the `pfs/out` pipeline.

A minimum spout specification must include the following
parameters:

| Parameter   | Description |
| ----------- | ----------- |
| `name`      | The name of your data pipeline and the output repository. You can set an <br> arbitrary name that is meaningful to the code you want to run. |
| `transform` | Specifies the code that you want to run against your data, such as a Python <br> or Go script. Also, specifies a Docker image that you want to use to run that script. |
| `overwrite` | Specifies whether to overwrite the existing content of the file <br> from previous commits or previous calls to the `put file` command <br> within this commit. The default value is `false`. |

The following text is an example of a minimum specification:

**Note:** The `env` property is an optional argument. You can define
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
  },
  "env": {
    "HOST": "kafkahost",
    "TOPIC": "mytopic",
    "PORT": "9092"
  },
  "spout": {
    "overwrite": false
  }
}
```
