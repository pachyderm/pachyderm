FROM ubuntu:14.10

ENV GOPATH /go
ENV PFS github.com/pachyderm-io/pfs

RUN apt-get update && apt-get install -y golang btrfs-tools git && rm -rf /var/lib/apt/lists/*
RUN go get github.com/coreos/go-etcd/etcd
ADD . /go/src/$PFS
RUN go install $PFS/services/master && go install $PFS/services/replica && go install $PFS/services/router && go install $PFS/deploy

EXPOSE 80
