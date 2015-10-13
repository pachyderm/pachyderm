[![CircleCI](https://circleci.com/gh/peter-edge/go-google-protobuf/tree/master.png)](https://circleci.com/gh/peter-edge/go-google-protobuf/tree/master)
[![API Documentation](http://img.shields.io/badge/api-Godoc-blue.svg?style=flat-square)](https://godoc.org/go.pedge.io/google-protobuf)

This package compiles all the proto files in [$(which protoc)../include/google/protobuf](https://github.com/google/protobuf/tree/master/src/google/protobuf) into golang structs,
and removes cyclical imports. This allows easy use of these commonly-used messages in other repositories.
