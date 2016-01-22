[![CircleCI](https://circleci.com/gh/peter-edge/google-protobuf-go/tree/master.png)](https://circleci.com/gh/peter-edge/google-protobuf-go/tree/master)
[![Go Report Card](http://goreportcard.com/badge/peter-edge/google-protobuf-go)](http://goreportcard.com/report/peter-edge/google-protobuf-go)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/google-protobuf)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/google-protobuf-go/blob/master/LICENSE)

This package compiles all the proto files in https://github.com/google/protobuf/tree/master/src/google/protobuf into golang structs. This allows easy use of these commonly-used messages in other repositories.

This package does not include descriptor.proto, this is in https://github.com/golang/protobuf/tree/master/protoc-gen-go/descriptor.
