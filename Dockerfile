FROM golang:1.4.2

RUN go get github.com/coreos/go-etcd/etcd && cd /go/src/github.com/coreos/go-etcd && git checkout release-0.4
RUN go get github.com/satori/go.uuid
RUN go get github.com/fsouza/go-dockerclient
RUN go get github.com/mitchellh/goamz/...
RUN go get github.com/go-fsnotify/fsnotify
ADD . /go/src/github.com/pachyderm/pfs
RUN ln -s /go/src/github.com/pachyderm/pfs/deploy/templates /templates
RUN go install github.com/pachyderm/pfs/...
RUN ln /go/src/github.com/pachyderm/pfs/etc/bin/btrfs-wrapper /bin/btrfs
RUN ln /go/src/github.com/pachyderm/pfs/etc/bin/fleetctl-wrapper /bin/fleetctl

EXPOSE 80
WORKDIR /go/src/github.com/pachyderm/pfs
