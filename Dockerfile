FROM ubuntu

ENV GOPATH /go
RUN apt-get update && apt-get install -y golang btrfs-tools git && rm -rf /var/lib/apt/lists/*
RUN go get github.com/coreos/go-etcd/etcd
ADD . /go/src/pfs
RUN go install pfs/services/master && go install pfs/services/replica && go install pfs/services/router && go install pfs/deploy

EXPOSE 80
