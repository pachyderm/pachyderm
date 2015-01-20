FROM ubuntu:14.10

ENV GOPATH /go
ENV PFS github.com/pachyderm/pfs

RUN apt-get update && apt-get install -y golang btrfs-tools git mercurial && rm -rf /var/lib/apt/lists/*
RUN go get github.com/coreos/go-etcd/etcd
RUN go get code.google.com/p/go-uuid/uuid
RUN go get github.com/samalba/dockerclient
ADD . /go/src/$PFS
RUN go install $PFS/services/master && go install $PFS/services/router && go install $PFS/deploy
RUN ln $GOPATH/src/$PFS/scripts/pfs-test /usr/local/bin/pfs-test
RUN ln $GOPATH/src/$PFS/scripts/pfs-bench /usr/local/bin/pfs-bench

EXPOSE 80
