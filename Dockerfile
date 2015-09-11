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
    mercurial && \
  apt-get clean && \
  rm -rf /var/lib/apt
RUN \
  curl -sSL https://storage.googleapis.com/golang/go1.5.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
ENV GO15VENDOREXPERIMENT 1
RUN go get golang.org/x/tools/cmd/vet github.com/kisielk/errcheck github.com/golang/lint/golint
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm/etc/git2go
WORKDIR /go/src/github.com/pachyderm/pachyderm
RUN \
  curl -sSL https://get.docker.com/builds/Linux/x86_64/docker-1.8.1 > /bin/docker && \
  chmod +x /bin/docker
RUN \
  curl -sSL https://github.com/docker/compose/releases/download/1.4.0/docker-compose-Linux-x86_64 > /bin/docker-compose && \
  chmod +x /bin/docker-compose
ADD etc/git2go/install.sh /go/src/github.com/pachyderm/pachyderm/etc/git2go/
RUN sh -x etc/git2go/install.sh
ADD . /go/src/github.com/pachyderm/pachyderm/
