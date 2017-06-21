# Pachyderm language clients

## Go Client

The Go client is officially supported by the Pachyderm team.  It implements almost all of the functionality that is provided with the `pachctl` CLI tool, and, thus, you can easily integrated operations like `put-file` into your applications. 

For more info, check out the [godocs](https://godoc.org/github.com/pachyderm/pachyderm/src/client).

## Python Client - `pypachy`

The Python client is a user contributed client that isn't officially maintained by the Pachyderm team.  However, it implements very similar functionality to that available in the Go client or CLI.  

For more info, check out [`pypachy` on GitHub](https://github.com/kalugny/pypachy).

## Scala Client 

Our users are currently working on a Scala client for Pachyderm. Please contact us if you are interested in helping with this or testing it out.

## Other languages

Pachyderm uses a simple [protocol buffer API](https://github.com/pachyderm/pachyderm/blob/master/src/client/pfs/pfs.proto). Protobufs support [a bunch of other languages](https://developers.google.com/protocol-buffers/), any of which can be used to programatically use Pachyderm. We haven’t built clients for them yet, but it’s not too hard. It’s an easy way to contribute to Pachyderm if you’re looking to get involved. 
