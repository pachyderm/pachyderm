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

* syslog writer
* journal writer?
* colors in terminals
* better text formatting/options
* third-party logs integration
* performance improvements/testing
* documentation

### Build requirements

##### Protocol Buffers

I'm not using the pedge/proto3grpc docker image to build the protos right now, so you need proto3 installed locally.

Mac:

```shell
brew update
brew install protobuf --c++11 --devel
```

Linux:

```
wget https://codeload.github.com/google/protobuf/tar.gz/v3.0.0-alpha-3.1 && \
  tar xvzf v3.0.0-alpha-3.1 && \
  rm v3.0.0-alpha-3.1 && \
  cd protobuf-3.0.0-alpha-3.1 && \
  ./autogen.sh && \
  ./configure --prefix=/usr && \
  make && \
  make check && \
  sudo make install && \
  cd - && \
  rm -rf protobuf-3.0.0-alpha-3.1
```

##### Golang

I would recommend Go v1.4.2, check with `go version`.

Mac:

```shell
brew update
brew reinstall go --with-cc-common # you probably want the cross-compilation stuff at some point
```

Linux:

```shell
sudo su
curl -sSL https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local
echo 'export PATH=${PATH}:/usr/local/go/bin' >> '/etc/profile'
echo 'export GOROOT=/usr/local/go' >> '/etc/profile'
exit # out of sudo su
echo 'mkdir -p ${HOME}/go/bin' >> ${HOME}/.bash_aliases
echo 'echo GOPATH=${HOME}/go' >> ${HOME}/.bash_aliases
echo 'export PATH=${PATH}:${HOME}/go/bin' >> ${HOME}/.bash_aliases
source ${HOME}/.bashrc
```
