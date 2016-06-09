[![CircleCI](https://circleci.com/gh/peter-edge/lion-go/tree/master.png)](https://circleci.com/gh/peter-edge/lion-go/tree/master)
[![Go Report Card](http://goreportcard.com/badge/peter-edge/lion-go)](http://goreportcard.com/report/peter-edge/lion-go)
[![GoDoc](http://img.shields.io/badge/GoDoc-Reference-blue.svg)](https://godoc.org/go.pedge.io/lion)
[![MIT License](http://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/peter-edge/lion-go/blob/master/LICENSE)

```shell
go get go.pedge.io/lion
```

Documentation coming soon, have not got around to it yet.

All public types are in [lion.go](lion.go) and [lion_level.go](lion_level.go).

In sub packages, are public types are in `nameofpackage.go`, and `nameofpackage.pb.go` in the case of `protolion`. Some of them need to be renamed,
ie [syslog/syslog.go](syslog/syslog.go) should be `syslog/sysloglion.go`, this is a holdover from the previous iteration of lion, called protolog.

Documentation in the form of me copying and pasting part of a Slack convo I had, yeahhhh:

```
It needs about 20 hours to document
But it's super configurable, optionally "high performance" (I mean, not really high performance, but higher than existing widely-used loggers other than maybe glog), and with plugins
It's literally almost completely compatible with logrus
And it benchmarks way better
The real original power of the design is protolion
It logs only protocol buffers, and can write them to a file and they can be read back
See the tailing and testing package
The envlion package sets it up with environment variables
The gcloudlion package pushes to gcloud logger
The sysloglion package pushes to syslog
The ginlion package has wrappers for the gin web framework
The goal is to be able to write protobufs quickly to a file, then tail it, and do whatever complicated stuff you need to do
There's a wrapper for go-kit too
But ya, see the WithField, WithFields, WithContext (protolion), and WithKeyValues methods
And also see LogDebug, LogInfo, which return a logger that does no processing of WithFields etc if the level is not at debug or info (useful for better performance)
The base lion package is almost drop-in-able for logrus, except I don't duplicate Info/Infoln, etc, only Infoln (Info is reserved for protobuf)
And you still get the rest of the power
The main thing is once I have it set up, it's plug and play to have structured, high performance logging with no loss of logs (if you use an in memory buffer, for example), and then integration with i.e. rsyslog to read back the logs and send them wherever
And if you don't want any of that, it's a super simple logger to use :)
But that's the goal - easy to start, lots of potential
```
