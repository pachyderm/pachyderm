# Pachyderm language clients

## Go Client

The Go client is officially supported by the Pachyderm team.  It implements almost all of the functionality that is provided with the `pachctl` CLI tool, and, thus, you can easily integrated operations like `put file` into your applications.

For more info, check out the [godocs](https://godoc.org/github.com/pachyderm/pachyderm/src/client).

**Note** - A compatible version of `grpc` is needed when using the Go client.  You can deduce the compatible version from our [vendor.json](https://github.com/pachyderm/pachyderm/blob/master/src/server/vendor/vendor.json) file, where you will see something like:

```json
		{
			"checksumSHA1": "mEyChIkG797MtkrJQXW8X/qZ0l0=",
			"path": "google.golang.org/grpc",
			"revision": "21f8ed309495401e6fd79b3a9fd549582aed1b4c",
			"revisionTime": "2017-01-27T15:26:01Z"
		},
```

You can then get this version via:

```shell
go get google.golang.org/grpc
cd $GOPATH/src/google.golang.org/grpc
git checkout 21f8ed309495401e6fd79b3a9fd549582aed1b4c
```

### Running Go Examples

The Pachyderm [godocs](https://godoc.org/github.com/pachyderm/pachyderm/src/client) reference
provides examples of how you can use the Go client API. You need to have a running Pachyderm cluster
to run these examples.

Make sure that you use your `pachd_address` in `client.NewFromAddress("<your-pachd-address>:30650")`.
For example, if you are testing on `minikube`, run
`minikube ip` to get this information.

See the [OpenCV Example in Go](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv) for more
information.

## Python Client

The Python client `python-pachyderm` is officially supported by the Pachyderm team. 
It implements most of the functionalities provided with the `pachctl` CLI tool allowing you to easily integrate operations like `create repo`, `put a file,` or `create pipeline` into your python applications.

!!! Note
    Use **python-pachyderm v7.0** with Pachyderm 2.0 and higher. In the [documentation](https://python-pachyderm.readthedocs.io/en/v7.x/), you will find: 
    - All [installation instructions](https://python-pachyderm.readthedocs.io/en/v7.x/getting_started.html#installation) and links to PyPI.
    - A quick ["Hello World" example](https://python-pachyderm.readthedocs.io/en/v7.x/getting_started.html#installation) to get you started.
    - Links to python-pachyderm main Github repository with a [list of useful examples](https://github.com/pachyderm/python-pachyderm/tree/master/examples). 
    - As well as the entire reference API.

## Node Client

Our Javascript client `node-pachyderm` is being **used in production by our Pachyderm Console** to interface with the Pachyderm server. The team officially supports it. Today, we provide all the read operations necessary to our own UI. However, we and have no near-term plans to reach parity with python-pachyderm yet.

Please contact us if you are [interested in contributing](https://github.com/pachyderm/node-pachyderm/blob/main/contributing.md) or ask your questions on our [slack channel](https://pachyderm-users.slack.com/archives/C028ZV066JY).

You will find installations instructions and a first quick overview of how to use the library in our [public repository](https://github.com/pachyderm/node-pachyderm). Check also our [opencv example](https://github.com/pachyderm/node-pachyderm/tree/main/examples/opencv)

## Other languages

Pachyderm uses a simple [protocol buffer API](https://github.com/pachyderm/pachyderm/blob/master/src/pfs/pfs.proto). Protobufs support [a bunch of other languages](https://developers.google.com/protocol-buffers/), any of which can be used to programmatically use Pachyderm. We haven’t built clients for them yet, but it’s not too hard. It’s an easy way to contribute to Pachyderm if you’re looking to get involved.
