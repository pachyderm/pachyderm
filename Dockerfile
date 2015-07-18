FROM golang:1.4.2

RUN mkdir -p /go/src/github.com/pachyderm/pachyderm
WORKDIR /go/src/github.com/pachyderm/pachyderm
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/bin
RUN \
	go get -v golang.org/x/tools/cmd/vet && \
	go get -v github.com/kisielk/errcheck && \
	go get -v github.com/golang/lint/golint && \
	go get -v golang.org/x/tools/cmd/vet && \
	go get -v github.com/stretchr/testify
RUN \
  go get github.com/coreos/go-etcd/etcd && \
  cd /go/src/github.com/coreos/go-etcd && \
  git checkout release-0.4 && \
  go get github.com/aws/aws-sdk-go/aws && \
  go get github.com/aws/aws-sdk-go/service/s3 && \
  go get github.com/fsouza/go-dockerclient && \
  go get github.com/go-fsnotify/fsnotify && \
  go get github.com/peter-edge/go-env && \
  go get github.com/peter-edge/go-google-protobuf && \
  go get github.com/satori/go.uuid && \
  go get github.com/spf13/cobra && \
  go get google.golang.org/grpc
ADD etc/bin /go/src/github.com/pachyderm/pachyderm/etc/bin/
RUN ln /go/src/github.com/pachyderm/pachyderm/etc/bin/btrfs-wrapper /bin/btrfs
RUN ln /go/src/github.com/pachyderm/pachyderm/etc/bin/fleetctl-wrapper /bin/fleetctl
ADD . /go/src/github.com/pachyderm/pachyderm/
