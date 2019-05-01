# Spouts

## Introduction

Spouts are a way to get streaming data from any source into Pachyderm.
To create a spout, you need three things

1. A source of streaming data, such as Kafka, nifi, rabbitMQ, etc.
2. A containerized server.
3. A spout pipeline specification file that uses the container.

The containerized server will do four things:

- connect to your source of streaming data
- read the data
- package it into files 
  in a `tar` stream
- write that `tar` stream
  to the named pipe `/pfs/out`


In this document, 
we'll take you through writing the server 
with an example using Go code
and writing the spout pipeline specification.

## Creating the containerized server

To create the server,
you'll need access to client libraries
for your streaming data source,
a library that can write the `tar` archive format
(such as [Go's tar package](https://golang.org/pkg/archive/tar/) 
or [Python's tarfile module](https://docs.python.org/3.7/library/tarfile.html)),
and requirements for how you would like to batch your data.
For the purposes of this document, 
we'll assume each message in the stream
is a single file.

You need that `tar` library because,
in spouts,
`/pfs/out` is a named pipe. 
This is different than in pipelines, 
where `/pfs/out` is a directory.

In the example below, 
written in Go
and taken from the [kafka example](https://github.com/pachyderm/pachyderm/tree/master/examples/kafka)
in the Pachyderm repo, 
we'll go through every step you need to take.

### Import necessary libraries

We'll import the libraries necessary for creating a `tar` archive data stream
and connecting to our Kafka data source.

```go
package main

import (
	"archive/tar"
	"context"
	"os"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

```
### Parameterizing connection information
It's a good idea to get parameters
for connecting to your data source 
from the environment
or command-line parameters. 
That way you can connect to new data sources
using the same container code
without recompiling
by just setting appropriate parameters
in the pipeline specification.
```Go
func main() {
	// Get the connection info from the ENV vars
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	topic := os.Getenv("TOPIC")
```

### Connect to the streaming data source

We're creating an object
that can be used to read from our data source.

That `defer` statement is the Go way to guarantee
that the open file will be closed
after `main()` returns.
That is, 
`reader.Close()` won't be executed
until after `main()` is finished.
(This is a common idiom in Go;
to `defer` the close of a resource
right after you open it.)


```Go
	// And create a new kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{host + ":" + port},
		Topic:    topic,
		MinBytes: 10e1,
		MaxBytes: 10e6,
	})
	defer reader.Close()
```

### Open /pfs/out for writing

We're opening the named pipe`/pfs/out` for writing
so we can send it a `tar` archive stream
with the file
we want to output.
Note that the named pipe has to be opened
with write-only permissions.


```Go
    // Open the /pfs/out pipe with write only permissons (the pachyderm spout will be reading at the other end of this)
	// Note: it won't work if you try to open this with read, or read/write permissions
	out, err := os.OpenFile("/pfs/out", os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer out.Close()
```

### Write the outer file loop
Here we open a tar stream 
into the directory we opened above,
so that Pachyderm can place files
into the output repo.
For clarity's sake,
we'll omit the message-processing loop
inside this file loop.
The `err` variable is used
in the message-processing loop
for errors reading from the stream
and writing to the directory.
The stream is opened 
at the top of the loop
and should be closed 
at the bottom.
In this case, 
it'll be closed 
after a message is processed.

```Go
	// this is the file loop
	for {
		if err := func() error {
			tw := tar.NewWriter(out)
			defer tw.Close()
			// this is the message loop
			for {

            // ...omitted
            
			}
		}(); err != nil {
			panic(err)
		}
	}
```

### Create the message processing loop

If you have trouble following this Go code,
just read the text
to get an idea of what you need to do.

First, 
we read a message
from our Kafka queue.
Note the use of a 5-second timeout
on the read.

```Go
			// this is the message loop
            for {
				// read a message
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				m, err := reader.ReadMessage(ctx)
				if err != nil {
					return nil
                }
```

Then we create a filename
and write that to the `tar` stream
we opened at the beginning
of the file loop.
```Go
                // give it a unique name
				name := topic + time.Now().Format(time.RFC3339Nano)

                // write the header
				for err = tw.WriteHeader(&tar.Header{
					Name: name,
					Mode: 0600,
					Size: int64(len(m.Value)),
				}); err != nil; {
					if !strings.Contains(err.Error(), "broken pipe") {
						return err
					}
					// if there's a broken pipe, just give it some time to get ready for the next message
					time.Sleep(5 * time.Millisecond)
				}
```

Then we write the actual message 
as the contents of the file.
Note the use of a timeout
in case the named pipe is broken 
for some reason.
In this example,
we're simply moving on to the next message.
A production system would likely have logging
and queue operations
so no messages get missed!
```Go
                
				// and the message
                              
				for _, err = tw.Write(m.Value); err != nil; {
					if !strings.Contains(err.Error(), "broken pipe") {
						return err
					}
					// if there's a broken pipe, just give it some time to get ready for the next message
					time.Sleep(5 * time.Millisecond)
				}
				return nil
			}
```

That's the rough outline of operations
for processing data in queues
and writing it to a spout
in a server.

### Create the container for the server
Once you have containerized your server,
you can place it in a Pachyderm spout
by writing an appropriate pipeline specification.

## Writing the spout pipeline specification

A spout is defined using a similar spec
to a pipeline, 
and created using the `pachctl create pipeline` command.

In our example above,
we created a server using Go.
The Dockerfile for creating a container
with that server in it is in the [kafka example](https://github.com/pachyderm/pachyderm/tree/master/examples/kafka).

Assuming that's been done,
we might define the specification 
for the spout
something like this.

```json
{
  "pipeline": {
    "name": "my-spout"
  },
  "transform": {
    "cmd": [ "go", "run", "./main.go" ],
    "image": "myaccount/myimage"
  },
  "env": {
    "HOST": "kafkahost",
    "TOPIC": "mytopic",
    "PORT": "9092"
  },
  "spout": true
}
```

We would then use `pachctl create pipeline -f my-spout.json` to install the spout.
It would begin processing messages
and placing them in the `my-spout` repo.
