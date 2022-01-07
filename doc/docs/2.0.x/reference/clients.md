# Pachyderm language clients

`pachctl` is the command-line tool you use 
to interact with a Pachyderm cluster in your terminal. 
However,  external applications might need to
interact with Pachyderm directly through our APIs.

In this case, Pachyderm offers language specific SDKs in Go, Python, and JS.

## Go Client

The Pachyderm team officially supports the Go client. It implements most of the functionality that is provided with the `pachctl` CLI tool.

For more info, check out the [godocs](https://pkg.go.dev/github.com/pachyderm/pachyderm/v2/src/client){target=_blank}.

!!! Attention
     A compatible version of `gRPC` is needed when using the Go client.  You can identify the compatible version by searching for the version number next to `replace google.golang.org/grpc => google.golang.org/grpc` in https://github.com/pachyderm/pachyderm/blob/master/go.mod then:


	```shell
	go get google.golang.org/grpc
	cd $GOPATH/src/google.golang.org/grpc
	git checkout v1.29.1
	```
     
### Running Go Examples

The Pachyderm [godocs](https://pkg.go.dev/github.com/pachyderm/pachyderm/v2/src/client){target=_blank} reference
provides examples of how you can use the Go client API. You need to have a running Pachyderm cluster
to run these examples.

Make sure that you use your `pachd_address` in `client.NewFromAddress("<your-pachd-address>:30650")`.
For example, if you are testing on `minikube`, run
`minikube ip` to get this information.

See the [OpenCV Example in Go](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv){target=_blank} for more
information.

## Python Client

The Python client `python-pachyderm` is officially supported by the Pachyderm team. 
It implements most of the functionalities provided with the `pachctl` CLI tool allowing you to easily integrate operations like `create repo`, `put a file,` or `create pipeline` into your python applications.

!!! Note
     Use **python-pachyderm v7.0** with Pachyderm 2.0 and higher. 

You will find all you need to get you started or dive into the details of the available modules and functions in the [API documentation](https://python-pachyderm.readthedocs.io/en/v7.1.x/){target=_blank}, namely:

- The [installation instructions](https://python-pachyderm.readthedocs.io/en/v7.1.x/getting_started.html#installation){target=_blank} and links to PyPI.
- A quick ["Hello World" example](https://python-pachyderm.readthedocs.io/en/v7.1.x/getting_started.html#installation){target=_blank} to jumpstart your understanding of the API.
- Links to python-pachyderm main Github repository with a [list of useful examples](https://github.com/pachyderm/python-pachyderm/tree/v7.1.x/examples){target=_blank}. 
- As well as the entire **reference API**.

## Node Client

Our Javascript client `node-pachyderm` is a library officially supported by Pachyderm and **used in production by Pachyderm Console**.  

Today, we provide only read operations as shown in Console. Over time, we will add additional functionality to the SDK. However, there are no near-term plans to reach parity with python-pachyderm yet.

Please get in touch with us if you are [interested in contributing](https://github.com/pachyderm/node-pachyderm/blob/main/contributing.md){target=_blank} or ask your questions on our dedicated [slack channel](https://pachyderm-users.slack.com/archives/C028ZV066JY){target=_blank}.

You will find installations instructions and a first quick overview of how to use the library in our [public repository](https://github.com/pachyderm/node-pachyderm){target=_blank}. 
Check also our [opencv example](https://github.com/pachyderm/node-pachyderm/tree/main/examples/opencv){target=_blank}.

## Other languages

Pachyderm uses a simple [protocol buffer API](https://github.com/pachyderm/pachyderm/blob/master/src/pfs/pfs.proto){target=_blank}. Protobufs support [other languages](https://developers.google.com/protocol-buffers/){target=_blank}, any of which can be used to programmatically use Pachyderm. We have not built clients for them yet. It is an easy way to contribute to Pachyderm if you are looking to get involved.
