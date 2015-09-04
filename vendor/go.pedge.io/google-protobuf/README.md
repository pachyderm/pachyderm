[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/go.pedge.io/google-protobuf)

This package compiles all the proto files in [$(which protoc)../include/google/protobuf](https://github.com/google/protobuf/tree/master/src/google/protobuf) into go structs,
and removes cyclical imports. This allows easy use of these commonly-used messages in other repositories.

The version of the proto files is pegged to whatever proto version the [pedge/proto3grpc](https://github.com/peter-edge/proto3grpc) docker image is using.
