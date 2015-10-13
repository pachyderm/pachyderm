# protolog

[![CircleCI](https://circleci.com/gh/peter-edge/go-protolog/tree/master.png)](https://circleci.com/gh/peter-edge/go-protolog/tree/master)

```shell
go get go.pedge.io/protolog
```

* Structured logging with Protocol buffers
* Child of https://github.com/peter-edge/go-ledge
* Some compatibility with existing libraries (specifically logrus and glog)
* Two-way serialization - write logs somewhere, read them back, language independent
* protoc-gen-protolog: generate mapping to struct constructors from proto definitions, no more ledge.Specification or reflection

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
