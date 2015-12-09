[![CircleCI](https://circleci.com/gh/peter-edge/go-protolog/tree/master.png)](https://circleci.com/gh/peter-edge/go-protolog/tree/master)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/protolog)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/go-protolog/blob/master/LICENSE)

```shell
go get go.pedge.io/protolog
```

Initial beta release coming soon - I need to do one more pass on this and then document all the functionality.

* Structured logging with Protocol buffers
* Child of https://github.com/peter-edge/go-ledge
* Some compatibility with existing libraries (specifically logrus and glog)
* Two-way serialization - write logs somewhere, read them back, language independent

### Where to poke around

* `protolog.go`: all the public definiions
* `protolog.proto`: the protos that are serialized over the wire
* `testing*`: test compilation of proto definitions
* `benchmark*`: more compilation of proto definitions, and benchmarks
* `make test`: will print out some logs with the default text marshaller
* `make bench`: some basic benchmarks

### TODO

* journal writer?
* colors in terminals
* better text formatting/options
* third-party logs integration
* performance improvements/testing
* documentation
