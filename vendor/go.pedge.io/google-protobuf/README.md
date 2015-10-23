[![CircleCI](https://circleci.com/gh/peter-edge/go-google-protobuf/tree/master.png)](https://circleci.com/gh/peter-edge/go-google-protobuf/tree/master)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/google-protobuf)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/go-google-protobuf/blob/master/LICENSE)

This package compiles all the proto files in [$(which protoc)../include/google/protobuf](https://github.com/google/protobuf/tree/master/src/google/protobuf) into golang structs,
and removes cyclical imports. This allows easy use of these commonly-used messages in other repositories.
