# When this fails the first thing you should try is:
# NOCACHE=1 make update-test-deps update-deps-list docker-build-test
FROM ubuntu:14.04
MAINTAINER peter@pachyderm.io

RUN \
  apt-get update -yq && \
  apt-get install -yq --no-install-recommends \
    btrfs-tools \
    build-essential \
    ca-certificates \
    cmake \
    curl \
    fuse \
    git \
    libssl-dev \
    pkg-config \
    mercurial
RUN \
  curl -sSL https://storage.googleapis.com/golang/go1.5beta3.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
RUN go get golang.org/x/tools/cmd/vet github.com/kisielk/errcheck github.com/golang/lint/golint
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/git2go
WORKDIR /go/src/github.com/pachyderm/pachyderm
ADD etc/git2go/install.sh /go/src/github.com/pachyderm/pachyderm/etc/git2go/
RUN sh -x etc/git2go/install.sh
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/deps
ADD etc/deps/deps.list /go/src/github.com/pachyderm/pachyderm/etc/deps/
RUN cat etc/deps/deps.list | xargs go get -insecure
RUN \
  cd /go/src/github.com/imdario/mergo && \
  git checkout 6633656539c1639d9d78127b7d47c622b5d7b6dc
ADD . /go/src/github.com/pachyderm/pachyderm/
