# Setup for contributors

## General requirements

First, go through the general [Local Installation Instructions](https://docs.pachyderm.com/latest/getting_started/local_installation/). Additionally, make sure you have the following installed:

- golang 1.12+
- docker
- [jq](https://stedolan.github.io/jq/)
- [pv](http://ivarch.com/programs/pv.shtml)

## Bash helpers

To stay up to date, we recommend doing the following.

First clone the code:
(Note, as of 07/11/19 pachyderm is using go modules and recommends cloning the code outside of the $GOPATH, we use the location ~/workspace as an example, but the code can live anywhere)
```shell
    cd ~/workspace
    git clone git@github.com:pachyderm/pachyderm
```
Then update your `~/.bash_profile` by adding the line:
```shell
    source ~/workspace/pachyderm/etc/contributing/bash_helpers
```
And you'll stay up to date!

## Special macOS configuration

### File descriptor limit

If you're running tests locally, you'll need to up your file descriptor limit. To do this, first setup a LaunchDaemon to up the limit with sudo privileges:
```shell
    sudo cp ~/workspace/pachyderm/etc/contributing/com.apple.launchd.limit.plist /Library/LaunchDaemons/
```
Once you restart, this will take effect. To see the limits, run:
```shell
    launchctl limit maxfiles
```
Before the change is in place you'll see something like `256    unlimited`. After the change you'll see a much bigger number in the first field. This ups the system wide limit, but you'll also need to set a per-process limit.

Second, up the per process limit by adding something like this to your `~/.bash_profile` :
```shell
    ulimit -n 12288
```
Unfortunately, even after setting that limit it never seems to report the updated version. So if you try
```shell
    ulimit
```
And just see `unlimited`, don't worry, it took effect.

To make sure all of these settings are working, you can test that you have the proper setup by running:
```shell
    make test-pfs-server
```
If this fails with a timeout, you'll probably also see 'too many files' type of errors. If that test passes, you're all good!

### Timeout helper

You'll need the `timeout` utility to run the `make launch` task. To install on mac, do:
```shell
    brew install coreutils
```
And then make sure to prepend the following to your path:
```shell
    PATH="/usr/local/opt/coreutils/libexec/gnubin:$PATH"
```
## Dev cluster

Now launch the dev cluster: `make launch-dev-vm`.

And check it's status: `kubectl get all`.

## pachctl

This will install the dev version of `pachctl`:

```shell
    cd ~/workspace/pachyderm
    make install
    pachctl version
```

And make sure that `$GOPATH/bin` is on your `$PATH` somewhere

## Fully resetting

Instead of running the makefile targets to re-compile `pachctl` and redeploy
a dev cluster, we have a script that you can use to fully reset your pachyderm
environment:

1. All existing cluster data is deleted
1. If possible, the virtual machine that the cluster is running on is wiped
out
1. `pachctl` is recompiled
1. The dev cluster is re-deployed

This reset is a bit more time consuming than running one-off Makefile targets,
but comprehensively ensures that the cluster is in its expected state, and is
especially helpful when you're first getting started with contributions and
don't yet have a complete intuition on the various ways a cluster may get in
an unexpected state. It's been tested on docker for mac and minikube, but
likely works in other kubernetes environments as well.

To run it, simply call `./etc/reset.py` from the pachyderm repo root.
