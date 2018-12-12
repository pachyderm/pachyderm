# Setup For contributors

## General requirements

First, go through the general local installation instructions [here](http://docs.pachyderm.io/en/latest/getting_started/local_installation.html). Additionally, make sure you have the following installed:

- etcd
- golang 1.11+
- docker
- [jq](https://stedolan.github.io/jq/)
- [pv](http://ivarch.com/programs/pv.shtml)

## Bash helpers

To stay up to date, we recommend doing the following.

First clone the code:

    cd $GOPATH/src
    mkdir -p github.com/pachyderm
    cd github.com/pachyderm
    git clone git@github.com:pachyderm/pachyderm

Then update your `~/.bash_profile` by adding the line:

    source $GOPATH/src/github.com/pachyderm/pachyderm/etc/contributing/bash_helpers

And you'll stay up to date!

## Protocol buffers

Install protoc 3+. On macOS, this can be done via `brew install protobuf`. Then installed protoc-gen-go by doing:

    go get -u -v github.com/golang/protobuf/proto
    go get -u -v github.com/golang/protobuf/protoc-gen-go
    go get -u -v github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway

## Special macOS configuration

### File Descriptor Limit

If you're running tests locally, you'll need to up your file descriptor limit. To do this, first setup a LaunchDaemon to up the limit with sudo privileges:

    sudo cp $GOPATH/src/github.com/pachyderm/pachyderm/etc/contributing/com.apple.launchd.limit.plist /Library/LaunchDaemons/

Once you restart, this will take effect. To see the limits, run:

    launchctl limit maxfiles

Before the change is in place you'll see something like `256    unlimited`. After the change you'll see a much bigger number in the first field. This ups the system wide limit, but you'll also need to set a per-process limit.

Second, up the per process limit by adding something like this to your `~/.bash_profile` :

    ulimit -n 12288

Unfortunately, even after setting that limit it never seems to report the updated version. So if you try

    ulimit

And just see `unlimited`, don't worry, it took effect.

To make sure all of these settings are working, you can test that you have the proper setup by running:

    make test-pfs-server

If this fails with a timeout, you'll probably also see 'too many files' type of errors. If that test passes, you're all good!

### Timeout helper

You'll need the `timeout` utility to run the `make launch` task. To install on mac, do:

    brew install coreutils

And then make sure to prepend the following to your path:

    PATH="/usr/local/opt/coreutils/libexec/gnubin:$PATH"

## Dev cluster

Now launch the dev cluster: `make launch-dev-vm`.

And check it's status: `kubectl get all`

## pachctl

This will install the dev version of `pachctl`:

    cd $GOPATH/src/github.com/pachyderm/pachyderm
    make install
    pachctl version

And make sure that `$GOPATH/bin` is on your `$PATH` somewhere
