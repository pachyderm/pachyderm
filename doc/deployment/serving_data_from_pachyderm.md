# Serving Data From Pachyderm

Once you have the output data from your DAG of pipelines, you may want to provide the results to some other part of your application.

To do so, use a Pachyderm Service

## Pachyderm Service

A Pachyderm Service is identical to a Job, except:

- it is long running (whereas a Job runs to completion)
- it is accessible from outside the container or cluster on a specific port

That means that you can now run an actual service (e.g. apache / a dashboard / a REST API / etc) on a container that has access to your data at `/pfs/...`

## Example Use Cases

Some common examples include:

- Using a service to attach a Jupyter Notebook to your DAG (common for developing or debugging a pipeline)
- Using a simple file server to serve the raw data from Pachyderm to something outside the cluster
- Using a postgres service to digest and serve the Pachyderm data in a queryable format

## Example Specification

For the complete example, refer to the [jupyter notebook](https://github.com/pachyderm/pachyderm/tree/master/doc/examples/jupyter_notebook) example.

Since a Pachyderm Service is based off of the Pachyderm Job primitive, the specification looks very similar:

```
{
	"service" : {
		"internal_port": 8888,
		"external_port": 30888
	},
	"transform": {
        "image": "pachyderm_jupyter",
        "cmd": [ "sh" ],
        "stdin": [
        	"/opt/conda/bin/jupyter notebook"
		],
	},
 	"parallelism_spec": {
		"strategy": "CONSTANT",
		"constant": 1
	},
    "inputs": [
		{
			"commit": {
            	"repo": {
            		"name": "foo"
            	},
	       		"id": "master/0"
			}
    	}
	]
}
```

The only difference from a normal Job specification is the first three lines. We specify it's a service by providing the `service` field and specifying at least an `internal_port`.

The `internal_port` specifies the container port to expose. If you're planning on consuming this service within the cluster, you can just specify an internal port.

The `external_port` specifies the node port to expose. If you're planning on consuming this service outside the cluster, you'll need to specify an external port in addition to the internal port. You may need to specify firewall rules with your provider to access this port ( e.g. [howto for gcloud](https://cloud.google.com/sdk/gcloud/reference/compute/firewall-rules/create)).


## Writing the Service

Just like a normal job, when your service runs, any inputs will be provided at `/pfs/...`

For example, for the above specification, the folder `/pfs/foo` would be mounted at the specific commit `master/0` so your service code could access any data from the `foo` repository at `master/0` commit locally.

If you need access to multiple data repos, just provide more inputs in the specification.

## Coming Soon

Soon we'll upgrade the Pachyderm Service to have a Pipeline analog. This will allow you to have a long running Pachyderm Service with data mounted locally. The only difference will be that as new commits become available, the service will be updated with the latest data.

