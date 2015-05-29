FROM ubuntu:15.04

ENV GOPATH /go
ENV PFS github.com/pachyderm/pfs

RUN apt-get update && apt-get install -y golang git mercurial && rm -rf /var/lib/apt/lists/*
RUN go get github.com/coreos/go-etcd/etcd && cd $GOPATH/src/github.com/coreos/go-etcd && git checkout release-0.4
RUN go get code.google.com/p/go-uuid/uuid
RUN go get github.com/fsouza/go-dockerclient
RUN go get github.com/bitly/go-simplejson
RUN go get github.com/mitchellh/goamz/...
RUN go get github.com/go-fsnotify/fsnotify
ADD . /go/src/$PFS
RUN ln -s /go/src/$PFS/deploy/templates templates
RUN go install -race $PFS/services/shard && go install $PFS/services/router && go install $PFS/deploy
RUN ln $GOPATH/src/$PFS/scripts/btrfs-wrapper /bin/btrfs
RUN ln $GOPATH/src/$PFS/scripts/fleetctl-wrapper /bin/fleetctl

EXPOSE 80
