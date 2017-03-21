# Setup For Contributors

If you're contributing to pachyderm, you may need the following additional setup.

1. General Requirements
  - golang version 1.7
  - docker
  - FUSE
2. bash helpers
3. protoc
4. gcloud CLI
5. kubectl
6. pachctl
7. Special Mac OS X Configuration

---

## General Requirements

Install:

- golang 1.7
- docker
- FUSE

## Bash helpers

To stay up to date, we recommend doing the following.

First clone the code:

    cd $GOPATH/src
    mkdir -p github.com/pachyderm
    cd github.com/pachyderm
    git clone git@github.com:pachyderm/pachyderm

Then update your `~/.bash_profile` by adding the line:

    source $GOPATH/src/github.com/pachyderm/pachyderm/contributing/bash_helpers

And you'll stay up to date!


## Protoc

[Download Page](https://github.com/google/protobuf/releases)

Mac OS X Note: brew install for mac didn't provide the right version - namely proto3

- I used [this version](https://github.com/google/protobuf/releases/download/v3.0.0-beta-2/protoc-3.0.0-beta-2-osx-x86_64.zip)
- then move the binary to your path (/usr/local/bin)
- I have the following version installed:

    $ protoc --version
    libprotoc 3.0.0

Then installed protoc-gen-go by doing:

    go get -u -v github.com/golang/protobuf/proto
    go get -u -v github.com/golang/protobuf/protoc-gen-go
    go get -u -v github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway

## gcloud CLI

Internally we develop using containerized kubernetes using GKE. The following instructions outline how to set this up.

[Download Page](https://cloud.google.com/sdk/)

Setup Google Cloud Platform via the web

- login with your Gmail or G Suite account
  - click the silhouette in the upper right to make sure you're logged in with the right account
- get your owner/admin to setup a project for you (e.g. YOURNAME-dev)
- then they need to go into the project > settings > permissions and add you
  - hint to owner/admin: its the permissions button in one of the left hand popin menus (GKE UI can be confusing)
- you should have an email invite to accept
- click 'use google APIS' (or something along the lines of enable/manage APIs)
- click through to google compute engine API and enable it or click the 'get started' button to make it provision

Then, locally, run the following commands one at a time:

    gcloud auth login
    gcloud init

    # This should have you logged in / w gcloud
    # The following will only work after your GKE owner/admin adds you to the right project on gcloud:

    gcloud config set project YOURNAME-dev
    gcloud compute instances list

    # Now create instance using our bash helper
    create_docker_machine

    # And attach to the right docker daemon
    eval "$(docker-machine env dev)"

Setup a project on gcloud

- go to console.cloud.google.com/start
- make sure you're logged in w your gmail account
- create project 'YOURNAME-dev'

## kubectl

Again, [the bash_profile helpers are here](https://github.com/pachyderm/pachyderm/blob/master/contributing/bash_helpers)

Now that you have gcloud, just do:

    gcloud components update kubectl
    # Now you need to start port forwarding to allow kubectl client talk to the kubernetes service on GCE

    portfowarding
    # To see this alias, look at the bash_helpers

    kubectl version
    # should report a client version, not a server version yet

    make launch-kube
    # to deploy kubernetes service

    kubectl version
    # now you should see a client and server version

    docker ps
    # you should see a few processes
    
    kubectl get all
    # and you should see the kubernetes controller running

## Pachyderm cluster deployment

    make launch

And now you should see rethink/pachd/etcd pods running when you do:

    kubectl get all    


## pachctl

To install this tool, you'll need the code!

    mkdir -p $GOPATH/src/github.com/pachyderm
    cd $GOPATH/src/github.com/pachyderm
    git clone git@github.com:pachyderm/pachyderm
    cd pachyderm
    make install
    pachctl version

And make sure that `$GOPATH/bin` is on your `$PATH` somewhere


## Special Mac OS X Configuration

### File Descriptor Limit

If you're running tests locally, you'll need to up your file descriptor limit. To do this, first setup a LaunchDaemon to up the limit with sudo privileges:

    sudo cp $GOPATH/src/github.com/pachyderm/pachyderm/contributing/com.apple.launchd.limit.plist /Library/LaunchDaemons/

Once you restart, this will take effect. To see the limits, run:

    launchctl limit maxfiles

Before the change is in place you'll see something like `256    unlimited`. After the change you'll see a much bigger number in the first field. This ups the system wide limit, but you'll also need to set a per-process limit. 

Second, up the per process limit by adding something like this to your `~/.bash_profile` :

    ulimit -n 12288

Unfortunately, even after setting that limit it never seems to report the updated version. So if you try

    ulimit

And just see `unlimited`, don't worry, it took effect.

To make sure all of these settings are working, you can test that you have the proper setup by running:

    CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 go test -timeout 10s -p 1 -parallel 1 -short ./src/server/pfs/server

If this fails w a timeout, you'll probably also see 'too many files' type of errors. If that test passes, you're all good!

### Timeout Helper

You'll need the `timeout` utility to run the `make launch` task. To install on mac, do:

    brew install coreutils

And then make sure to prepend the following to your path:

    PATH="/usr/local/opt/coreutils/libexec/gnubin:$PATH"

