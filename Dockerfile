FROM golang:1.4.2

RUN mkdir -p /go/src/github.com/pachyderm/pachyderm
WORKDIR /go/src/github.com/pachyderm/pachyderm
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/bin
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/deps
RUN \
	go get -v golang.org/x/tools/cmd/vet && \
	go get -v github.com/kisielk/errcheck && \
	go get -v github.com/golang/lint/golint && \
	go get -v golang.org/x/tools/cmd/vet && \
	go get -v github.com/stretchr/testify
ADD etc/deps/deps.list /go/src/github.com/pachyderm/pachyderm/etc/deps/
RUN cat etc/deps/deps.list | xargs go get
RUN \
  go get github.com/coreos/go-etcd/etcd && \
  cd /go/src/github.com/coreos/go-etcd && \
  git checkout release-0.4
ADD etc/bin /go/src/github.com/pachyderm/pachyderm/etc/bin/
RUN \
  ln /go/src/github.com/pachyderm/pachyderm/etc/bin/btrfs-wrapper /bin/btrfs && \
  ln /go/src/github.com/pachyderm/pachyderm/etc/bin/fleetctl-wrapper /bin/fleetctl
ADD . /go/src/github.com/pachyderm/pachyderm/
