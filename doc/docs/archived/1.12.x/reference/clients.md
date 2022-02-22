# Pachyderm language clients

`pachctl` is the command-line tool you use 
to interact with a Pachyderm cluster in your terminal. 
However,  external applications might need to
interact with Pachyderm directly through our APIs.

In this case, Pachyderm offers language specific SDKs in Go, Python, and JS.

## Go Client

The Pachyderm team officially supports the Go client. It implements most of the functionality that is provided with the `pachctl` CLI tool.

For more info, check out the [godocs](https://pkg.go.dev/github.com/pachyderm/pachyderm@v1.12.5/src/client).

!!! Attention
     A compatible version of `gRPC` is needed when using the Go client.  You can identify the compatible version by searching for the version number next to `replace google.golang.org/grpc => google.golang.org/grpc` in https://github.com/pachyderm/pachyderm/blob/master/go.mod then:


	```shell
	go get google.golang.org/grpc
	cd $GOPATH/src/google.golang.org/grpc
	git checkout v1.29.1
	```    
### Running Go Examples

The Pachyderm [godocs](https://pkg.go.dev/github.com/pachyderm/pachyderm@v1.12.5/src/client) reference
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
     Use **python-pachyderm v6.x** with Pachyderm 1.13.x and under. 

You will find all available modules and functions in [`python-pachyderm` API documentation](https://python-pachyderm.readthedocs.io/en/v6.x/).

!!! Attention "Get started"
	 - Take a look at `python-pachyderm` main Github repository and the [list of useful examples](https://github.com/pachyderm/python-pachyderm/tree/v6.x/examples) that should get you started.
	 - [Read about the few operations you will need to know to start using python-pachyderm](../../how-tos/use-pachyderm-ide/using-pachyderm-ide/).

## Other languages

Pachyderm uses a simple [protocol buffer API](https://github.com/pachyderm/pachyderm/blob/master/src/pfs/pfs.proto). Protobufs support [other languages](https://developers.google.com/protocol-buffers/), any of which can be used to programmatically use Pachyderm. We haven’t built clients for them yet. It’s an easy way to contribute to Pachyderm if you’re looking to get involved.
