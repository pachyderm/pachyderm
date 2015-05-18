FROM golang:1.4

ENV GOPATH /go
ENV PFS github.com/pachyderm/pfs

RUN apt-get update && apt-get install -y golang git mercurial && rm -rf /var/lib/apt/lists/*
RUN go get github.com/coreos/go-etcd/etcd
RUN go get code.google.com/p/go-uuid/uuid
RUN go get github.com/samalba/dockerclient
RUN go get github.com/bitly/go-simplejson
RUN go get github.com/mitchellh/goamz/...
RUN go get github.com/go-fsnotify/fsnotify
ADD . /go/src/$PFS
RUN go install -race $PFS/...
RUN ln $GOPATH/src/$PFS/scripts/pfs-test /usr/local/bin/pfs-test
RUN ln $GOPATH/src/$PFS/scripts/pfs-bench /usr/local/bin/pfs-bench
RUN ln $GOPATH/src/$PFS/scripts/btrfs-wrapper /bin/btrfs

EXPOSE 80
