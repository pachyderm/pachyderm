FROM ubuntu:14.04
LABEL maintainer="jdoliner@pachyderm.io"

RUN \
  apt-get update -yq && \
  apt-get install ca-certificates aria2 -yq --no-install-recommends && \
  aria2c -Z https://raw.githubusercontent.com/ilikenwf/apt-fast/master/apt-fast https://raw.githubusercontent.com/ilikenwf/apt-fast/master/apt-fast.conf && \
  mv ./apt-fast /bin/apt-fast && \
  chmod +x /bin/apt-fast && \
  mv ./apt-fast.conf /etc/apt-fast.conf && \
  /bin/apt-fast install -y -q --no-install-recommends \
    build-essential \
    curl \
    ca-certificates \
    cmake \
    fuse \
    git \
    libssl-dev \
    mercurial \
    pkg-config && \
  apt-get clean && \
  rm -rf /var/lib/apt
RUN \
  curl -fsSL https://get.docker.com/builds/Linux/x86_64/docker-1.12.1.tgz | tar -C /bin -xz docker/docker --strip-components=1 && \
  chmod +x /bin/docker
RUN \
  curl -sSL https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz | tar -C /usr/local -xz && \
  mkdir -p /go/bin
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
ENV GO15VENDOREXPERIMENT 1
RUN go get github.com/kisielk/errcheck github.com/golang/lint/golint
RUN mkdir -p /go/src/github.com/pachyderm/pachyderm
ADD . /go/src/github.com/pachyderm/pachyderm/
WORKDIR /go/src/github.com/pachyderm/pachyderm
