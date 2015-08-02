# When this fails the first thing you should try is:
# NOCACHE=1 make update-test-deps update-deps-list docker-build
FROM ubuntu:15.04

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools \
    build-essential \
    curl \
    libgit2-dev \
    pkg-config \
    git \
    fuse
RUN \
  curl -sSL https://storage.googleapis.com/golang/go1.5beta3.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
RUN go get golang.org/x/tools/cmd/vet github.com/kisielk/errcheck github.com/golang/lint/golint
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm
WORKDIR /go/src/github.com/pachyderm/pachyderm
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/deps
ADD etc/deps/deps.list /go/src/github.com/pachyderm/pachyderm/etc/deps/
RUN cat etc/deps/deps.list | xargs go get -insecure
ADD . /go/src/github.com/pachyderm/pachyderm/
